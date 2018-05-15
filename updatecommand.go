package main

import (
	"github.com/urfave/cli"
)

func updateCommand() cli.Command {
	return cli.Command{
		Name:        "update",
		Aliases:     []string{"u"},
		Usage:       "update command",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: []cli.Command{contextCmd("update")},
		Action:      func(c *cli.Context) error { return nil },
	}
}
