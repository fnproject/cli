package main

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

var configSubCommands []cli.Command

func configCommand() cli.Command {
	apiClient := clientCmd{}

	return cli.Command{
		Name:    "config",
		Aliases: []string{"con"},
		Usage:   "config command",
		Before: func(c *cli.Context) error {
			var err error
			apiClient.client, err = client.APIClient()
			return err
		},
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: apiClient.getSubCommands(ConfigCmd),
	}
}
