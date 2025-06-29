package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"os"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"gopkg.in/yaml.v3"
)

// MCPServer represents the core MCP server implementation
type MCPServer struct {
	server *server.MCPServer
	config *Config
}

// Config holds MCP server configuration
type Config struct {
	Host           string        `yaml:"host"`
	Port           int           `yaml:"port"`
	MaxConnections int           `yaml:"max_connections"`
	Timeout        time.Duration `yaml:"timeout"`
	LLMConfig      LLMConfig     `yaml:"llm_config"`
	Modules        []Module      `yaml:"modules"`
	IPFS           IPFSConfig    `yaml:"ipfs"`
}

type IPFSConfig struct {
	Enable    bool      `yaml:"enable"`
	LassieNet LassieNet `yaml:"lassie_net"`
	CIDS      []string  `yaml:"cids"`
}

type LassieNet struct {
	Scheme string `yaml:"scheme"`
	Host   string `yaml:"host"`
	Port   int    `yaml:"port"`
}

// LLMConfig contains LLM provider configurations
type LLMConfig struct {
	BaseURL  string       `yaml:"base_url"`
	Provider string       `yaml:"provider"`
	OpenAI   OpenAIConfig `yaml:"openai"`
}

// OpenAIConfig contains OpenAI-specific settings
type OpenAIConfig struct {
	APIKey      string  `yaml:"api_key"`
	Model       string  `yaml:"model"`
	Temperature float64 `yaml:"temperature"`
	MaxTokens   int     `yaml:"max_tokens"`
}

// Module defines a WASM module and its exposed tools
type Module struct {
	Name     string `yaml:"name"`
	WASMPath string `yaml:"wasm_path"`
	Tools    []Tool `yaml:"tools"`
}

// Tool defines an MCP tool interface
type Tool struct {
	Name        string      `yaml:"name"`
	Description string      `yaml:"description"`
	Inputs      []ToolInput `yaml:"inputs"`
	Outputs     ToolOutput  `yaml:"outputs"`
}

// ToolInput defines tool input parameters
type ToolInput struct {
	Name        string `yaml:"name"`
	Type        string `yaml:"type"`
	Required    bool   `yaml:"required"`
	Description string `yaml:"description"`
}

// ToolOutput defines tool output structure
type ToolOutput struct {
	Type        string `yaml:"type"`
	Description string `yaml:"description"`
}

// NewServer creates a new MCP server instance from config file
func NewServer(configPath string) (*MCPServer, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	return NewMCPServer(&config), nil
}

// NewMCPServer creates a new MCP server instance from config
func NewMCPServer(config *Config) *MCPServer {
	log.Println("Initializing new MCP server instance")
	hooks := &server.Hooks{}
	wasmEngine := NewWASMEngine(config)

	// Setup default hooks
	hooks.AddBeforeAny(func(ctx context.Context, id any, method mcp.MCPMethod, message any) {
		log.Printf("BeforeAny: %s, %v, %v\n", method, id, message)
	})
	hooks.AddOnSuccess(func(ctx context.Context, id any, method mcp.MCPMethod, message any, result any) {
		log.Printf("OnSuccess: %s, %v, %v, %v\n", method, id, message, result)
	})
	hooks.AddOnError(func(ctx context.Context, id any, method mcp.MCPMethod, message any, err error) {
		log.Printf("OnError: %s, %v, %v, %v\n", method, id, message, err)
	})

	log.Println("Creating MCP server with capabilities")
	mcpServer := server.NewMCPServer(
		"dANP-MCP",
		"1.0.0",
		server.WithResourceCapabilities(true, true),
		server.WithPromptCapabilities(true),
		server.WithToolCapabilities(true),
		server.WithLogging(),
		server.WithHooks(hooks),
	)

	// Register WASM module tools from config
	log.Printf("Registering %d WASM modules from config", len(config.Modules))
	for _, module := range config.Modules {
		log.Printf("Processing module: %s (WASM path: %s)", module.Name, module.WASMPath)
		log.Printf("Loading module with %d tools", len(module.Tools))

		if err := wasmEngine.LoadModule(context.Background(), module.WASMPath); err != nil {
			log.Printf("Failed to load WASM module %s: %v", module.WASMPath, err)
			continue
		}

		log.Printf("Registering tools for module: %s", module.Name)
		if err := wasmEngine.RegisterWASMTools(mcpServer, module.WASMPath, module.Tools); err != nil {
			log.Printf("Failed to register tools from WASM module %s: %v", module.WASMPath, err)
		} else {
			log.Printf("Successfully registered %d tools for module: %s", len(module.Tools), module.Name)
		}
	}

	return &MCPServer{
		server: mcpServer,
		config: config,
	}
}

