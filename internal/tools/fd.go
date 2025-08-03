package tools

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type FdInput struct {
	Pattern     string `json:"pattern,omitempty" jsonschema_description:"The pattern to search for in file/directory names. Can be a regex or simple string"`
	Path        string `json:"path,omitempty" jsonschema_description:"Optional path to search in. Defaults to current directory if not provided"`
	Type        string `json:"type,omitempty" jsonschema_description:"Filter by type: 'f' for files, 'd' for directories, 'l' for symlinks"`
	Extension   string `json:"extension,omitempty" jsonschema_description:"Filter by file extension (e.g., 'go', 'js', 'py')"`
	CaseSensitive bool `json:"case_sensitive,omitempty" jsonschema_description:"Whether the search should be case sensitive. Defaults to false"`
	HiddenFiles bool   `json:"hidden_files,omitempty" jsonschema_description:"Include hidden files and directories. Defaults to false"`
	MaxDepth    int    `json:"max_depth,omitempty" jsonschema_description:"Maximum search depth. 0 means no limit"`
	MaxResults  int    `json:"max_results,omitempty" jsonschema_description:"Maximum number of results to return. Defaults to 100"`
}

var (
	FdInputSchema = generateSchema[FdInput]()
	FdDefinition  = ToolDefinition{
		Name: "fd",
		Description: `Find files and directories by name using fd.

Searches for files and directories by name patterns in the specified directory (or current directory if not specified).
Returns matching file/directory paths.

Examples:
- Find all files: no pattern needed
- Find files by name: pattern="main.go"
- Find by pattern: pattern=".*test.*"
- Find in specific directory: pattern="config", path="src/"
- Find only files: pattern="*.js", type="f"
- Find only directories: type="d"
- Find by extension: extension="go"
- Case sensitive search: pattern="Main", case_sensitive=true
- Include hidden files: hidden_files=true
- Limit depth: max_depth=2
- Limit results: max_results=50
`,
		InputSchema: FdInputSchema,
		Function:    Fd,
	}
)

func Fd(input json.RawMessage) (string, error) {
	fdInput := FdInput{}
	if err := json.Unmarshal(input, &fdInput); err != nil {
		return "", err
	}

	args := []string{}

	if !fdInput.CaseSensitive {
		args = append(args, "-i")
	}

	if fdInput.HiddenFiles {
		args = append(args, "-H")
	}

	if fdInput.Type != "" {
		args = append(args, "-t", fdInput.Type)
	}

	if fdInput.Extension != "" {
		args = append(args, "-e", fdInput.Extension)
	}

	if fdInput.MaxDepth > 0 {
		args = append(args, "-d", fmt.Sprintf("%d", fdInput.MaxDepth))
	}

	maxResults := fdInput.MaxResults
	if maxResults == 0 {
		maxResults = 100
	}

	if fdInput.Pattern != "" {
		args = append(args, fdInput.Pattern)
	}

	if fdInput.Path != "" {
		args = append(args, fdInput.Path)
	} else {
		args = append(args, ".")
	}

	cmd := exec.Command("fd", args...)
	output, err := cmd.Output()
	if err != nil {
		if err.Error() == "exec: \"fd\" not found in $PATH" {
			return "", fmt.Errorf("fd is not installed. Please install it first")
		}
		if exitError, ok := err.(*exec.ExitError); ok && exitError.ExitCode() == 1 {
			return "No matches found", nil
		}
		return "", fmt.Errorf("fd error: %w", err)
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return "No matches found", nil
	}

	lines := strings.Split(result, "\n")
	if len(lines) > maxResults {
		lines = lines[:maxResults]
		result = strings.Join(lines, "\n") + fmt.Sprintf("\n... (showing first %d results)", maxResults)
	}

	return result, nil
}
