package commands

import (
	"github.com/fnproject/cli/common"
	"github.com/urfave/cli"
)

// UseCommand returns use cli.command
func UseCommand() cli.Command {
	return cli.Command{
		Name:         "use",
		Aliases:      []string{"u"},
		Usage:        "\tSelect context for further commands",
		Category:     "MANAGEMENT COMMANDS",
		Hidden:       false,
		ArgsUsage:    "<subcommand>",
		Description:  "This command uses a selected object ('context') for further invocations.",
		Subcommands:  GetCommands(UseCmds),
		BashComplete: common.DefaultBashComplete,
	}
}
