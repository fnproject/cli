package main

import (
	"github.com/fnproject/cli/client"
	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/objects"
	"github.com/urfave/cli"
)

func configCommand() cli.Command {
	apiClient := common.FnClient{}

	return cli.Command{
		Name:    "config",
		Aliases: []string{"con"},
		Usage:   "config command",
		Before: func(c *cli.Context) error {
			var err error
			apiClient.Client, err = client.APIClient()
			return err
		},
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: objects.GetSubCommands(common.ConfigCmd, &apiClient),
	}
}
