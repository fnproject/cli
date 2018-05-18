package main

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

func callCommand() cli.Command {
	apiClient := fnClient{}

	return cli.Command{
		Name:    "call",
		Aliases: []string{"cl"},
		Usage:   "call command",
		Before: func(c *cli.Context) error {
			var err error
			apiClient.Client, err = client.APIClient()
			return err
		},
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: apiClient.getSubCommands(CallCmd),
	}
}
