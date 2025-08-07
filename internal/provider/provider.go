package provider

import (
	"context"
	"encoding/json"

	"github.com/carlosarraes/lit/internal/tools"
)

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ToolCall struct {
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Input    json.RawMessage `json:"input"`
}

type Response struct {
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

type Provider interface {
	Chat(ctx context.Context, messages []Message, tools []tools.ToolDefinition) (*Response, error)
	GetModel() string
}