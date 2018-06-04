package commands

import (
	"github.com/urfave/cli"
)

// ConfigureCommand returns configure cli.command
func ConfigureCommand() cli.Command {
	return cli.Command{
		Name:        "config",
		Aliases:     []string{"cf"},
		Usage:       "set configuration for an object",
		Category:    "DEVELOPMENT COMMANDS",
		ArgsUsage:   "<command>",
		Subcommands: GetCommands(ConfigCmds),
	}
}
