package app

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

func Create() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:      "apps",
		ShortName: "app",
		Usage:     "create a new application",
		Aliases:   []string{"a"},
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			a.client = provider.APIClient()
			return nil
		},
		ArgsUsage: "<app>",
		Action:    a.create,
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "config",
				Usage: "application configuration",
			},
			cli.StringSliceFlag{
				Name:  "annotation",
				Usage: "application annotations",
			},
		},
	}
}

func List() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:      "apps",
		ShortName: "app",
		Usage:     "list all applications",
		Aliases:   []string{"a"},
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

func Delete() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:      "apps",
		ShortName: "app",
		Usage:     "delete an application",
		Aliases:   []string{"a"},
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
		Name:      "apps",
		ShortName: "app",
		Usage:     "retrieve one or all apps properties",
		Aliases:   []string{"a"},
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			a.client = provider.APIClient()
			return nil
		},
		ArgsUsage: "<app> [property.[key]]",
		Action:    a.inspect,
	}
}

func Update() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:      "apps",
		ShortName: "app",
		Usage:     "update an application",
		Aliases:   []string{"a"},
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			a.client = provider.APIClient()
			return nil
		},
		ArgsUsage: "<app>",
		Action:    a.update,
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "config,c",
				Usage: "route configuration",
			},
			cli.StringSliceFlag{
				Name:  "annotation",
				Usage: "application annotations",
			},
		},
	}
}

func SetConfig() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:      "apps",
		ShortName: "app",
		Usage:     "store a configuration key for this application",
		Aliases:   []string{"a"},
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			a.client = provider.APIClient()
			return nil
		},
		ArgsUsage: "<app> <key> <value>",
		Action:    a.setConfig,
	}
}

func ListConfig() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:      "apps",
		ShortName: "app",
		Usage:     "list configuration key/value pairs for this application",
		Aliases:   []string{"a"},
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			a.client = provider.APIClient()
			return nil
		},
		ArgsUsage: "<app>",
		Action:    a.listConfig,
	}
}

func GetConfig() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:      "apps",
		ShortName: "app",
		Usage:     "inspect configuration key for this application",
		Aliases:   []string{"a"},
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			a.client = provider.APIClient()
			return nil
		},
		ArgsUsage: "<app> <key>",
		Action:    a.getConfig,
	}
}

func UnsetConfig() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:      "apps",
		ShortName: "app",
		Usage:     "remove a configuration key for this application",
		Aliases:   []string{"a"},
		Before: func(cxt *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			a.client = provider.APIClient()
			return nil
		},
		ArgsUsage: "<app> <key>",
		Action:    a.unsetConfig,
	}
}
