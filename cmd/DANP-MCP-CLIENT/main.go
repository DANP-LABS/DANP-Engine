package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sashabaranov/go-openai"
	"github.com/DANP-LABS/DANP-Engine/pkg/mcpclient"
)

func main() {
	// Define command line flags
	stdioCmd := flag.String("stdio", "", "Command to execute for stdio transport (e.g. 'python server.py')")
	httpURL := flag.String("http", "", "URL for HTTP transport (e.g. 'http://localhost:18080/')")
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: No .env file found - using environment variables only")
	}

	deepseekKey := flag.String("deepseek-key", os.Getenv("DEEPSEEK_KEY"), "DeepSeek API key (required for LLM access)")
	deepseekModel := flag.String("deepseek-model", "deepseek-chat", "DeepSeek model to use (deepseek-chat or deepseek-reasoner)")
	flag.Parse()

	// Validate DeepSeek key
	if *deepseekKey == "" {
		log.Fatal("DeepSeek API key is required. Set it via --deepseek-key flag or DEEPSEEK_KEY environment variable")
	}

	// Validate flags
	if (*stdioCmd == "" && *httpURL == "") || (*stdioCmd != "" && *httpURL != "") {
		fmt.Println("Error: You must specify exactly one of --stdio or --http")
		flag.Usage()
		os.Exit(1)
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create client
	client, err := mcpclient.NewClient(ctx, *stdioCmd, *httpURL)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Set up notification handler
	client.OnNotification(func(notification mcp.JSONRPCNotification) {
		fmt.Printf("Received notification: %s\n", notification.Method)
	})

	// Initialize the client
	fmt.Println("Initializing client...")
	serverInfo, err := client.Initialize(ctx)
	if err != nil {
		log.Fatalf("Failed to initialize: %v", err)
	}

	// Display server information
	fmt.Printf("Connected to server: %s (version %s)\n",
		serverInfo.ServerInfo.Name,
		serverInfo.ServerInfo.Version)
	fmt.Printf("Server capabilities: %+v\n", serverInfo.Capabilities)

	// Register and manage tools if the server supports them
	if serverInfo.Capabilities.Tools != nil {
		fmt.Println("Registering and managing tools...")
		tm := mcpclient.NewToolManager(client)
		tools, err := tm.ListTools(ctx)
		if err != nil {
			log.Printf("Failed to list tools: %v", err)
		} else {
			// Add DeepSeek LLM as a virtual tool if API key is provided
			if *deepseekKey != "" {
				tools = append(tools, mcp.Tool{
					Name:        "deepseek-llm",
					Description: "DeepSeek large language model with MCP tool integration",
				})
			}

			fmt.Printf("Available tools: %d\n", len(tools))
			for _, tool := range tools {
				fmt.Printf("  - %s: %s\n", tool.Name, tool.Description)
			}

			// AI-driven interaction loop
			if len(tools) > 0 {
				fmt.Println("\nAI Tool Manager is ready. Type 'exit' to quit.")
				dsm := mcpclient.NewDeepSeekManager(*deepseekKey, *deepseekModel, tm)

				// Create a longer-lived context for the conversation
				convCtx, convCancel := context.WithCancel(context.Background())
				defer convCancel()

				// Store conversation history
				var conversationHistory []openai.ChatCompletionMessage
				conversationHistory = append(conversationHistory, openai.ChatCompletionMessage{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You are an AI assistant that can intelligently call MCP tools. Analyze the user's request and determine if any tools should be used. When using tools, provide clear explanations of your decisions. You can chain multiple tool calls if needed to complete complex tasks. Maintain conversation context across multiple interactions.",
				})

				// Interactive conversation loop
				for {
					// Read multiline input until empty line
					fmt.Print("\nEnter your request (empty line to submit, 'exit' to quit):\n> ")
					var promptLines []string
					scanner := bufio.NewScanner(os.Stdin)
					for scanner.Scan() {
						line := scanner.Text()
						if line == "" {
							break
						}
						promptLines = append(promptLines, line)
						fmt.Print("> ")
					}

					prompt := strings.Join(promptLines, "\n")
					if prompt == "" {
						continue
					}

					if strings.ToLower(prompt) == "exit" {
						break
					}

					// Show loading animation
					fmt.Print("\n\x1b[36m") // Cyan color
					for i := 0; i < 3; i++ {
						fmt.Print(".")
						time.Sleep(400 * time.Millisecond)
					}
					fmt.Print("\x1b[0m\n") // Reset color with newline

					// Add user message to history
					conversationHistory = append(conversationHistory, openai.ChatCompletionMessage{
						Role:    openai.ChatMessageRoleUser,
						Content: prompt,
					})

					// Call DeepSeek API
					response, newHistory, err := dsm.CallWithHistory(convCtx, conversationHistory, tools)
					if err != nil {
						log.Printf("AI API error: %v", err)
						continue
					}

					conversationHistory = newHistory
					
					// Show loading animation
					fmt.Print("\n\x1b[36m") // Cyan color
					for i := 0; i < 3; i++ {
						fmt.Print(".")
						time.Sleep(200 * time.Millisecond)
					}
					fmt.Print("\x1b[0m") // Reset color

					// Print response
					fmt.Printf("\n\x1b[32mAI Response:\x1b[0m\n\x1b[33m%s\x1b[0m\n", response)
				}
			}
		}
	}

	// List available resources if the server supports them
	if serverInfo.Capabilities.Resources != nil {
		fmt.Println("Fetching available resources...")
		if err := mcpclient.ListResources(ctx, client); err != nil {
			log.Printf("Failed to list resources: %v", err)
		}
	}

	fmt.Println("Client initialized successfully. Shutting down...")
}

func parseCommand(cmd string) []string {
	// This is a simple implementation that doesn't handle quotes or escapes
	// For a more robust solution, consider using a shell parser library
	var result []string
	var current string
	var inQuote bool
	var quoteChar rune

	for _, r := range cmd {
		switch {
		case r == ' ' && !inQuote:
			if current != "" {
				result = append(result, current)
				current = ""
			}
		case (r == '"' || r == '\''):
			if inQuote && r == quoteChar {
				inQuote = false
				quoteChar = 0
			} else if !inQuote {
				inQuote = true
				quoteChar = r
			} else {
				current += string(r)
			}
		default:
			current += string(r)
		}
	}

	if current != "" {
		result = append(result, current)
	}

	return result
}
