package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/carlosarraes/lit/internal/agent"
	"github.com/carlosarraes/lit/internal/config"
	"github.com/carlosarraes/lit/internal/provider"
	"github.com/carlosarraes/lit/internal/tools"
)

var version = "dev"

func main() {
	var initConfig bool
	flag.BoolVar(&initConfig, "init", false, "Create default configuration file")
	flag.Parse()

	if initConfig {
		if err := config.CreateDefaultConfig(); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating default config: %v\n", err)
			os.Exit(1)
		}
		homeDir, _ := os.UserHomeDir()
		fmt.Printf("Default configuration created at %s/.config/lit.toml\n", homeDir)
		fmt.Println("Please edit the configuration file to set your API key and preferences.")
		return
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	prov, err := provider.NewProvider(cfg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating provider: %v\n", err)
		os.Exit(1)
	}

	scanner := bufio.NewScanner(os.Stdin)

	getUserMessage := func() (string, bool) {
		if !scanner.Scan() {
			return "", false
		}
		return scanner.Text(), true
	}

	tools := []tools.ToolDefinition{
		tools.ReadFileDefinition,
		tools.ListFilesDefinition,
		tools.EditFileDefinition,
		tools.RipgrepDefinition,
		tools.FdDefinition,
		tools.RmDefinition,
		tools.MvDefinition,
		tools.GitStatusDefinition,
		tools.GitAddDefinition,
		tools.GitCommitDefinition,
		tools.GitDiffDefinition,
	}

	agent := agent.NewAgent(prov, getUserMessage, tools)
	if err := agent.Run(context.TODO()); err != nil {
		fmt.Fprintf(os.Stderr, "Error running agent: %v\n", err)
		os.Exit(1)
	}
}
