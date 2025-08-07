package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
	
	"github.com/carlosarraes/lit/internal/tools"
)

type OpenAIProvider struct {
	client *openai.Client
	model  string
}

func NewOpenAIProvider(apiKey, baseURL, model string) *OpenAIProvider {
	config := openai.DefaultConfig(apiKey)
	if baseURL != "" {
		config.BaseURL = baseURL
	}

	client := openai.NewClientWithConfig(config)

	if model == "" {
		model = "gpt-4o-mini"
	}

	return &OpenAIProvider{
		client: client,
		model:  model,
	}
}

func (p *OpenAIProvider) Chat(ctx context.Context, messages []Message, toolDefs []tools.ToolDefinition) (*Response, error) {
	openaiMessages := make([]openai.ChatCompletionMessage, 0, len(messages))

	for _, msg := range messages {
		role := strings.ToLower(msg.Role)
		switch role {
		case "user":
			openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: msg.Content,
			})
		case "assistant":
			openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleAssistant,
				Content: msg.Content,
			})
		default:
			openaiMessages = append(openaiMessages, openai.ChatCompletionMessage{
				Role:    openai.ChatMessageRoleUser,
				Content: msg.Content,
			})
		}
	}

	openaiTools := make([]openai.Tool, 0, len(toolDefs))
	for _, toolDef := range toolDefs {
		schema, err := convertToOpenAISchema(toolDef.InputSchema)
		if err != nil {
			return nil, fmt.Errorf("failed to convert schema for tool %s: %w", toolDef.Name, err)
		}

		function := openai.FunctionDefinition{
			Name:        toolDef.Name,
			Description: toolDef.Description,
			Parameters:  schema,
		}

		tool := openai.Tool{
			Type:     openai.ToolTypeFunction,
			Function: &function,
		}

		openaiTools = append(openaiTools, tool)
	}

	request := openai.ChatCompletionRequest{
		Model:    p.model,
		Messages: openaiMessages,
		Tools:    openaiTools,
	}

	response, err := p.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return nil, err
	}

	if len(response.Choices) == 0 {
		return nil, fmt.Errorf("no response choices returned")
	}

	choice := response.Choices[0]
	result := &Response{
		Content: choice.Message.Content,
	}

	if len(choice.Message.ToolCalls) > 0 {
		toolCalls := make([]ToolCall, 0, len(choice.Message.ToolCalls))
		for _, tc := range choice.Message.ToolCalls {
			if tc.Function.Arguments != "" {
				toolCalls = append(toolCalls, ToolCall{
					ID:    tc.ID,
					Name:  tc.Function.Name,
					Input: json.RawMessage(tc.Function.Arguments),
				})
			}
		}
		result.ToolCalls = toolCalls
	}

	return result, nil
}

func (p *OpenAIProvider) GetModel() string {
	return p.model
}

func convertToOpenAISchema(schema interface{}) (jsonschema.Definition, error) {
	jsonBytes, err := json.Marshal(schema)
	if err != nil {
		return jsonschema.Definition{}, err
	}

	var result jsonschema.Definition
	if err := json.Unmarshal(jsonBytes, &result); err != nil {
		return jsonschema.Definition{}, err
	}

	return result, nil
}