package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sashabaranov/go-openai"
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
	registerTools := flag.Bool("register-tools", true, "Register MCP tools for use with LLM")
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

	// Create client based on transport type
	var c *client.Client

	if *stdioCmd != "" {
		fmt.Println("Initializing stdio client...")
		// Parse command and arguments
		args := parseCommand(*stdioCmd)
		if len(args) == 0 {
			fmt.Println("Error: Invalid stdio command")
			os.Exit(1)
		}

		// Create command and stdio transport
		command := args[0]
		cmdArgs := args[1:]

		// Create stdio transport with verbose logging
		stdioTransport := transport.NewStdio(command, nil, cmdArgs...)

		// Create client with the transport
		c = client.NewClient(stdioTransport)

		// Start the client
		if err := c.Start(ctx); err != nil {
			log.Fatalf("Failed to start client: %v", err)
		}

		// Set up logging for stderr if available
		if stderr, ok := client.GetStderr(c); ok {
			go func() {
				buf := make([]byte, 4096)
				for {
					n, err := stderr.Read(buf)
					if err != nil {
						if err != io.EOF {
							log.Printf("Error reading stderr: %v", err)
						}
						return
					}
					if n > 0 {
						fmt.Fprintf(os.Stderr, "[Server] %s", buf[:n])
					}
				}
			}()
		}
	} else {
		fmt.Println("Initializing HTTP client...")
		// Create HTTP transport
		httpTransport, err := transport.NewStreamableHTTP(*httpURL)
		// NOTE: the default streamableHTTP transport is not 100% identical to the stdio client.
		// By default, it could not receive global notifications (e.g. toolListChanged).
		// You need to enable the `WithContinuousListening()` option to establish a long-live connection,
		// and receive the notifications any time the server sends them.
		//
		//   httpTransport, err := transport.NewStreamableHTTP(*httpURL, transport.WithContinuousListening())
		if err != nil {
			log.Fatalf("Failed to create HTTP transport: %v", err)
		}

		// Create client with the transport
		c = client.NewClient(httpTransport)
	}

	// Set up notification handler
	c.OnNotification(func(notification mcp.JSONRPCNotification) {
		fmt.Printf("Received notification: %s\n", notification.Method)
	})

	// Initialize the client
	fmt.Println("Initializing client...")
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "MCP-Go Client",
		Version: "1.0.0",
	}
	initRequest.Params.Capabilities = mcp.ClientCapabilities{}

	serverInfo, err := c.Initialize(ctx, initRequest)
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
		toolsRequest := mcp.ListToolsRequest{}
		toolsResult, err := c.ListTools(ctx, toolsRequest)
		if err != nil {
			log.Printf("Failed to list tools: %v", err)
		} else {
			// Register MCP tools if enabled
			if *registerTools {
				fmt.Println("Registering MCP tools...")
				// Register built-in MCP tools
				// mcpTools := []mcp.Tool{
				// 	{
				// 		Name:        "mcp-file-read",
				// 		Description: "Read a file from the filesystem",
				// 	},
				// 	{
				// 		Name:        "mcp-file-write",
				// 		Description: "Write content to a file",
				// 	},
				// 	{
				// 		Name:        "mcp-file-list",
				// 		Description: "List files in a directory",
				// 	},
				// 	{
				// 		Name:        "mcp-exec",
				// 		Description: "Execute a command on the system",
				// 	},
				// }

				// // Register each tool with the server
				// // Note: mcp-go v0.32.0 doesn't have RegisterTool method
				// // We'll use the available tools from the server instead
				// for _, tool := range mcpTools {
				// 	log.Printf("Tool %s is available for use", tool.Name)
				// 	fmt.Printf("Tool available: %s\n", tool.Name)
				// }

				// Refresh the tools list after registration
				toolsResult, err = c.ListTools(ctx, toolsRequest)
				if err != nil {
					log.Printf("Failed to refresh tools list: %v", err)
				}
			}

			// Add DeepSeek LLM as a virtual tool if API key is provided
			virtualTools := toolsResult.Tools
			if *deepseekKey != "" {
				virtualTools = append(virtualTools, mcp.Tool{
					Name:        "deepseek-llm",
					Description: "DeepSeek large language model with MCP tool integration",
				})
			}

			fmt.Printf("Available tools: %d\n", len(virtualTools))
			for _, tool := range virtualTools {
				fmt.Printf("  - %s: %s\n", tool.Name, tool.Description)
			}

			// AI-driven interaction loop
			if len(virtualTools) > 0 {
				fmt.Println("\nAI Tool Manager is ready. Type 'exit' to quit.")

				// Create a longer-lived context for the conversation
				convCtx, convCancel := context.WithCancel(context.Background())
				defer convCancel()

				// Add client to context for tool execution
				ctxWithClient := context.WithValue(convCtx, "client", c)

				// Store conversation history for context
				var conversationHistory []openai.ChatCompletionMessage

				// Add system message
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

					// Show initial loading animation after input submission
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

					// Call DeepSeek API with conversation history and available tools
					response, newHistory, err := callDeepSeekAPIWithHistory(ctxWithClient, *deepseekKey, *deepseekModel, conversationHistory, virtualTools)
					if err != nil {
						log.Printf("AI API error: %v", err)
						continue
					}

					// Update conversation history
					conversationHistory = newHistory

					// Show loading animation
					fmt.Print("\n\x1b[36m") // Cyan color
					for i := 0; i < 3; i++ {
						fmt.Print(".")
						time.Sleep(200 * time.Millisecond)
					}
					fmt.Print("\x1b[0m") // Reset color

					// Print response with color
					fmt.Printf("\n\x1b[32mAI Response:\x1b[0m\n\x1b[33m%s\x1b[0m\n", response) // Green title, yellow response
				}
			}
		}
	}

	// List available resources if the server supports them
	if serverInfo.Capabilities.Resources != nil {
		fmt.Println("Fetching available resources...")
		resourcesRequest := mcp.ListResourcesRequest{}
		resourcesResult, err := c.ListResources(ctx, resourcesRequest)
		if err != nil {
			log.Printf("Failed to list resources: %v", err)
		} else {
			fmt.Printf("Server has %d resources available\n", len(resourcesResult.Resources))
			for i, resource := range resourcesResult.Resources {
				fmt.Printf("  %d. %s - %s\n", i+1, resource.URI, resource.Name)
			}
		}
	}

	fmt.Println("Client initialized successfully. Shutting down...")
	c.Close()
}

