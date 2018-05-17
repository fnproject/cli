package main

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

func listCommand() cli.Command {
	apiClient := fnClient{}
	return cli.Command{
		Name: "list",
		Before: func(c *cli.Context) error {
			var err error
			apiClient.client, err = client.APIClient()
			return err
		},
		Aliases:     []string{"l"},
		Usage:       "list command",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: apiClient.getSubCommands(ListCmd),
	}
}
