package commands

import (
	"github.com/fnproject/cli/common"
	"github.com/urfave/cli"
)

// InspectCommand returns inspect cli.command
func InspectCommand() cli.Command {
	return cli.Command{
		Name:         "inspect",
		UsageText:    "inspect",
		Aliases:      []string{"i"},
		Usage:        "\tRetrieve properties of an object",
		Category:     "MANAGEMENT COMMANDS",
		Hidden:       false,
		ArgsUsage:    "<subcommand>",
		Description:  "This command allows to inspect the properties of an object ('app', 'context', function' or 'trigger').",
		Subcommands:  GetCommands(InspectCmds),
		BashComplete: common.DefaultBashComplete,
	}
}
