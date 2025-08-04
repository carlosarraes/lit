package tools

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

type ReadFileInput struct {
	Path   string `json:"path" jsonschema_description:"The relative path of a file in the working directory."`
	Offset int    `json:"offset,omitempty" jsonschema_description:"Starting line number (1-based). Use for large files to read specific sections."`
	Limit  int    `json:"limit,omitempty" jsonschema_description:"Maximum number of lines to read. Use for large files to avoid token limits."`
}

var (
	ReadFileInputSchema = generateSchema[ReadFileInput]()
	ReadFileDefinition  = ToolDefinition{
		Name:        "read_file",
		Description: "Read the contents of a given relative file path. For large files (>10k lines), automatically reads first 2000 lines. Use offset and limit parameters for specific sections. Do not use this with directory names.",
		InputSchema: ReadFileInputSchema,
		Function:    ReadFile,
	}
)

func ReadFile(input json.RawMessage) (string, error) {
	readFileInput := ReadFileInput{}
	if err := json.Unmarshal(input, &readFileInput); err != nil {
		panic(err)
	}

	file, err := os.Open(readFileInput.Path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	totalLines := 0
	for scanner.Scan() {
		totalLines++
	}
	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error counting lines: %w", err)
	}

	file.Seek(0, 0)

	var offset, limit int
	if readFileInput.Offset > 0 {
		offset = readFileInput.Offset
	} else {
		offset = 1
	}

	if readFileInput.Limit > 0 {
		limit = readFileInput.Limit
	} else if totalLines > 10000 {
		limit = 2000
	} else {
		limit = totalLines
	}

	scanner = bufio.NewScanner(file)
	var lines []string
	currentLine := 1
	linesRead := 0

	for scanner.Scan() && linesRead < limit {
		if currentLine >= offset {
			lines = append(lines, fmt.Sprintf("%5d\t%s", currentLine, scanner.Text()))
			linesRead++
		}
		currentLine++
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	MarkFileAsRead(readFileInput.Path)

	result := strings.Join(lines, "\n")
	
	if totalLines > 10000 && readFileInput.Limit == 0 && readFileInput.Offset == 0 {
		result += fmt.Sprintf("\n\n⚠️  Large file detected (%d total lines). Showing first %d lines.\n", totalLines, limit)
		result += "💡 Navigation tips:\n"
		result += fmt.Sprintf("   - Next section: offset=%d, limit=2000\n", offset+limit)
		result += fmt.Sprintf("   - Last section: offset=%d, limit=2000\n", max(1, totalLines-2000+1))
		result += "   - Use ripgrep first to find specific functions/patterns, then read around those lines\n"
		result += "   - Keep reading sections until you find the information you need"
	} else if readFileInput.Offset > 0 || readFileInput.Limit > 0 {
		endLine := offset + linesRead - 1
		result += fmt.Sprintf("\n\n📍 Showing lines %d-%d of %d total lines.\n", offset, endLine, totalLines)
		
		if offset > 1 {
			result += fmt.Sprintf("   - Previous section: offset=%d, limit=%d\n", max(1, offset-limit), limit)
		}
		if endLine < totalLines {
			result += fmt.Sprintf("   - Next section: offset=%d, limit=%d\n", endLine+1, limit)
		}
		if totalLines > endLine {
			result += "   - Continue reading sections to explore more of the file"
		}
	}

	return result, nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
