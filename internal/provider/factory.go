package provider

import (
	"fmt"
	"os"

	"github.com/carlosarraes/lit/internal/config"
)

func NewProvider(cfg *config.Config) (Provider, error) {
	switch cfg.Provider {
	case "anthropic":
		apiKey := cfg.Anthropic.APIKey
		if apiKey == "" {
			apiKey = os.Getenv("ANTHROPIC_API_KEY")
		}
		return NewAnthropicProvider(apiKey, cfg.Model), nil

	case "openai":
		apiKey := cfg.OpenAI.APIKey
		if apiKey == "" {
			apiKey = os.Getenv("OPENAI_API_KEY")
		}
		return NewOpenAIProvider(apiKey, cfg.OpenAI.BaseURL, cfg.Model), nil

	default:
		return nil, fmt.Errorf("unsupported provider: %s", cfg.Provider)
	}
}