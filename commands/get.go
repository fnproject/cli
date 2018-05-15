package commands

import (
	"github.com/urfave/cli"
)

func GetCommand() cli.Command {
	return cli.Command{
		Name:        "get",
		Aliases:     []string{"g"},
		Usage:       "get command",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<command>",
		Subcommands: GetCommands(GetCmds),
	}
}
