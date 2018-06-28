package commands

import (
	"github.com/urfave/cli"
)

// GetCommand returns get cli.command
func GetCommand() cli.Command {
	return cli.Command{
		Name:        "get",
		Aliases:     []string{"g"},
		Usage:       "Get an object to retrieve its information",
		Category:    "MANAGEMENT COMMANDS",
		Hidden:      false,
		ArgsUsage:   "<subcommand>",
		Description: "This is the description",
		Subcommands: GetCommands(GetCmds),
	}
}