// parseCommand splits a command string into command and arguments
// callDeepSeekAPIWithHistory makes a request to DeepSeek API with conversation history and tool definitions
func callDeepSeekAPIWithHistory(ctx context.Context, apiKey string, model string, history []openai.ChatCompletionMessage, tools []mcp.Tool) (string, []openai.ChatCompletionMessage, error) {
	// Create OpenAI client configured for DeepSeek API
	config := openai.DefaultConfig(apiKey)
	config.BaseURL = "https://api.deepseek.com/v1"
	client := openai.NewClientWithConfig(config)

	// Convert MCP tools to OpenAI tool definitions
	var openaiTools []openai.Tool
	for _, tool := range tools {
		openaiTools = append(openaiTools, openai.Tool{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        tool.Name,
				Description: tool.Description,
				Parameters: map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"input": map[string]interface{}{
							"type":        "string",
							"description": "Input for the tool execution",
						},
					},
					"required": []string{"input"},
				},
			},
		})
	}

	// Create chat completion request with history
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       model,
		Messages:    history,
		Tools:       openaiTools,
		ToolChoice:  "auto",
		Temperature: 0.7,
	})
	if err != nil {
		return "", history, fmt.Errorf("failed to call API: %v", err)
	}

	if len(resp.Choices) == 0 {
		return "", history, fmt.Errorf("no choices in response")
	}

	// Get the assistant's response
	assistantMessage := resp.Choices[0].Message

	// Add assistant message to history
	newHistory := append(history, assistantMessage)

	// Handle tool calls if any
	if len(assistantMessage.ToolCalls) > 0 {
		// Show loading animation
		fmt.Print("\n\x1b[36m") // Cyan color
		for i := 0; i < 3; i++ {
			fmt.Print(".")
			time.Sleep(200 * time.Millisecond)
		}
		fmt.Print("\x1b[0m") // Reset color
		fmt.Println("\nAI is using tools to respond to your request...")

		// Process each tool call
		for _, toolCall := range assistantMessage.ToolCalls {
			toolName := toolCall.Function.Name
			toolInput := toolCall.Function.Arguments

			fmt.Printf("Executing tool: %s\n", toolName)

			// Execute the tool
			result, err := executeTool(ctx, toolName, toolInput)
			if err != nil {
				result = fmt.Sprintf("Error: %v", err)
			}

			// Add tool result to history
			newHistory = append(newHistory, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    result,
				ToolCallID: toolCall.ID,
			})
		}

		// Make a follow-up request to get the final response
		return callDeepSeekAPIFollowUp(ctx, client, model, newHistory, openaiTools)
	}

	return assistantMessage.Content, newHistory, nil
}