// Start begins the MCP server
func (s *MCPServer) Start() error {
	log.Println("Starting MCP server")
	if s.config.Host == "" {
		s.config.Host = "0.0.0.0"
		log.Println("Using default host: 0.0.0.0")
	}
	if s.config.Port == 0 {
		s.config.Port = 18080
		log.Println("Using default port: 18080")
	}

	// Create HTTP server with MCP server
	httpServer := server.NewStreamableHTTPServer(s.server)

	// Create a custom HTTP server with additional routes
	mux := http.NewServeMux()

	// Register the MCP server handler for the root path
	mux.Handle("/", httpServer)

	// Add a handler for /tools endpoint to list available tools
	mux.HandleFunc("/tools", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get all registered tools
		toolNames := []string{}

		// Manually track tools from our config
		for _, module := range s.config.Modules {
			for _, tool := range module.Tools {
				toolNames = append(toolNames, tool.Name)
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"tools": toolNames,
		})
	})

	// Add handlers for individual tool endpoints
	mux.HandleFunc("/tools/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) < 3 {
			http.Error(w, "Invalid tool path", http.StatusBadRequest)
			return
		}

		toolName := pathParts[2]

		// Find the tool in our config
		var foundTool *Tool
		var foundModule *Module

		for i, module := range s.config.Modules {
			for j, tool := range module.Tools {
				if tool.Name == toolName {
					foundTool = &s.config.Modules[i].Tools[j]
					foundModule = &s.config.Modules[i]
					break
				}
			}
			if foundTool != nil {
				break
			}
		}

		if foundTool == nil {
			http.Error(w, "Tool not found", http.StatusNotFound)
			return
		}

		// Read raw request body
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusBadRequest)
			return
		}

		// Get the WASM plugin
		wasmEngine := NewWASMEngine(s.config)
		if err := wasmEngine.LoadModule(r.Context(), foundModule.WASMPath); err != nil {
			http.Error(w, fmt.Sprintf("Failed to load WASM module: %v", err), http.StatusInternalServerError)
			return
		}

		plugin, ok := wasmEngine.plugins[foundModule.WASMPath]
		if !ok {
			http.Error(w, "WASM module not loaded", http.StatusInternalServerError)
			return
		}

		// Call the WASM function with raw input
		input := bodyBytes

		log.Printf("Calling WASM function: %s with input: %s", toolName, string(input))
		_, output, err := plugin.Plugin.Call(toolName, input)
		if err != nil {
			http.Error(w, fmt.Sprintf("WASM call failed: %v", err), http.StatusInternalServerError)
			return
		}

		// Return the result
		w.Header().Set("Content-Type", "application/json")
		w.Write(output)
	})

	// Start the HTTP server
	addr := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)
	log.Printf("MCP server listening on %s", addr)
	log.Printf("Server configuration: %+v", s.config)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return server.ListenAndServe()
}

// Stop gracefully shuts down the MCP server
func (s *MCPServer) Stop(ctx context.Context) error {
	log.Println("Initiating MCP server shutdown")

	// Close WASM resources
	wasmEngine := NewWASMEngine(s.config)
	if err := wasmEngine.Close(ctx); err != nil {
		log.Printf("Error closing WASM resources: %v", err)
	}

	// TODO: Add additional cleanup if needed

	log.Println("MCP server shutdown complete")
	return nil
}
