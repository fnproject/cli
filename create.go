package main

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

var createSubCommands []cli.Command
var createAPIClient clientCmd

func createCommand() cli.Command {
	createAPIClient = clientCmd{}

	return cli.Command{
		Name:    "create",
		Aliases: []string{"c"},
		Usage:   "create command",
		Before: func(c *cli.Context) error {
			var err error
			createAPIClient.client, err = client.APIClient()
			return err
		},
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: createAPIClient.getCreateSubCommands(),
	}
}

func (a *clientCmd) getCreateSubCommands() []cli.Command {
	createSubCommands = append(createSubCommands, a.apps(appsCreate))
	createSubCommands = append(createSubCommands, a.routes(routesCreate))
	createSubCommands = append(createSubCommands, contextCommand(contextCreate))

	return createSubCommands
}
