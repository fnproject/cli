package main

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

func createCommand() cli.Command {
	apiClient := fnClient{}

	return cli.Command{
		Name:    "create",
		Aliases: []string{"c"},
		Usage:   "create",
		Before: func(c *cli.Context) error {
			var err error
			apiClient.client, err = client.APIClient()
			return err
		},
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<object>",
		Subcommands: apiClient.getSubCommands(CreateCmd),
	}
}
