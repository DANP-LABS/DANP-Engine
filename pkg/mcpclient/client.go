package mcpclient

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
)

type Client struct {
	client *client.Client
}

func NewClient(ctx context.Context, stdioCmd, httpURL string) (*Client, error) {
	var c *client.Client

	if stdioCmd != "" {
		fmt.Println("Initializing stdio client...")
		args := parseCommand(stdioCmd)
		if len(args) == 0 {
			return nil, fmt.Errorf("invalid stdio command")
		}

		command := args[0]
		cmdArgs := args[1:]
		stdioTransport := transport.NewStdio(command, nil, cmdArgs...)
		c = client.NewClient(stdioTransport)

		if err := c.Start(ctx); err != nil {
			return nil, fmt.Errorf("failed to start client: %v", err)
		}

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
		httpTransport, err := transport.NewStreamableHTTP(httpURL)
		if err != nil {
			return nil, fmt.Errorf("failed to create HTTP transport: %v", err)
		}
		c = client.NewClient(httpTransport)
	}

	return &Client{client: c}, nil
}

func (c *Client) Initialize(ctx context.Context) (*mcp.InitializeResult, error) {
	initRequest := mcp.InitializeRequest{}
	initRequest.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initRequest.Params.ClientInfo = mcp.Implementation{
		Name:    "MCP-Go Client",
		Version: "1.0.0",
	}
	initRequest.Params.Capabilities = mcp.ClientCapabilities{}

	return c.client.Initialize(ctx, initRequest)
}

func (c *Client) Close() {
	c.client.Close()
}

func (c *Client) OnNotification(handler func(notification mcp.JSONRPCNotification)) {
	c.client.OnNotification(handler)
}

func (c *Client) GetRawClient() *client.Client {
	return c.client
}

func parseCommand(cmd string) []string {
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
