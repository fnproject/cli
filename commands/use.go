package commands

import (
	"github.com/fnproject/cli/client"
	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/objects"
	"github.com/urfave/cli"
)

func UseCommand() cli.Command {
	apiClient := common.FnClient{}

	return cli.Command{
		Name:    "use",
		Aliases: []string{"u"},
		Usage:   "use command",
		Before: func(c *cli.Context) error {
			var err error
			apiClient.Client, err = client.APIClient()
			return err
		},
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: objects.GetSubCommands(common.UseCmd, &apiClient),
	}
}
