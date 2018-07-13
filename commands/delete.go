package commands

import (
	"github.com/urfave/cli"
)

// DeleteCommand returns delete cli.command
func DeleteCommand() cli.Command {
	return cli.Command{
		Name:        "delete",
		Aliases:     []string{"d"},
		Usage:       "\tDelete an object",
		Category:    "MANAGEMENT COMMANDS",
		Description: "This command deletes a created object ('app', 'context', 'function', 'route' or 'trigger').",
		Hidden:      false,
		ArgsUsage:   "<subcommand>",
		Subcommands: GetCommands(DeleteCmds),
	}
}
