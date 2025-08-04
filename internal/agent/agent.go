package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/carlosarraes/lit/internal/input"
	"github.com/carlosarraes/lit/internal/tools"
)

const MODEL = "Lit"

type Agent struct {
	client         *anthropic.Client
	getUserMessage func() (string, bool)
	tools          []tools.ToolDefinition
	useInteractive bool
}

func NewAgent(client *anthropic.Client, getUserMessage func() (string, bool), tools []tools.ToolDefinition) *Agent {
	return &Agent{
		client:         client,
		getUserMessage: getUserMessage,
		tools:          tools,
		useInteractive: true,
	}
}

func (a *Agent) Run(ctx context.Context) error {
	conversation := []anthropic.MessageParam{}

	fmt.Println("Chat with Claude (use 'ctrl-c' to quit)")
	var interactiveInput *input.InteractiveInput
	if a.useInteractive {
		interactiveInput = input.NewInteractiveInput()
	}

	readUserInput := true
	for {
		if readUserInput {
			var userInput string
			var err error

			if a.useInteractive && interactiveInput != nil {
				userInput, err = interactiveInput.ReadLine()
				if err == io.EOF {
					fmt.Println("\nExiting chat.")
					break
				} else if err != nil {
					fmt.Printf("Input error: %v\n", err)
					continue
				}
			} else {
				fmt.Print("\u001b[94mYou\u001b[0m: ")
				input, ok := a.getUserMessage()
				if !ok {
					fmt.Println("\nExiting chat.")
					break
				}
				userInput = input
			}

			if strings.TrimSpace(userInput) == "" {
				continue
			}

			processedInput := processAtReferences(userInput)

			userMessage := anthropic.NewUserMessage(anthropic.NewTextBlock(processedInput))
			conversation = append(conversation, userMessage)
		}

		message, err := a.runInference(ctx, conversation)
		if err != nil {
			return err
		}
		conversation = append(conversation, message.ToParam())

		toolResults := []anthropic.ContentBlockParamUnion{}
		for _, content := range message.Content {
			switch content.Type {
			case "text":
				fmt.Printf("\u001b[93m%s\u001b[0m: %s\n", MODEL, content.Text)
			case "tool_use":
				result := a.executeTool(content.ID, content.Name, content.Input)
				toolResults = append(toolResults, result)
			}
		}
		if len(toolResults) == 0 {
			readUserInput = true
			continue
		}
		readUserInput = false
		fmt.Println()
		conversation = append(conversation, anthropic.NewUserMessage(toolResults...))
	}

	return nil
}

func (a *Agent) executeTool(id, name string, input json.RawMessage) anthropic.ContentBlockParamUnion {
	var toolDef tools.ToolDefinition
	var found bool

	for _, tool := range a.tools {
		if tool.Name == name {
			toolDef = tool
			found = true
			break
		}
	}

	if !found {
		return anthropic.NewToolResultBlock(id, "tool not found", true)
	}

	fmt.Printf("\u001b[92mtool\u001b[0m: %s(%s) ", name, input)
	response, err := toolDef.Function(input)
	if err != nil {
		fmt.Println("❌")
		return anthropic.NewToolResultBlock(id, err.Error(), true)
	}
	fmt.Println("✅")

	return anthropic.NewToolResultBlock(id, response, false)
}

func (a *Agent) runInference(ctx context.Context, conversation []anthropic.MessageParam) (*anthropic.Message, error) {
	anthropicTools := []anthropic.ToolUnionParam{}
	for _, tool := range a.tools {
		anthropicTools = append(anthropicTools, anthropic.ToolUnionParam{
			OfTool: &anthropic.ToolParam{
				Name:        tool.Name,
				Description: anthropic.String(tool.Description),
				InputSchema: tool.InputSchema,
			},
		})
	}

	message, err := a.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_5HaikuLatest,
		MaxTokens: 8192,
		Messages:  conversation,
		Tools:     anthropicTools,
	})
	return message, err
}

func processAtReferences(input string) string {
	re := regexp.MustCompile(`@([^\s]+)`)

	return re.ReplaceAllStringFunc(input, func(match string) string {
		path := strings.TrimPrefix(match, "@")

		if strings.Contains(path, "/") ||
			strings.Contains(path, ".") ||
			!strings.Contains(path, "@") {
			return path
		}

		return match
	})
}
