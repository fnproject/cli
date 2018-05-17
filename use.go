package main

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

var useSubCommands []cli.Command

func useCommand() cli.Command {
	apiClient := clientCmd{}

	return cli.Command{
		Name:    "use",
		Aliases: []string{"u"},
		Usage:   "use command",
		Before: func(c *cli.Context) error {
			var err error
			apiClient.client, err = client.APIClient()
			return err
		},
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: apiClient.getSubCommands(UseCmd),
	}
}
