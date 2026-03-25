package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Config struct {
	DuneAPIKey string
	BaseURL    string
}

func Load() (*Config, error) {
	// Auto-load .env file if it exists (walk up to find project root).
	loadEnvFile()

	apiKey := os.Getenv("DUNE_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("DUNE_API_KEY environment variable is required\nCreate a .env file or export DUNE_API_KEY=<your-key>")
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

// loadEnvFile reads .env from current dir or parent dirs, sets vars that aren't already set.
func loadEnvFile() {
	paths := []string{".env", "../.env", "../../.env"}
	for _, p := range paths {
		if f, err := os.Open(p); err == nil {
			parseEnv(f)
			f.Close()
			return
		}
	}
}

func parseEnv(f *os.File) {
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		k, v, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		// Strip surrounding quotes.
		if len(v) >= 2 && ((v[0] == '"' && v[len(v)-1] == '"') || (v[0] == '\'' && v[len(v)-1] == '\'')) {
			v = v[1 : len(v)-1]
		}
		// Don't override existing env vars.
		if os.Getenv(k) == "" {
			os.Setenv(k, v)
		}
	}
}
