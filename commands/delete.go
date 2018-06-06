package commands

import (
	"github.com/urfave/cli"
)

// DeleteCommand returns delete cli.command
func DeleteCommand() cli.Command {
	return cli.Command{
		Name:        "delete",
		Aliases:     []string{"d"},
		Usage:       "delete command",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: GetCommands(DeleteCmds),
	}
}
