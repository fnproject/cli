package main

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

var createSubCommands []cli.Command

var newClient NewCmd

func newCreate() string {
	newClient = NewCmd{}
	return "Hello New World!"
}

func createCommand() cli.Command {
	apiClient := clientCmd{}

	return cli.Command{
		Name:    "create",
		Aliases: []string{"c"},
		Usage:   "create command",
		Before: func(c *cli.Context) error {
			var err error
			apiClient.client, err = client.APIClient()
			return err
		},
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: apiClient.getSubCommands(CreateCmd),
		//		Subcommands: createAPIClient.getCreateSubCommands(),
	}
}
