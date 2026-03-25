package main

// Command registry — each module registers itself via init() with build tags.

type Command struct {
	Name  string
	Desc  string
	Run   func(args []string)
}

var commands = map[string]*Command{}

func Register(cmd *Command) {
	commands[cmd.Name] = cmd
}
