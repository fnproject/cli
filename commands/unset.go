package commands

import (
	"github.com/urfave/cli"
)

// UnsetCommand returns unset cli.command
func UnsetCommand() cli.Command {
	return cli.Command{
		Name:        "unset",
		Aliases:     []string{"un"},
		Usage:       "unset command",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: GetCommands(UnsetCmds),
	}
}
