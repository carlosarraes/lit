package tools

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type RmInput struct {
	Path      string `json:"path" jsonschema_description:"The file or directory path to remove"`
	Recursive bool   `json:"recursive,omitempty" jsonschema_description:"Remove directories and their contents recursively. Defaults to false"`
	Force     bool   `json:"force,omitempty" jsonschema_description:"Force removal without some safety checks. Defaults to false"`
}

var (
	RmInputSchema = generateSchema[RmInput]()
	RmDefinition  = ToolDefinition{
		Name: "rm",
		Description: `Remove files and directories with confirmation prompt.

SAFETY: This tool will always show what will be removed and ask for user confirmation (y/N).
The user must explicitly type 'y' or 'yes' to proceed. Defaults to 'N' (no).

Examples:
- Remove a file: path="old_file.txt"
- Remove directory recursively: path="temp_dir", recursive=true
- Force remove: path="stubborn_file", force=true
`,
		InputSchema: RmInputSchema,
		Function:    Rm,
	}
)

func Rm(input json.RawMessage) (string, error) {
	rmInput := RmInput{}
	if err := json.Unmarshal(input, &rmInput); err != nil {
		return "", err
	}

	if rmInput.Path == "" {
		return "", fmt.Errorf("path is required")
	}

	_, err := os.Stat(rmInput.Path)
	if os.IsNotExist(err) {
		return fmt.Sprintf("Path does not exist: %s", rmInput.Path), nil
	}
	if err != nil {
		return "", fmt.Errorf("error checking path: %w", err)
	}

	command := "rm"
	if rmInput.Recursive {
		command += " -r"
	}
	if rmInput.Force {
		command += " -f"
	}
	command += " " + rmInput.Path

	fmt.Printf("\n⚠️  About to execute: %s\n", command)
	fmt.Print("Are you sure you want to proceed? (y/N): ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("error reading user input: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		return "❌ Operation cancelled by user", nil
	}

	err = os.RemoveAll(rmInput.Path)
	if err != nil {
		return "", fmt.Errorf("failed to remove %s: %w", rmInput.Path, err)
	}

	return fmt.Sprintf("✅ Successfully removed: %s", rmInput.Path), nil
}
