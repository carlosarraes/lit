package main

import (
	"bufio"
	"context"
	"fmt"
	"os"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/carlosarraes/lit/internal/agent"
	"github.com/carlosarraes/lit/internal/tools"
)

var version = "dev"

func main() {
	client := anthropic.NewClient()

	scanner := bufio.NewScanner(os.Stdin)

	getUserMessage := func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}
		return scanner.Text(), true
	}

	tools := []tools.ToolDefinition{tools.ReadFileDefinition, tools.ListFilesDefinition, tools.EditFileDefinition, tools.RipgrepDefinition, tools.FdDefinition, tools.RmDefinition, tools.MvDefinition, tools.GitStatusDefinition, tools.GitAddDefinition, tools.GitCommitDefinition, tools.GitDiffDefinition}

	agent := agent.NewAgent(&client, getUserMessage, tools)
	if err := agent.Run(context.TODO()); err != nil {
		fmt.Fprintf(os.Stderr, "Error running agent: %v\n", err)
		os.Exit(1)
	}
}
