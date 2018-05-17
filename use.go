package main

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

var useSubCommands []cli.Command
var useAPIClient clientCmd

func useCommand() cli.Command {
	useAPIClient = clientCmd{}

	return cli.Command{
		Name:    "use",
		Aliases: []string{"u"},
		Usage:   "use command",
		Before: func(c *cli.Context) error {
			var err error
			useAPIClient.client, err = client.APIClient()
			return err
		},
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: useAPIClient.getUseSubCommands(),
	}
}

func (a *clientCmd) getUseSubCommands() []cli.Command {
	useSubCommands = append(useSubCommands, contextCommand(contextUse))

	return useSubCommands
}
