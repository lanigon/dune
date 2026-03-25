package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bergtatt/morpheco/scripts/internal/config"
	"github.com/bergtatt/morpheco/scripts/pkg/dune"
)

func mustClient() *dune.Client {
	cfg, err := config.Load()
	if err != nil {
		fatal(err)
	}
	return dune.New(cfg.DuneAPIKey)
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

func outputJSON(v interface{}) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	enc.Encode(v)
}
