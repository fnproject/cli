package app

import (
	"github.com/fnproject/cli/common"
	"github.com/urfave/cli"
)

type app common.FnClient

func createAppCmd(client *common.FnClient) (appCmd app) {
	appCmd.Client = client.Client
	return
}

// GetCommand returns the correct application subcommand for the specified command.
func GetCommand(command string, client *common.FnClient) cli.Command {
	var aCmd cli.Command
	appCmd := createAppCmd(client)
	switch command {
	case common.CreateCmd:
		aCmd = appCmd.getCreateAppCommand()
	case common.ListCmd:
		aCmd = appCmd.getListAppsCommand()
	case common.DeleteCmd:
		aCmd = appCmd.getDeleteAppCommand()
	case common.InspectCmd:
		aCmd = appCmd.getInspectAppsCommand()
	case common.UpdateCmd:
		aCmd = appCmd.getUpdateAppCommand()
	case common.ConfigCmd:
		aCmd = appCmd.getConfigAppsCommand()
	}

	return aCmd
}

func (appCmd *app) getCreateAppCommand() cli.Command {
	return cli.Command{
		Name:      "app",
		Usage:     "Create a new application",
		ArgsUsage: "<app>",
		Action:    appCmd.createApp,
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "config",
				Usage: "application configuration",
			},
		},
	}
}

func (appCmd *app) getListAppsCommand() cli.Command {
	return cli.Command{
		Name:   "apps",
		Usage:  "List all applications ",
		Action: appCmd.listApps,
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

func (appCmd *app) getDeleteAppCommand() cli.Command {
	return cli.Command{
		Name:   "app",
		Usage:  "Delete an application",
		Action: appCmd.deleteApps,
	}
}

func (appCmd *app) getInspectAppsCommand() cli.Command {
	return cli.Command{
		Name:      "apps",
		Usage:     "retrieve one or all apps properties",
		ArgsUsage: "<app> [property.[key]]",
		Action:    appCmd.inspectApps,
	}
}

func (appCmd *app) getUpdateAppCommand() cli.Command {
	return cli.Command{
		Name:      "app",
		Usage:     "update an application",
		ArgsUsage: "<app>",
		Action:    appCmd.updateApps,
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "config,c",
				Usage: "route configuration",
			},
		},
	}
}

func (appCmd *app) getConfigAppsCommand() cli.Command {
	return cli.Command{
		Name:  "apps",
		Usage: "manage apps configs",
		Subcommands: []cli.Command{
			{
				Name:      "set",
				Aliases:   []string{"s"},
				Usage:     "store a configuration key for this application",
				ArgsUsage: "<app> <key> <value>",
				Action:    appCmd.configSetApps,
			},
			{
				Name:      "get",
				Aliases:   []string{"g"},
				Usage:     "inspect configuration key for this application",
				ArgsUsage: "<app> <key>",
				Action:    appCmd.configGetApps,
			},
			{
				Name:      "list",
				Aliases:   []string{"l"},
				Usage:     "list configuration key/value pairs for this application",
				ArgsUsage: "<app>",
				Action:    appCmd.configListApps,
			},
			{
				Name:      "unset",
				Aliases:   []string{"u"},
				Usage:     "remove a configuration key for this application",
				ArgsUsage: "<app> <key>",
				Action:    appCmd.configUnsetApps,
			},
		},
	}
}
