package main

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

//var unsetSubCommands []cli.Command

func unsetCommand() cli.Command {
	apiClient := clientCmd{}

	return cli.Command{
		Name:    "unset",
		Aliases: []string{"un"},
		Usage:   "unset command",
		Before: func(c *cli.Context) error {
			var err error
			apiClient.client, err = client.APIClient()
			return err
		},
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: apiClient.getSubCommands(UnsetCmd),
	}
}
