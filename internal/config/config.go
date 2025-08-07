package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

type Config struct {
	Provider string `toml:"provider"`
	Model    string `toml:"model"`
	Anthropic AnthropicConfig `toml:"anthropic"`
	OpenAI    OpenAIConfig    `toml:"openai"`
}

type AnthropicConfig struct {
	APIKey string `toml:"api_key"`
}

type OpenAIConfig struct {
	APIKey string `toml:"api_key"`
	BaseURL string `toml:"base_url,omitempty"`
}

func LoadConfig() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}

	configPath := filepath.Join(homeDir, ".config", "lit.toml")

	config := &Config{
		Provider: "anthropic",
		Model:    "claude-3-5-haiku-latest",
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return config, nil
	}

	if _, err := toml.DecodeFile(configPath, config); err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return config, nil
}

func validateConfig(config *Config) error {
	switch config.Provider {
	case "anthropic":
		if config.Anthropic.APIKey == "" {
			if os.Getenv("ANTHROPIC_API_KEY") == "" {
				return fmt.Errorf("anthropic provider requires api_key in config or ANTHROPIC_API_KEY environment variable")
			}
		}
	case "openai":
		if config.OpenAI.APIKey == "" {
			if os.Getenv("OPENAI_API_KEY") == "" {
				return fmt.Errorf("openai provider requires api_key in config or OPENAI_API_KEY environment variable")
			}
		}
	default:
		return fmt.Errorf("unsupported provider: %s (supported: anthropic, openai)", config.Provider)
	}

	return nil
}

func CreateDefaultConfig() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	configDir := filepath.Join(homeDir, ".config")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(configDir, "lit.toml")
	if _, err := os.Stat(configPath); !os.IsNotExist(err) {
		return nil
	}

	defaultConfig := `# Lit Configuration File
# Choose your AI provider: "anthropic" or "openai"
provider = "anthropic"

# Model to use (provider-specific)
# Anthropic: "claude-3-5-haiku-latest", "claude-3-5-sonnet-latest", "claude-3-opus-latest"
# OpenAI: "gpt-4o", "gpt-4o-mini", "gpt-4-turbo", "gpt-3.5-turbo"
model = "claude-3-5-haiku-latest"

[anthropic]
# API key (can also be set via ANTHROPIC_API_KEY environment variable)
# api_key = "your-anthropic-api-key"

[openai]
# API key (can also be set via OPENAI_API_KEY environment variable)
# api_key = "your-openai-api-key"
# Optional: Custom base URL for OpenAI-compatible APIs
# base_url = "https://api.openai.com/v1"
`

	if err := os.WriteFile(configPath, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to create default config file: %w", err)
	}

	return nil
}