package main

import (
	"fmt"

	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

var deleteSubCommands []cli.Command
var deleteAPIClient appsCmd

func deleteCommand() cli.Command {
	deleteAPIClient = appsCmd{}
	var err error
	deleteAPIClient.client, err = client.APIClient()
	if err != nil {
		fmt.Println("Error: ", err)
	}

	return cli.Command{
		Name:        "delete",
		Aliases:     []string{"d"},
		Usage:       "delete command",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: deleteAPIClient.getDeleteSubCommands(),
	}
}

func (a *appsCmd) getDeleteSubCommands() []cli.Command {

	deleteSubCommands = append(deleteSubCommands, a.appsCommand(appsDelete))
	deleteSubCommands = append(deleteSubCommands, contextCmd(contextDelete))

	return deleteSubCommands
}
