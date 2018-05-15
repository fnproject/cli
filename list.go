package main

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/fnproject/cli/client"
	"github.com/fnproject/cli/config"
	"github.com/urfave/cli"
)

type listCommands struct {
	apps    string
	context string
}

var subCommands []cli.Command
var listAPIClient appsCmd

func listCommand() cli.Command {
	listAPIClient = appsCmd{}
	var err error
	listAPIClient.client, err = client.APIClient()
	fmt.Println("CONTEXT: ", viper.GetString(config.CurrentContext))
	if err != nil {
		fmt.Println("Error: ", err)
	}

	return cli.Command{
		Name:        "list",
		Aliases:     []string{"l"},
		Usage:       "list command",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: listAPIClient.getListSubCommands(),
	}
}

func (a *appsCmd) getListSubCommands() []cli.Command {
	subCommands = append(subCommands, a.appsCommand(appsList))

	return subCommands
}
