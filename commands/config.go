package commands

import (
	"github.com/urfave/cli"
)

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
		Usage:       "get configurations for apps and routes",
		Aliases:     []string{"cf"},
		ArgsUsage:   "<object>",
		Subcommands: cmds,
	}
}
