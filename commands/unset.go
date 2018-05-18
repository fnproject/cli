package commands

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

func unsetCommand() cli.Command {
	apiClient := fnClient{}

	return cli.Command{
		Name:    "unset",
		Aliases: []string{"un"},
		Usage:   "unset command",
		Before: func(c *cli.Context) error {
			var err error
			apiClient.client, err = client.APIClient()
			return err
		},
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: apiClient.getSubCommands(UnsetCmd),
	}
}
