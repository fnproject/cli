package app

import (
	"github.com/fnproject/cli/common"
	"cmd github.com/fnproject/cli/commands"
	"github.com/urfave/cli"
)

type appCmd common.FnClient

// GetCommand returns the correct application subcommand for the specified command.
func GetCommand(command string, client *appCmd) cli.Command {
	var aCmd cli.Command

	switch command {
	case cmd.CreateCmd:
		aCmd = client.getCreateAppCommand()
	case common.ListCmd:
		aCmd = client.getListAppsCommand()
	case common.DeleteCmd:
		aCmd = client.getDeleteAppCommand()
	case common.InspectCmd:
		aCmd = client.getInspectAppsCommand()
	case common.UpdateCmd:
		aCmd = client.getUpdateAppCommand()
	case common.ConfigCmd:
		aCmd = client.getConfigAppsCommand()
	}

	return aCmd
}

func (client *appCmd) getCreateAppCommand() cli.Command {
	return cli.Command{
		Name:      "app",
		Usage:     "Create a new application",
		ArgsUsage: "<app>",
		Action:    client.createApp,
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "config",
				Usage: "application configuration",
			},
		},
	}
}

func (client *appCmd) getListAppsCommand() cli.Command {
	return cli.Command{
		Name:   "apps",
		Usage:  "List all applications ",
		Action: client.listApps,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "cursor",
				Usage: "pagination cursor",
			},
			cli.Int64Flag{
				Name:  "n",
				Usage: "number of apps to return",
				Value: int64(100),
			},
		},
	}
}

func (client *appCmd) getDeleteAppCommand() cli.Command {
	return cli.Command{
		Name:   "app",
		Usage:  "Delete an application",
		Action: client.deleteApps,
	}
}

func (client *appCmd) getInspectAppsCommand() cli.Command {
	return cli.Command{
		Name:      "apps",
		Usage:     "retrieve one or all apps properties",
		ArgsUsage: "<app> [property.[key]]",
		Action:    client.inspectApps,
	}
}

func (client *appCmd) getUpdateAppCommand() cli.Command {
	return cli.Command{
		Name:      "app",
		Usage:     "update an application",
		ArgsUsage: "<app>",
		Action:    client.updateApps,
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "config,c",
				Usage: "route configuration",
			},
		},
	}
}

func (client *appCmd) getConfigAppsCommand() cli.Command {
	return cli.Command{
		Name:  "apps",
		Usage: "manage apps configs",
		Subcommands: []cli.Command{
			{
				Name:      "set",
				Aliases:   []string{"s"},
				Usage:     "store a configuration key for this application",
				ArgsUsage: "<app> <key> <value>",
				Action:    client.configSetApps,
			},
			{
				Name:      "get",
				Aliases:   []string{"g"},
				Usage:     "inspect configuration key for this application",
				ArgsUsage: "<app> <key>",
				Action:    client.configGetApps,
			},
			{
				Name:      "list",
				Aliases:   []string{"l"},
				Usage:     "list configuration key/value pairs for this application",
				ArgsUsage: "<app>",
				Action:    client.configListApps,
			},
			{
				Name:      "unset",
				Aliases:   []string{"u"},
				Usage:     "remove a configuration key for this application",
				ArgsUsage: "<app> <key>",
				Action:    client.configUnsetApps,
			},
		},
	}
}
