package provider

import (
	"context"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
	"github.com/carlosarraes/lit/internal/tools"
)

type AnthropicProvider struct {
	client anthropic.Client
	model  string
}

func NewAnthropicProvider(apiKey, model string) *AnthropicProvider {
	var client anthropic.Client
	if apiKey != "" {
		client = anthropic.NewClient(option.WithAPIKey(apiKey))
	} else {
		client = anthropic.NewClient()
	}

	if model == "" {
		model = "claude-3-5-haiku-latest"
	}

	return &AnthropicProvider{
		client: client,
		model:  model,
	}
}

func (p *AnthropicProvider) Chat(ctx context.Context, messages []Message, tools []tools.ToolDefinition) (*Response, error) {
	anthropicMessages := make([]anthropic.MessageParam, 0, len(messages))

	for _, msg := range messages {
		switch strings.ToLower(msg.Role) {
		case "user":
			userMessage := anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content))
			anthropicMessages = append(anthropicMessages, userMessage)
		case "assistant":
			assistantMessage := anthropic.NewAssistantMessage(anthropic.NewTextBlock(msg.Content))
			anthropicMessages = append(anthropicMessages, assistantMessage)
		default:
			userMessage := anthropic.NewUserMessage(anthropic.NewTextBlock(msg.Content))
			anthropicMessages = append(anthropicMessages, userMessage)
		}
	}

	anthropicTools := make([]anthropic.ToolUnionParam, 0, len(tools))
	for _, tool := range tools {
		anthropicTools = append(anthropicTools, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Name,
				Description: anthropic.String(tool.Description),
				InputSchema: tool.InputSchema,
			},
		})
	}

	modelName := anthropic.ModelClaude3_5HaikuLatest
	if p.model == "claude-3-5-sonnet-latest" {
		modelName = anthropic.ModelClaude3_5SonnetLatest
	} else if p.model == "claude-3-opus-latest" {
		modelName = anthropic.ModelClaude3OpusLatest
	}

	response, err := p.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     modelName,
		MaxTokens: 8192,
		Messages:  anthropicMessages,
		Tools:     anthropicTools,
	})
	if err != nil {
		return nil, err
	}

	result := &Response{}
	toolCalls := make([]ToolCall, 0)

	for _, content := range response.Content {
		switch content.Type {
		case "text":
			if result.Content == "" {
				result.Content = content.Text
			} else {
				result.Content += "\n" + content.Text
			}
		case "tool_use":
			toolCalls = append(toolCalls, ToolCall{
				ID:    content.ID,
				Name:  content.Name,
				Input: content.Input,
			})
		}
	}

	result.ToolCalls = toolCalls
	return result, nil
}

func (p *AnthropicProvider) GetModel() string {
	return p.model
}