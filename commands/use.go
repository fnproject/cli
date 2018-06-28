package commands

import (
	"github.com/urfave/cli"
)

// UseCommand returns use cli.command
func UseCommand() cli.Command {
	return cli.Command{
		Name:        "use",
		Aliases:     []string{"u"},
		Usage:       "Select context for further commands",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<subcommand>",
		Description: "This is the description",
		Subcommands: GetCommands(UseCmds),
	}
}
