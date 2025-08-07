package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/carlosarraes/lit/internal/input"
	"github.com/carlosarraes/lit/internal/provider"
	"github.com/carlosarraes/lit/internal/tools"
)

const MODEL = "Lit"

type Agent struct {
	provider       provider.Provider
	getUserMessage func() (string, bool)
	tools          []tools.ToolDefinition
	useInteractive bool
}

func NewAgent(prov provider.Provider, getUserMessage func() (string, bool), tools []tools.ToolDefinition) *Agent {
	return &Agent{
		provider:       prov,
		getUserMessage: getUserMessage,
		tools:          tools,
		useInteractive: true,
	}
}

func (a *Agent) Run(ctx context.Context) error {
	conversation := []provider.Message{}

	fmt.Printf("Chat with %s (%s) - use 'ctrl-c' to quit\n", a.provider.GetModel(), MODEL)
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
			conversation = append(conversation, provider.Message{
				Role:    "user",
				Content: processedInput,
			})
		}

		response, err := a.provider.Chat(ctx, conversation, a.tools)
		if err != nil {
			return err
		}

		if response.Content != "" {
			fmt.Printf("\u001b[93m%s\u001b[0m: %s\n", MODEL, response.Content)
			conversation = append(conversation, provider.Message{
				Role:    "assistant",
				Content: response.Content,
			})
		}

		if len(response.ToolCalls) > 0 {
			toolResults := []string{}
			for _, toolCall := range response.ToolCalls {
				result := a.executeTool(toolCall.ID, toolCall.Name, toolCall.Input)
				toolResults = append(toolResults, result)
			}
			
			if len(toolResults) > 0 {
				toolResultsContent := strings.Join(toolResults, "\n\n")
				conversation = append(conversation, provider.Message{
					Role:    "user",
					Content: toolResultsContent,
				})
				readUserInput = false
				fmt.Println()
				continue
			}
		}

		readUserInput = true
	}

	return nil
}

func (a *Agent) executeTool(id, name string, input json.RawMessage) string {
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
		return fmt.Sprintf("Tool %s not found", name)
	}

	fmt.Printf("\u001b[92mtool\u001b[0m: %s(%s) ", name, input)
	response, err := toolDef.Function(input)
	if err != nil {
		fmt.Println("❌")
		return fmt.Sprintf("Tool %s failed: %s", name, err.Error())
	}
	fmt.Println("✅")

	return response
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
