package main

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

func inspectCommand() cli.Command {
	apiClient := fnClient{}

	return cli.Command{
		Name:    "inspect",
		Aliases: []string{"i"},
		Usage:   "inspect command",
		Before: func(c *cli.Context) error {
			var err error
			apiClient.client, err = client.APIClient()
			return err
		},
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: apiClient.getSubCommands(InspectCmd),
	}
}
