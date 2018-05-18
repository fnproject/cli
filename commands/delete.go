package commands

import (
	"github.com/fnproject/cli/client"
	"github.com/fnproject/cli/commands"
	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/objects"
	"github.com/urfave/cli"
)

func DeleteCommand() cli.Command {
	apiClient := common.FnClient{}
	return cli.Command{
		Name: "delete",
		Before: func(c *cli.Context) error {
			var err error
			apiClient.Client, err = client.APIClient()
			return err
		},
		Aliases:     []string{"d"},
		Usage:       "delete command",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: objects.getSubCommands(commands.DeleteCmd, apiClient),
	}
}
