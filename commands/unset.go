package commands

import (
	"github.com/fnproject/cli/client"
	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/objects"
	"github.com/urfave/cli"
)

func UnsetCommand() cli.Command {
	apiClient := common.FnClient{}

	return cli.Command{
		Name:    "unset",
		Aliases: []string{"un"},
		Usage:   "unset command",
		Before: func(c *cli.Context) error {
			var err error
			apiClient.Client, err = client.APIClient()
			return err
		},
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: objects.GetSubCommands(common.UnsetCmd, &apiClient),
	}
}
