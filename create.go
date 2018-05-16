package main

import (
	"github.com/fnproject/cli/client"
	fnclient "github.com/fnproject/fn_go/client"
	"github.com/urfave/cli"
)

type createCmd struct {
	client *fnclient.Fn
}

var createSubCommands []cli.Command
var createAPIClient createCmd

func createCommand() cli.Command {
	createAPIClient = createCmd{}

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

func (a *createCmd) getCreateSubCommands() []cli.Command {
	createSubCommands = append(createSubCommands, a.appsCommand(appsCreate))
	createSubCommands = append(createSubCommands, contextCmd(contextCreate))

	return createSubCommands
}
