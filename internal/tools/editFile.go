package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
)

type EditFileInput struct {
	Path   string `json:"path" jsonschema_description:"The path to the file"`
	OldStr string `json:"old_str" jsonschema_description:"Text to search for - must match exactly and must only have one match exactly"`
	NewStr string `json:"new_str" jsonschema_description:"Text to replace old_str with"`
}

var (
	readFilesMutex sync.RWMutex
	readFiles      = make(map[string]bool)
)

func MarkFileAsRead(filePath string) {
	readFilesMutex.Lock()
	defer readFilesMutex.Unlock()
	readFiles[filePath] = true
}

func hasBeenRead(filePath string) bool {
	readFilesMutex.RLock()
	defer readFilesMutex.RUnlock()
	return readFiles[filePath]
}

var (
	EditFileInputSchema = generateSchema[EditFileInput]()
	EditFileDefinition  = ToolDefinition{
		Name: "edit_file",
		Description: `Make edits to a text file.

IMPORTANT: You MUST use read_file first to see the current contents before editing any existing file. This tool will fail if you attempt to edit an existing file without reading it first.

Replaces 'old_str' with 'new_str' in the given file. 'old_str' and 'new_str' MUST be different from each other.

If the file doesn't exist, it will be created (no need to read first for new files).
`,
		InputSchema: EditFileInputSchema,
		Function:    EditFile,
	}
)

func EditFile(input json.RawMessage) (string, error) {
	editFileInput := EditFileInput{}
	if err := json.Unmarshal(input, &editFileInput); err != nil {
		return "", err
	}

	if editFileInput.Path == "" || editFileInput.OldStr == editFileInput.NewStr {
		return "", fmt.Errorf("invalid input parameters")
	}

	content, err := os.ReadFile(editFileInput.Path)
	if err != nil {
		if os.IsNotExist(err) && editFileInput.OldStr == "" {
			return createNewFile(editFileInput.Path, editFileInput.NewStr)
		}
		return "", err
	}

	if !hasBeenRead(editFileInput.Path) {
		return "", fmt.Errorf("ERROR: File '%s' exists but you haven't read it yet. You MUST use the read_file tool first to see the current contents before editing. This prevents accidental duplications or overwrites. Use: read_file with path '%s' then try edit_file again", editFileInput.Path, editFileInput.Path)
	}

	oldContent := string(content)
	newContent := strings.ReplaceAll(oldContent, editFileInput.OldStr, editFileInput.NewStr)
	if oldContent == newContent && editFileInput.OldStr != "" {
		return "", fmt.Errorf("old_str not found in file")
	}

	if err := os.WriteFile(editFileInput.Path, []byte(newContent), 0644); err != nil {
		return "", err
	}

	return "OK", nil
}

func createNewFile(filePath, content string) (string, error) {
	dir := path.Dir(filePath)
	if dir != "." {
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return "", fmt.Errorf("failed to create directory: %w", err)
		}
	}

	err := os.WriteFile(filePath, []byte(content), 0644)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}

	return fmt.Sprintf("Successfully created file %s", filePath), nil
}
