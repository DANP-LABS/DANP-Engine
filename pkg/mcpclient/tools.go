package mcpclient

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sashabaranov/go-openai"
)

type ToolManager struct {
	client *Client
}

func NewToolManager(client *Client) *ToolManager {
	return &ToolManager{client: client}
}

func (tm *ToolManager) ListTools(ctx context.Context) ([]mcp.Tool, error) {
	toolsRequest := mcp.ListToolsRequest{}
	toolsResult, err := tm.client.GetRawClient().ListTools(ctx, toolsRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to list tools: %v", err)
	}
	return toolsResult.Tools, nil
}

func (tm *ToolManager) ExecuteTool(ctx context.Context, toolName string, input string) (string, error) {
	var inputParams map[string]interface{}
	if err := json.Unmarshal([]byte(input), &inputParams); err != nil {
		inputParams = map[string]interface{}{
			"input": input,
		}
	}

	if toolName == "deepseek-llm" {
		return fmt.Sprintf("Processed by DeepSeek LLM: %s", input), nil
	}

	callToolRequest := mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Name:      toolName,
			Arguments: inputParams,
		},
	}

	result, err := tm.client.GetRawClient().CallTool(ctx, callToolRequest)
	if err != nil {
		return "", fmt.Errorf("failed to execute tool %s: %v", toolName, err)
	}

	if result != nil && len(result.Content) > 0 {
		if textContent, ok := result.Content[0].(mcp.TextContent); ok {
			return textContent.Text, nil
		}
		returnData, err := json.Marshal(result.Content)
		if err != nil {
			return "", fmt.Errorf("failed to marshal result: %v", err)
		}
		return string(returnData), nil
	}

	return "", fmt.Errorf("empty result from tool %s", toolName)
}

func (tm *ToolManager) ConvertToolsToOpenAI(tools []mcp.Tool) []openai.Tool {
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
	return openaiTools
}
