package commands

import (
	"github.com/urfave/cli"
)

// InspectCommand returns inspect cli.command
func InspectCommand() cli.Command {
	return cli.Command{
		Name:        "inspect",
		Aliases:     []string{"i"},
		Usage:       "\tRetrieve properties of an object",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<subcommand>",
		Description: "This is the description",
		Subcommands: GetCommands(InspectCmds),
	}
}
