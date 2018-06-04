package commands

import (
	"github.com/urfave/cli"
)

// InspectCommand returns inspect cli.command
func InspectCommand() cli.Command {
	return cli.Command{
		Name:        "inspect",
		Aliases:     []string{"i"},
		Usage:       "inspect command",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: GetCommands(InspectCmds),
	}
}