// callDeepSeekAPIFollowUp makes a follow-up request after tool calls
func callDeepSeekAPIFollowUp(ctx context.Context, client *openai.Client, model string, history []openai.ChatCompletionMessage, tools []openai.Tool) (string, []openai.ChatCompletionMessage, error) {
	// Create follow-up request
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       model,
		Messages:    history,
		Tools:       tools,
		ToolChoice:  "auto",
		Temperature: 0.7,
	})
	if err != nil {
		return "", history, fmt.Errorf("failed to call API for follow-up: %v", err)
	}

	if len(resp.Choices) == 0 {
		return "", history, fmt.Errorf("no choices in follow-up response")
	}

	// Get the assistant's response
	assistantMessage := resp.Choices[0].Message

	// Add assistant message to history
	newHistory := append(history, assistantMessage)

	// Check if there are more tool calls
	if len(assistantMessage.ToolCalls) > 0 {
		// Recursively handle more tool calls
		return callDeepSeekAPIFollowUp(ctx, client, model, newHistory, tools)
	}

	return assistantMessage.Content, newHistory, nil
}

// callDeepSeekAPI makes a request to DeepSeek API with tool definitions (legacy version)
func callDeepSeekAPI(ctx context.Context, apiKey string, model string, prompt string, tools []mcp.Tool) (string, error) {
	// Create initial history
	history := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "You are an AI assistant that can intelligently call MCP tools. Analyze the user's request and determine if any tools should be used. When using tools, provide clear explanations of your decisions. You can chain multiple tool calls if needed to complete complex tasks.",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: prompt,
		},
	}

	// Call with history
	response, _, err := callDeepSeekAPIWithHistory(ctx, apiKey, model, history, tools)
	return response, err
}

// executeTool handles the execution of a specific MCP tool
func executeTool(ctx context.Context, toolName string, input string) (string, error) {
	// Get the client from context
	c, ok := ctx.Value("client").(*client.Client)
	if !ok {
		return "", fmt.Errorf("client not found in context")
	}

	// Parse the input JSON if it's a JSON string
	var inputParams map[string]interface{}
	if err := json.Unmarshal([]byte(input), &inputParams); err != nil {
		// If not valid JSON, use as raw input
		inputParams = map[string]interface{}{
			"input": input,
		}
	}

	// Handle special case for deepseek-llm
	if toolName == "deepseek-llm" {
		// Just return the input as the result for this virtual tool
		return fmt.Sprintf("Processed by DeepSeek LLM: %s", input), nil
	}

	// Create CallToolRequest
	callToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      toolName,
			Arguments: inputParams,
		},
	}

	// Call the tool using CallTool method
	result, err := c.CallTool(ctx, callToolRequest)
	if err != nil {
		return "", fmt.Errorf("failed to execute tool %s: %v", toolName, err)
	}

	// Extract text content from result
	if result != nil && len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			return textContent.Text, nil
		}
		// Fallback to string representation
		returnData, err := json.Marshal(result.Content)
		if err != nil {
			return "", fmt.Errorf("failed to marshal result: %v", err)
		}
		return string(returnData), nil
	}

	return "", fmt.Errorf("empty result from tool %s", toolName)
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
