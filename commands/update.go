package commands

import (
	"github.com/fnxproject/cli/common"
	"github.com/urfave/cli/v2"
)

// UpdateCommand returns update cli.command
func UpdateCommand() *cli.Command {
	return &cli.Command{
		Name:         "update",
		Aliases:      []string{"up"},
		Usage:        "\tUpdate a created object",
		Category:     "MANAGEMENT COMMANDS",
		Hidden:       false,
		ArgsUsage:    "<subcommand>",
		Description:  "This command updates an object ('app', 'context', 'function', 'server' or 'trigger').",
		Subcommands:  GetCommands(UpdateCmds),
		BashComplete: common.DefaultBashComplete,
	}
}
