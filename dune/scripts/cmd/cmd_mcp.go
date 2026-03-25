//go:build mcp || all

package main

import (
	"fmt"
	"os"

	"github.com/bergtatt/morpheco/scripts/pkg/mcp"
)

func init() {
	Register(&Command{Name: "mcp", Desc: "Start MCP server (stdio)", Run: runMCP})
}

func runMCP(args []string) {
	client := mustClient()
	server := mcp.NewServer(client)

	fmt.Fprintln(os.Stderr, "dune-local MCP server started (stdio transport)")

	if err := server.Run(); err != nil {
		fatal(err)
	}
}
