package config

import (
	"fmt"
	"os"
)

type Config struct {
	DuneAPIKey string
	BaseURL    string
}

func Load() (*Config, error) {
	apiKey := os.Getenv("DUNE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("DUNE_API_KEY environment variable is required")
	}

	baseURL := os.Getenv("DUNE_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.dune.com/api/v1"
	}

	return &Config{
		DuneAPIKey: apiKey,
		BaseURL:    baseURL,
	}, nil
}
