package app

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

func Create() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:     "app",
		Usage:    "Create a new application",
		Category: "MANAGEMENT COMMAND",
		Description: "This command creates a new application.\n		Fn supports grouping functions into a set that defines an application (or API), making it easy to organize and deploy.\n 		Applications define a namespace to organize functions and can contain configuration values that are shared across all functions in that application.",
		Aliases: []string{"apps", "a"},
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			a.client = provider.APIClient()
			return nil
		},
		ArgsUsage: "<app_name>",
		Action:    a.create,
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "config",
				Usage: "Application configuration",
			},
			cli.StringSliceFlag{
				Name:  "annotation",
				Usage: "Application annotations",
			},
		},
	}
}

func List() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:     "apps",
		Usage:    "List all created applications",
		Category: "MANAGEMENT COMMANDS",
		Aliases:  []string{"app", "a"},
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			a.client = provider.APIClient()
			return nil
		},
		Action: a.list,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "cursor",
				Usage: "Pagination cursor",
			},
			cli.Int64Flag{
				Name:  "n",
				Usage: "Number of apps to return",
				Value: int64(100),
			},
		},
	}
}

func Delete() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:        "apps",
		Usage:       "Delete an application",
		Category:    "MANAGEMENT COMMANDS",
		Description: "This command deletes a created application.",
		Aliases:     []string{"apps", "a"},
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			a.client = provider.APIClient()
			return nil
		},
		Action: a.delete,
	}
}

func Inspect() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:     "apps",
		Usage:    "Retrieve one or all apps properties",
		Category: "MANAGEMENT COMMANDS",
		Aliases:  []string{"apps", "a"},
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			a.client = provider.APIClient()
			return nil
		},
		ArgsUsage: "<app_name> [property.[key]]",
		Action:    a.inspect,
	}
}

func Update() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:     "apps",
		Usage:    "Update an application",
		Category: "MANAGEMENT COMMANDS",
		Aliases:  []string{"apps", "a"},
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			a.client = provider.APIClient()
			return nil
		},
		ArgsUsage: "<app_name>",
		Action:    a.update,
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "config,c",
				Usage: "Route configuration",
			},
			cli.StringSliceFlag{
				Name:  "annotation",
				Usage: "Application annotations",
			},
		},
	}
}

func SetConfig() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:    "apps",
		Usage:   "Store a configuration key for this application",
		Aliases: []string{"apps", "a"},
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			a.client = provider.APIClient()
			return nil
		},
		ArgsUsage: "<app_name> <key> <value>",
		Action:    a.setConfig,
	}
}

func ListConfig() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:    "apps",
		Usage:   "List configuration key/value pairs for this application",
		Aliases: []string{"apps", "a"},
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			a.client = provider.APIClient()
			return nil
		},
		ArgsUsage: "<app_name>",
		Action:    a.listConfig,
	}
}

func GetConfig() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:    "apps",
		Usage:   "Inspect configuration key for this application",
		Aliases: []string{"apps", "a"},
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			a.client = provider.APIClient()
			return nil
		},
		ArgsUsage: "<app_name> <key>",
		Action:    a.getConfig,
	}
}

func UnsetConfig() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:    "apps",
		Usage:   "Remove a configuration key for this application",
		Aliases: []string{"apps", "a"},
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			a.client = provider.APIClient()
			return nil
		},
		ArgsUsage: "<app_name> <key>",
		Action:    a.unsetConfig,
	}
}
