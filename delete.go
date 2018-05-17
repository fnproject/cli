package main

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

var deleteSubCommands []cli.Command
var deleteAPIClient clientCmd

func deleteCommand() cli.Command {
	deleteAPIClient = clientCmd{}
	return cli.Command{
		Name: "delete",
		Before: func(c *cli.Context) error {
			var err error
			deleteAPIClient.client, err = client.APIClient()
			return err
		},
		Aliases:     []string{"d"},
		Usage:       "delete command",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: deleteAPIClient.getDeleteSubCommands(),
	}
}

func (a *clientCmd) getDeleteSubCommands() []cli.Command {
	deleteSubCommands = append(deleteSubCommands, a.apps(appsDelete))
	deleteSubCommands = append(deleteSubCommands, a.routes(routesDelete))
	deleteSubCommands = append(deleteSubCommands, contextCommand(contextDelete))

	return deleteSubCommands
}
