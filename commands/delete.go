package commands

import (
	"github.com/urfave/cli"
)

// DeleteCommand returns delete cli.command
func DeleteCommand() cli.Command {
	return cli.Command{
		Name:        "delete",
		Aliases:     []string{"d"},
		Usage:       "Delete an object",
		Category:    "MANAGEMENT COMMANDS",
		Description: "This is the description",
		Hidden:      false,
		ArgsUsage:   "<subcommand>",
		Subcommands: GetCommands(DeleteCmds),
	}
}
