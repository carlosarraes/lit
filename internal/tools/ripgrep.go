package tools

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type RipgrepInput struct {
	Pattern     string `json:"pattern" jsonschema_description:"The regex pattern to search for"`
	Path        string `json:"path,omitempty" jsonschema_description:"Optional path to search in. Defaults to current directory if not provided"`
	CaseSensitive bool `json:"case_sensitive,omitempty" jsonschema_description:"Whether the search should be case sensitive. Defaults to false"`
	WholeWord   bool   `json:"whole_word,omitempty" jsonschema_description:"Whether to match whole words only. Defaults to false"`
	FileType    string `json:"file_type,omitempty" jsonschema_description:"Optional file type filter (e.g., 'go', 'js', 'py')"`
	MaxResults  int    `json:"max_results,omitempty" jsonschema_description:"Maximum number of results to return. Defaults to 50"`
}

var (
	RipgrepInputSchema = generateSchema[RipgrepInput]()
	RipgrepDefinition  = ToolDefinition{
		Name: "ripgrep",
		Description: `Search for patterns in files using ripgrep (rg).

Searches for regex patterns across files in the specified directory (or current directory if not specified).
Returns matching lines with file names and line numbers.

Examples:
- Search for "function" in all files: pattern="function"
- Search in specific directory: pattern="TODO", path="src/"  
- Case sensitive search: pattern="Error", case_sensitive=true
- Filter by file type: pattern="import", file_type="go"
- Limit results: pattern="console.log", max_results=10
`,
		InputSchema: RipgrepInputSchema,
		Function:    Ripgrep,
	}
)

func Ripgrep(input json.RawMessage) (string, error) {
	ripgrepInput := RipgrepInput{}
	if err := json.Unmarshal(input, &ripgrepInput); err != nil {
		return "", err
	}

	if ripgrepInput.Pattern == "" {
		return "", fmt.Errorf("pattern is required")
	}

	args := []string{}
	
	if !ripgrepInput.CaseSensitive {
		args = append(args, "-i")
	}
	
	if ripgrepInput.WholeWord {
		args = append(args, "-w")
	}
	
	if ripgrepInput.FileType != "" {
		args = append(args, "-t", ripgrepInput.FileType)
	}
	
	args = append(args, "-n", "-H")
	
	maxResults := ripgrepInput.MaxResults
	if maxResults == 0 {
		maxResults = 50
	}
	args = append(args, "-m", fmt.Sprintf("%d", maxResults))
	
	args = append(args, ripgrepInput.Pattern)
	
	if ripgrepInput.Path != "" {
		args = append(args, ripgrepInput.Path)
	} else {
		args = append(args, ".")
	}

	cmd := exec.Command("rg", args...)
	output, err := cmd.Output()
	if err != nil {
		if err.Error() == "exec: \"rg\" not found in $PATH" {
			return "", fmt.Errorf("ripgrep (rg) is not installed. Please install it first")
		}
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return "No matches found", nil
		}
		return "", fmt.Errorf("ripgrep error: %w", err)
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return "No matches found", nil
	}

	return result, nil
}
