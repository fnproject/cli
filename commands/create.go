package commands

import (
	"github.com/urfave/cli"
)

func CreateCommand() cli.Command {
	return cli.Command{
		Name:        "create",
		Aliases:     []string{"c"},
		Usage:       "create",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<object>",
		Subcommands: GetCommands(CreateCmds),
	}
}
