package commands

import (
	"github.com/fnproject/cli/client"
	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/objects"
	"github.com/urfave/cli"
)

func ListCommand() cli.Command {
	apiClient := common.FnClient{}
	return cli.Command{
		Name: "list",
		Before: func(c *cli.Context) error {
			var err error
			apiClient.Client, err = client.APIClient()
			return err
		},
		Aliases:     []string{"l"},
		Usage:       "list command",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: objects.GetSubCommands(common.ListCmd, &apiClient),
	}
}
