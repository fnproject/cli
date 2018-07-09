package commands

import (
	"github.com/urfave/cli"
)

// ListCommand returns list cli.command
func ListCommand() cli.Command {
	return cli.Command{
		Name:        "list",
		Aliases:     []string{"ls"},
		Usage:       "Return a list of created objects",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<subcommand>",
		Description: "This is the description",
		Subcommands: GetCommands(ListCmds),
	}
}
