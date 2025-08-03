package tools

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type MvInput struct {
	Source      string `json:"source" jsonschema_description:"The source file or directory path to move/rename"`
	Destination string `json:"destination" jsonschema_description:"The destination path where the source should be moved/renamed to"`
	CreateDirs  bool   `json:"create_dirs,omitempty" jsonschema_description:"Create destination directories if they don't exist. Defaults to false"`
}

var (
	MvInputSchema = generateSchema[MvInput]()
	MvDefinition  = ToolDefinition{
		Name: "mv",
		Description: `Move or rename files and directories.

Moves a file or directory from source to destination. Can be used for both moving and renaming.
Will fail if destination already exists unless it's a directory and source is being moved into it.

Examples:
- Rename a file: source="old_name.txt", destination="new_name.txt"
- Move to directory: source="file.txt", destination="dir/"
- Move with rename: source="old.txt", destination="new_dir/new.txt"
- Create dirs: source="file.txt", destination="new/path/file.txt", create_dirs=true
`,
		InputSchema: MvInputSchema,
		Function:    Mv,
	}
)

func Mv(input json.RawMessage) (string, error) {
	mvInput := MvInput{}
	if err := json.Unmarshal(input, &mvInput); err != nil {
		return "", err
	}

	if mvInput.Source == "" {
		return "", fmt.Errorf("source path is required")
	}
	if mvInput.Destination == "" {
		return "", fmt.Errorf("destination path is required")
	}

	_, err := os.Stat(mvInput.Source)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("source path does not exist: %s", mvInput.Source)
	}
	if err != nil {
		return "", fmt.Errorf("error checking source path: %w", err)
	}

	if mvInput.CreateDirs {
		destDir := filepath.Dir(mvInput.Destination)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create destination directories: %w", err)
		}
	}

	destInfo, err := os.Stat(mvInput.Destination)
	if err == nil {
		if destInfo.IsDir() {
			filename := filepath.Base(mvInput.Source)
			mvInput.Destination = filepath.Join(mvInput.Destination, filename)
		} else {
			return "", fmt.Errorf("destination already exists: %s", mvInput.Destination)
		}
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("error checking destination path: %w", err)
	}

	err = os.Rename(mvInput.Source, mvInput.Destination)
	if err != nil {
		return "", fmt.Errorf("failed to move %s to %s: %w", mvInput.Source, mvInput.Destination, err)
	}

	operation := "moved"
	if filepath.Dir(mvInput.Source) == filepath.Dir(mvInput.Destination) {
		operation = "renamed"
	}

	return fmt.Sprintf("✅ Successfully %s %s to %s", operation, mvInput.Source, mvInput.Destination), nil
}
