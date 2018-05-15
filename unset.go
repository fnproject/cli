package main

import (
	"github.com/urfave/cli"
)

func unsetCommand() cli.Command {
	return cli.Command{
		Name:        "unset",
		Aliases:     []string{"un"},
		Usage:       "list command",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: []cli.Command{contextCmd("unset")},
	}
}
