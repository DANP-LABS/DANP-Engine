package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/IceFireLabs/DANP-Engine/pkg/ipfs"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	extism "github.com/extism/go-sdk"
	"github.com/tetratelabs/wazero"
)

// WASMEngine manages WASM module execution
type WASMEngine struct {
	plugins map[string]*WASMPlugin
	mu      sync.Mutex
	config  *Config
}

// WASMPlugin represents a loaded WASM plugin
type WASMPlugin struct {
	Plugin *extism.Plugin
	Mutex  sync.Mutex
}

// NewWASMEngine creates a new WASM execution environment
func NewWASMEngine(config *Config) *WASMEngine {
	return &WASMEngine{
		plugins: make(map[string]*WASMPlugin),
		config:  config,
	}
}

// LoadModule loads a WASM module from file or IPFS
func (w *WASMEngine) LoadModule(ctx context.Context, path string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	log.Printf("Loading WASM module from: %s", path)

	manifest := extism.Manifest{Wasm: []extism.Wasm{}}

	// Handle protocol prefixes
	if strings.HasPrefix(path, "file://") {
		// File protocol - strip prefix and load from filesystem
		filePath := strings.TrimPrefix(path, "file://")
		if _, err := os.Stat(filePath); err == nil {
			log.Printf("Loading WASM module from filesystem: %s", filePath)
			manifest.Wasm = append(manifest.Wasm, extism.WasmFile{Path: filePath})
		} else {
			return fmt.Errorf("WASM module not found: %s", filePath)
		}
	} else if strings.HasPrefix(path, "IPFS://") {
		// IPFS protocol - requires IPFS to be enabled
		if !w.config.IPFS.Enable {
			return fmt.Errorf("IPFS support is not enabled")
		}
		cid := strings.TrimPrefix(path, "IPFS://")
		log.Printf("Loading WASM module from IPFS CID: %s", cid)
		
		ipfsC := ipfs.NewIPFSClient(w.config.IPFS.LassieNet.Scheme,
			w.config.IPFS.LassieNet.Host,
			w.config.IPFS.LassieNet.Port)

		data, err := ipfs.GetDATAFromIPFSCID(ipfsC, cid)
		if err != nil {
			return fmt.Errorf("failed to load WASM from IPFS: %w", err)
		}

		for _, wasmData := range data {
			manifest.Wasm = append(manifest.Wasm, extism.WasmData{Data: wasmData})
		}
	} else {
		// No protocol - try direct path (backward compatibility)
		if _, err := os.Stat(path); err == nil {
			log.Printf("Loading WASM module from filesystem: %s", path)
			manifest.Wasm = append(manifest.Wasm, extism.WasmFile{Path: path})
		} else if w.config.IPFS.Enable {
			log.Printf("Loading WASM module from IPFS (direct CID): %s", path)
			ipfsC := ipfs.NewIPFSClient(w.config.IPFS.LassieNet.Scheme,
				w.config.IPFS.LassieNet.Host,
				w.config.IPFS.LassieNet.Port)

			data, err := ipfs.GetDATAFromIPFSCID(ipfsC, path)
			if err != nil {
				return fmt.Errorf("failed to load WASM from IPFS: %w", err)
			}

			for _, wasmData := range data {
				manifest.Wasm = append(manifest.Wasm, extism.WasmData{Data: wasmData})
			}
		} else {
			return fmt.Errorf("WASM module not found: %s", path)
		}
	}

	config := extism.PluginConfig{
		ModuleConfig: wazero.NewModuleConfig().WithSysWalltime(),
		EnableWasi:   true,
	}

	plugin, err := extism.NewPlugin(ctx, manifest, config, nil)
	if err != nil {
		log.Printf("Failed to create WASM plugin: %v", err)
		return fmt.Errorf("failed to create WASM plugin: %w", err)
	}

	w.plugins[path] = &WASMPlugin{
		Plugin: plugin,
	}

	log.Printf("Successfully loaded WASM module: %s", path)
	return nil
}

// RegisterWASMTools registers all tools from a WASM module
func (w *WASMEngine) RegisterWASMTools(s *server.MCPServer, modulePath string, tools []Tool) error {
	plugin, ok := w.plugins[modulePath]
	if !ok {
		log.Printf("WASM module not found during tool registration: %s", modulePath)
		return fmt.Errorf("WASM module not loaded: %s", modulePath)
	}

	log.Printf("Registering %d tools from WASM module: %s", len(tools), modulePath)

	// Check for available functions in the WASM module
	log.Printf("Checking available functions in WASM module: %s", modulePath)

	for _, tool := range tools {
		log.Printf("Registering tool: %s", tool.Name)
		
		// Check if the function exists in the WASM module
		functionExists := plugin.Plugin.FunctionExists(tool.Name)
		
		if !functionExists {
			log.Printf("Warning: Function %s not found in WASM module, skipping", tool.Name)
			continue
		}
		
		// Create tool with description and parameters
		toolOptions := []mcp.ToolOption{
			mcp.WithDescription(tool.Description),
		}
		
		// Add input parameters - we don't use schema in this version
		// Just log the parameters for debugging
		log.Printf("Tool %s has %d input parameters", tool.Name, len(tool.Inputs))
		for _, input := range tool.Inputs {
			log.Printf("  Parameter: %s (type: %s, required: %v)", 
				input.Name, input.Type, input.Required)
		}
		
		// Create the tool
		t := mcp.NewTool(tool.Name, toolOptions...)

		// Register the tool handler
		s.AddTool(t, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			plugin.Mutex.Lock()
			defer plugin.Mutex.Unlock()

			// Convert MCP request to WASM input
			input, err := json.Marshal(req.Params)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal input: %w", err)
			}

			// Call WASM function
			log.Printf("Calling WASM function: %s with input: %s", tool.Name, string(input))
			_, output, err := plugin.Plugin.Call(tool.Name, input)
			if err != nil {
				log.Printf("WASM call failed for %s: %v", tool.Name, err)
				return nil, fmt.Errorf("WASM call failed: %w", err)
			}
			log.Printf("WASM function %s executed successfully with output: %s", tool.Name, string(output))

			// Create MCP result
			result := &mcp.CallToolResult{
				Content: []mcp.Content{
					mcp.TextContent{
						Type: "text",
						Text: string(output),
					},
				},
			}

			return result, nil
		})
		
		log.Printf("Successfully registered tool: %s", tool.Name)
	}

	return nil
}

// Close cleans up WASM resources
func (w *WASMEngine) Close(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	log.Printf("Cleaning up %d WASM plugins", len(w.plugins))

	for path, plugin := range w.plugins {
		log.Printf("Closing WASM plugin: %s", path)
		if err := plugin.Plugin.Close(ctx); err != nil {
			log.Printf("Failed to close WASM plugin %s: %v", path, err)
			return err
		}
		log.Printf("Successfully closed WASM plugin: %s", path)
	}
	log.Println("All WASM plugins closed")
	return nil
}
