package main

import (
	"github.com/urfave/cli"
)

func useCommand() cli.Command {
	return cli.Command{
		Name:        "use",
		Aliases:     []string{"u"},
		Usage:       "use command",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: []cli.Command{contextCmd("use")},
	}
}
