package main

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

func updateCommand() cli.Command {
	apiClient := fnClient{}

	return cli.Command{
		Name:    "update",
		Aliases: []string{"up"},
		Usage:   "update command",
		Before: func(c *cli.Context) error {
			var err error
			apiClient.client, err = client.APIClient()
			return err
		},
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: apiClient.getSubCommands(UpdateCmd),
	}
}
