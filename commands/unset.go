package commands

import (
	"github.com/urfave/cli"
)

// UnsetCommand returns unset cli.command
func UnsetCommand() cli.Command {
	return cli.Command{
		Name:        "unset",
		Aliases:     []string{"un"},
		Usage:       "\tUnset elements of a created object",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<subcommand>",
		Description: "This command unsets elements ('configurations') for a created object ('app', 'route' or 'context').",
		Subcommands: GetCommands(UnsetCmds),
	}
}
