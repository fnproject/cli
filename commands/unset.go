package commands

import (
	"github.com/fnxproject/cli/common"
	"github.com/urfave/cli/v2"
)

// UnsetCommand returns unset cli.command
func UnsetCommand() *cli.Command {
	return &cli.Command{
		Name:         "unset",
		Aliases:      []string{"un"},
		Usage:        "\tUnset elements of a created object",
		Category:     "MANAGEMENT COMMANDS",
		Hidden:       false,
		ArgsUsage:    "<subcommand>",
		Description:  "This command unsets elements ('configurations') for a created object ('app', 'function' or 'context').",
		Subcommands:  GetCommands(UnsetCmds),
		BashComplete: common.DefaultBashComplete,
	}
}
