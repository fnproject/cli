package main

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

func deleteCommand() cli.Command {
	apiClient := fnClient{}
	return cli.Command{
		Name: "delete",
		Before: func(c *cli.Context) error {
			var err error
			apiClient.client, err = client.APIClient()
			return err
		},
		Aliases:     []string{"d"},
		Usage:       "delete command",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: apiClient.getSubCommands(DeleteCmd),
	}
}
