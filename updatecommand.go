package main

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

var updateSubCommands []cli.Command
var updateAPIClient clientCmd

func updateCommand() cli.Command {
	updateAPIClient = clientCmd{}

	return cli.Command{
		Name:    "update",
		Aliases: []string{"up"},
		Usage:   "update command",
		Before: func(c *cli.Context) error {
			var err error
			updateAPIClient.client, err = client.APIClient()
			return err
		},
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: updateAPIClient.getUpdateSubCommands(),
	}
}

func (a *clientCmd) getUpdateSubCommands() []cli.Command {
	updateSubCommands = append(updateSubCommands, a.apps(appsUpdate))
	updateSubCommands = append(updateSubCommands, a.routes(routesDelete))
	updateSubCommands = append(updateSubCommands, contextCommand(contextUpdate))

	return updateSubCommands
}
