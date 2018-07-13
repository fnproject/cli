package commands

import (
	"github.com/urfave/cli"
)

// ConfigCommand returns config cli.command dependant on command parameter
func ConfigCommand(command string) cli.Command {
	var cmds []cli.Command
	switch command {
	case "list":
		cmds = GetCommands(ConfigListCmds)
	case "get":
		cmds = GetCommands(ConfigGetCmds)
	case "configure":
		cmds = GetCommands(ConfigCmds)
	case "unset":
		cmds = GetCommands(ConfigUnsetCmds)
	}

	return cli.Command{
		Name:        "config",
		Usage:       "\tGet configurations for apps and routes",
		Aliases:     []string{"cf"},
		ArgsUsage:   "<subcommand>",
		Description: "This is the description",
		Subcommands: cmds,
	}
}
