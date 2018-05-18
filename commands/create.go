package commands

import (
	"github.com/fnproject/cli/client"
	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/objects"
	"github.com/urfave/cli"
)

func CreateCommand() cli.Command {
	apiClient := common.FnClient{}

	return cli.Command{
		Name:    "create",
		Aliases: []string{"c"},
		Usage:   "create",
		Before: func(c *cli.Context) error {
			var err error
			apiClient.Client, err = client.APIClient()
			return err
		},
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<object>",
		Subcommands: objects.GetSubCommands(common.CreateCmd, &apiClient),
	}
}
