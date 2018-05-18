package commands

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

func UseCommand() cli.Command {
	apiClient := FnClient{}

	return cli.Command{
		Name:    "use",
		Aliases: []string{"u"},
		Usage:   "use command",
		Before: func(c *cli.Context) error {
			var err error
			apiClient.client, err = client.APIClient()
			return err
		},
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: apiClient.getSubCommands(UseCmd),
	}
}
