package mcpclient

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sashabaranov/go-openai"
)

type DeepSeekManager struct {
	apiKey  string
	model   string
	tm      *ToolManager
}

func NewDeepSeekManager(apiKey, model string, tm *ToolManager) *DeepSeekManager {
	return &DeepSeekManager{
		apiKey: apiKey,
		model: model,
		tm: tm,
	}
}

func (ds *DeepSeekManager) CallWithHistory(ctx context.Context, history []openai.ChatCompletionMessage, tools []mcp.Tool) (string, []openai.ChatCompletionMessage, error) {
	config := openai.DefaultConfig(ds.apiKey)
	config.BaseURL = "https://api.deepseek.com/v1"
	client := openai.NewClientWithConfig(config)

	openaiTools := ds.tm.ConvertToolsToOpenAI(tools)

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       ds.model,
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

	assistantMessage := resp.Choices[0].Message
	newHistory := append(history, assistantMessage)

	if len(assistantMessage.ToolCalls) > 0 {
		for _, toolCall := range assistantMessage.ToolCalls {
			toolName := toolCall.Function.Name
			toolInput := toolCall.Function.Arguments

			result, err := ds.tm.ExecuteTool(ctx, toolName, toolInput)
			if err != nil {
				result = fmt.Sprintf("Error: %v", err)
			}

			newHistory = append(newHistory, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    result,
				ToolCallID: toolCall.ID,
			})
		}
		return ds.CallFollowUp(ctx, client, newHistory, openaiTools)
	}

	return assistantMessage.Content, newHistory, nil
}

func (ds *DeepSeekManager) CallFollowUp(ctx context.Context, client *openai.Client, history []openai.ChatCompletionMessage, tools []openai.Tool) (string, []openai.ChatCompletionMessage, error) {
	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       ds.model,
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

	assistantMessage := resp.Choices[0].Message
	newHistory := append(history, assistantMessage)

	if len(assistantMessage.ToolCalls) > 0 {
		return ds.CallFollowUp(ctx, client, newHistory, tools)
	}

	return assistantMessage.Content, newHistory, nil
}
