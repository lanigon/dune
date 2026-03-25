package main

import (
	"fmt"
	"os"
	"sort"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	name := os.Args[1]
	cmd, ok := commands[name]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", name)
		printUsage()
		os.Exit(1)
	}

	cmd.Run(os.Args[2:])
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage: morpheco <command> [args...]\n\nCommands:\n")

	names := make([]string, 0, len(commands))
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)

	for _, name := range names {
		fmt.Fprintf(os.Stderr, "  %-14s %s\n", name, commands[name].Desc)
	}

	fmt.Fprintf(os.Stderr, "\nEnvironment:\n  DUNE_API_KEY    Required.\n")
}
