package main

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

var unsetSubCommands []cli.Command
var unsetAPIClient clientCmd

func unsetCommand() cli.Command {
	unsetAPIClient = clientCmd{}

	return cli.Command{
		Name:    "unset",
		Aliases: []string{"un"},
		Usage:   "unset command",
		Before: func(c *cli.Context) error {
			var err error
			unsetAPIClient.client, err = client.APIClient()
			return err
		},
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: createAPIClient.getUnsetSubCommands(),
	}
}

func (a *clientCmd) getUnsetSubCommands() []cli.Command {
	unsetSubCommands = append(unsetSubCommands, contextCommand(contextUnset))

	return unsetSubCommands
}
