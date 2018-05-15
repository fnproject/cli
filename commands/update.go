package commands

import (
	"github.com/urfave/cli"
)

func UpdateCommand() cli.Command {
	return cli.Command{
		Name:        "update",
		Aliases:     []string{"up"},
		Usage:       "update command",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: GetCommands(UpdateCmds),
	}
}
