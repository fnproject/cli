package main

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

var inspectSubCommands []cli.Command
var inspectAPIClient clientCmd

func inspectCommand() cli.Command {
	inspectAPIClient = clientCmd{}

	return cli.Command{
		Name:    "inspect",
		Aliases: []string{"i"},
		Usage:   "inspect command",
		Before: func(c *cli.Context) error {
			var err error
			inspectAPIClient.client, err = client.APIClient()
			return err
		},
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: inspectAPIClient.getInspectSubCommands(),
	}
}

func (a *clientCmd) getInspectSubCommands() []cli.Command {
	inspectSubCommands = append(inspectSubCommands, a.apps(appsInspect))
	inspectSubCommands = append(inspectSubCommands, a.routes(routesInspect))

	return inspectSubCommands
}
