package tools

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type GitStatusInput struct {
	Short bool `json:"short,omitempty" jsonschema_description:"Show status in short format. Defaults to false"`
}

var (
	GitStatusInputSchema = generateSchema[GitStatusInput]()
	GitStatusDefinition  = ToolDefinition{
		Name: "git_status",
		Description: `Show git working tree status.

Displays the status of files in the working directory and staging area.
Shows which files are modified, staged, untracked, etc.

Examples:
- Normal status: (no parameters needed)
- Short format: short=true
`,
		InputSchema: GitStatusInputSchema,
		Function:    GitStatus,
	}
)

func GitStatus(input json.RawMessage) (string, error) {
	gitInput := GitStatusInput{}
	if err := json.Unmarshal(input, &gitInput); err != nil {
		return "", err
	}

	args := []string{"status"}
	if gitInput.Short {
		args = append(args, "--short")
	}

	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		if err.Error() == "exec: \"git\" not found in $PATH" {
			return "", fmt.Errorf("git is not installed")
		}
		if exitError, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git error: %s", string(exitError.Stderr))
		}
		return "", fmt.Errorf("git error: %w", err)
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		return "Working tree clean", nil
	}

	return result, nil
}

type GitAddInput struct {
	Paths []string `json:"paths" jsonschema_description:"List of file/directory paths to add to staging area"`
	All   bool     `json:"all,omitempty" jsonschema_description:"Add all modified and untracked files. Defaults to false"`
}

var (
	GitAddInputSchema = generateSchema[GitAddInput]()
	GitAddDefinition  = ToolDefinition{
		Name: "git_add",
		Description: `Add files to git staging area.

Adds specified files or all changes to the staging area for the next commit.

Examples:
- Add specific files: paths=["file1.txt", "dir/file2.go"]
- Add all changes: all=true
- Add current directory: paths=["."]
`,
		InputSchema: GitAddInputSchema,
		Function:    GitAdd,
	}
)

func GitAdd(input json.RawMessage) (string, error) {
	gitInput := GitAddInput{}
	if err := json.Unmarshal(input, &gitInput); err != nil {
		return "", err
	}

	args := []string{"add"}

	if gitInput.All {
		args = append(args, "-A")
	} else if len(gitInput.Paths) == 0 {
		return "", fmt.Errorf("either paths must be specified or all must be true")
	} else {
		args = append(args, gitInput.Paths...)
	}

	cmd := exec.Command("git", args...)
	_, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git add error: %s", string(exitError.Stderr))
		}
		return "", fmt.Errorf("git add error: %w", err)
	}

	if gitInput.All {
		return "✅ Added all changes to staging area", nil
	}
	return fmt.Sprintf("✅ Added %v to staging area", gitInput.Paths), nil
}

type GitCommitInput struct {
	Message string `json:"message" jsonschema_description:"Commit message"`
	Amend   bool   `json:"amend,omitempty" jsonschema_description:"Amend the previous commit. Defaults to false"`
}

var (
	GitCommitInputSchema = generateSchema[GitCommitInput]()
	GitCommitDefinition  = ToolDefinition{
		Name: "git_commit",
		Description: `Create a git commit.

Creates a commit with the specified message. Requires files to be staged first.

Examples:
- Simple commit: message="Add new feature"
- Amend previous: message="Updated message", amend=true
`,
		InputSchema: GitCommitInputSchema,
		Function:    GitCommit,
	}
)

func GitCommit(input json.RawMessage) (string, error) {
	gitInput := GitCommitInput{}
	if err := json.Unmarshal(input, &gitInput); err != nil {
		return "", err
	}

	if gitInput.Message == "" {
		return "", fmt.Errorf("commit message is required")
	}

	args := []string{"commit", "-m", gitInput.Message}
	if gitInput.Amend {
		args = []string{"commit", "--amend", "-m", gitInput.Message}
	}

	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git commit error: %s", string(exitError.Stderr))
		}
		return "", fmt.Errorf("git commit error: %w", err)
	}

	result := strings.TrimSpace(string(output))
	return fmt.Sprintf("✅ %s", result), nil
}

type GitDiffInput struct {
	Staged   bool     `json:"staged,omitempty" jsonschema_description:"Show staged changes instead of working directory changes. Defaults to false"`
	Paths    []string `json:"paths,omitempty" jsonschema_description:"Specific files/paths to show diff for"`
	NameOnly bool     `json:"name_only,omitempty" jsonschema_description:"Show only file names that changed. Defaults to false"`
}

var (
	GitDiffInputSchema = generateSchema[GitDiffInput]()
	GitDiffDefinition  = ToolDefinition{
		Name: "git_diff",
		Description: `Show git diff of changes.

Shows differences between working directory and index, or between index and HEAD.

Examples:
- Working directory changes: (no parameters)
- Staged changes: staged=true
- Specific files: paths=["file1.txt", "src/"]
- Just file names: name_only=true
- Staged file names: staged=true, name_only=true
`,
		InputSchema: GitDiffInputSchema,
		Function:    GitDiff,
	}
)

func GitDiff(input json.RawMessage) (string, error) {
	gitInput := GitDiffInput{}
	if err := json.Unmarshal(input, &gitInput); err != nil {
		return "", err
	}

	args := []string{"diff"}

	if gitInput.Staged {
		args = append(args, "--staged")
	}

	if gitInput.NameOnly {
		args = append(args, "--name-only")
	}

	if len(gitInput.Paths) > 0 {
		args = append(args, "--")
		args = append(args, gitInput.Paths...)
	}

	cmd := exec.Command("git", args...)
	output, err := cmd.Output()
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git diff error: %s", string(exitError.Stderr))
		}
		return "", fmt.Errorf("git diff error: %w", err)
	}

	result := strings.TrimSpace(string(output))
	if result == "" {
		if gitInput.Staged {
			return "No staged changes", nil
		}
		return "No changes in working directory", nil
	}

	return result, nil
}
