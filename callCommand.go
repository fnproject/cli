package main

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

var callSubCommands []cli.Command
var callAPIClient clientCmd

func callCommand() cli.Command {
	callAPIClient = clientCmd{}

	return cli.Command{
		Name:    "call",
		Aliases: []string{"cl"},
		Usage:   "call command",
		Before: func(c *cli.Context) error {
			var err error
			callAPIClient.client, err = client.APIClient()
			return err
		},
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: callAPIClient.getCallSubCommands(),
	}
}

func (a *clientCmd) getCallSubCommands() []cli.Command {
	callSubCommands = append(callSubCommands, a.routes(routesCall))

	return callSubCommands
}
