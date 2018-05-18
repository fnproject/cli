package commands

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

func InspectCommand() cli.Command {
	apiClient := fnClient{}

	return cli.Command{
		Name:    "inspect",
		Aliases: []string{"i"},
		Usage:   "inspect command",
		Before: func(c *cli.Context) error {
			var err error
			apiClient.client, err = client.APIClient()
			return err
		},
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: apiClient.getSubCommands(InspectCmd),
	}
}
