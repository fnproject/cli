package main

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

type listCommands struct {
	apps    string
	context string
}

var listSubCommands []cli.Command
var listAPIClient clientCmd

func listCommand() cli.Command {
	listAPIClient = clientCmd{}
	return cli.Command{
		Name: "list",
		Before: func(c *cli.Context) error {
			var err error
			listAPIClient.client, err = client.APIClient()
			return err
		},
		Aliases:     []string{"l"},
		Usage:       "list command",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: listAPIClient.getListSubCommands(),
	}
}

func (a *clientCmd) getListSubCommands() []cli.Command {
	listSubCommands = append(listSubCommands, a.apps(ListCmd))
	listSubCommands = append(listSubCommands, a.routes(ListCmd))
	listSubCommands = append(listSubCommands, contextCommand(ListCmd))

	return listSubCommands
}
