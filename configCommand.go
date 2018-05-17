package main

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

var configSubCommands []cli.Command
var configAPIClient clientCmd

func configCommand() cli.Command {
	configAPIClient = clientCmd{}

	return cli.Command{
		Name:    "config",
		Aliases: []string{"con"},
		Usage:   "config command",
		Before: func(c *cli.Context) error {
			var err error
			configAPIClient.client, err = client.APIClient()
			return err
		},
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: configAPIClient.getConfigSubCommands(),
	}
}

func (a *clientCmd) getConfigSubCommands() []cli.Command {
	configSubCommands = append(configSubCommands, a.apps(ConfigCmd))
	configSubCommands = append(configSubCommands, a.routes(ConfigCmd))

	return configSubCommands
}
