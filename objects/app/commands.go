package app

import (
	"encoding/json"
	"fmt"

	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

// Create app command
func Create() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:     "app",
		Usage:    "Create a new application",
		Category: "MANAGEMENT COMMAND",
		Description: "This command creates a new application.\n	Fn supports grouping functions into a set that defines an application (or API), making it easy to organize and deploy.\n	Applications define a namespace to organize functions and can contain configuration values that are shared across all functions in that application.",
		Aliases: []string{"apps", "a"},
		Before: func(c *cli.Context) error {
			providerAdapter, err := client.CurrentProviderAdapter()
			if err != nil {
				return err
			}
			a.clientAdapter = providerAdapter.GetClientAdapter()
			return nil
		},
		ArgsUsage: "<app-name>",
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
			cli.StringFlag{
				Name:  "syslog-url",
				Usage: "Syslog URL to send application logs to",
			},
		},
	}
}

// List apps command
func List() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:        "apps",
		Usage:       "List all created applications",
		Category:    "MANAGEMENT COMMAND",
		Description: "This command provides a list of defined applications.",
		Aliases:     []string{"app", "a"},
		Before: func(c *cli.Context) error {
			providerAdapter, err := client.CurrentProviderAdapter()
			if err != nil {
				return err
			}
			a.clientAdapter = providerAdapter.GetClientAdapter()
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
			cli.StringFlag{
				Name:  "output",
				Usage: "Output format (json)",
			},
		},
	}
}

// Delete app command
func Delete() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:        "app",
		Usage:       "Delete an application",
		Category:    "MANAGEMENT COMMAND",
		Description: "This command deletes a created application.",
		ArgsUsage:   "<app-name>",
		Aliases:     []string{"apps", "a"},
		Before: func(c *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			a.client = provider.APIClientv2()
			return nil
		},
		Action: a.delete,
		BashComplete: func(c *cli.Context) {
			args := c.Args()
			if len(args) == 0 {
				BashCompleteApps(c)
			}
		},
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "force, f",
				Usage: "Forces this delete (you will not be asked if you wish to continue with the delete)",
			},
			cli.BoolFlag{
				Name:  "recursive, r",
				Usage: "Delete this app and all associated resources (can fail part way through execution after deleting some resources without the ability to undo)",
			},
		},
	}
}

// Inspect app command
func Inspect() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:        "app",
		Usage:       "Retrieve one or all apps properties",
		Description: "This command inspects properties of an application.",
		Category:    "MANAGEMENT COMMAND",
		Aliases:     []string{"apps", "a"},
		Before: func(c *cli.Context) error {
			providerAdapter, err := client.CurrentProviderAdapter()
			if err != nil {
				return err
			}
			a.clientAdapter = providerAdapter.GetClientAdapter()
			return nil
		},
		ArgsUsage: "<app-name> [property.[key]]",
		Action:    a.inspect,
		BashComplete: func(c *cli.Context) {
			switch len(c.Args()) {
			case 0:
				BashCompleteApps(c)
			case 1:
				providerAdapter, err := client.CurrentProviderAdapter()
				if err != nil {
					return
				}
				app, err := providerAdapter.GetClientAdapter().GetAppsClient().GetApp(c.Args()[0])
				if err != nil {
					return
				}
				data, err := json.Marshal(app)
				if err != nil {
					return
				}
				var inspect map[string]interface{}
				err = json.Unmarshal(data, &inspect)
				if err != nil {
					return
				}
				for key := range inspect {
					fmt.Println(key)
				}
			}
		},
	}
}

// Update app command
func Update() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:        "app",
		Usage:       "Update an application",
		Category:    "MANAGEMENT COMMAND",
		Description: "This command updates a created application.",
		Aliases:     []string{"apps", "a"},
		Before: func(c *cli.Context) error {
			providerAdapter, err := client.CurrentProviderAdapter()
			if err != nil {
				return err
			}
			a.clientAdapter = providerAdapter.GetClientAdapter()
			return nil
		},
		ArgsUsage: "<app-name>",
		Action:    a.update,
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "config,c",
				Usage: "Application configuration",
			},
			cli.StringSliceFlag{
				Name:  "annotation",
				Usage: "Application annotations",
			},
			cli.StringFlag{
				Name:  "syslog-url",
				Usage: "Syslog URL to send application logs to",
			},
		},
		BashComplete: func(c *cli.Context) {
			args := c.Args()
			if len(args) == 0 {
				BashCompleteApps(c)
			}
		},
	}
}

// SetConfig for function command
func SetConfig() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:        "app",
		Usage:       "Store a configuration key for this application",
		Description: "This command sets configurations for an application.",
		Category:    "MANAGEMENT COMMAND",
		Aliases:     []string{"apps", "a"},
		Before: func(c *cli.Context) error {
			providerAdapter, err := client.CurrentProviderAdapter()
			if err != nil {
				return err
			}
			a.clientAdapter = providerAdapter.GetClientAdapter()
			return nil
		},
		ArgsUsage: "<app-name> <key> <value>",
		Action:    a.setConfig,
		BashComplete: func(c *cli.Context) {
			args := c.Args()
			if len(args) == 0 {
				BashCompleteApps(c)
			}
		},
	}
}

// ListConfig for app command
func ListConfig() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:        "app",
		Usage:       "List configuration key/value pairs for this application",
		Category:    "MANAGEMENT COMMAND",
		Description: "This command lists the configuration of an application.",
		Aliases:     []string{"apps", "a"},
		Before: func(c *cli.Context) error {
			providerAdapter, err := client.CurrentProviderAdapter()
			if err != nil {
				return err
			}
			a.clientAdapter = providerAdapter.GetClientAdapter()
			return nil
		},
		ArgsUsage: "<app-name>",
		Action:    a.listConfig,
		BashComplete: func(c *cli.Context) {
			args := c.Args()
			if len(args) == 0 {
				BashCompleteApps(c)
			}
		},
	}
}

// GetConfig for function command
func GetConfig() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:        "app",
		Usage:       "Inspect configuration key for this application",
		Description: "This command gets the configuration of an application.",
		Category:    "MANAGEMENT COMMAND",
		Aliases:     []string{"apps", "a"},
		Before: func(c *cli.Context) error {
			providerAdapter, err := client.CurrentProviderAdapter()
			if err != nil {
				return err
			}
			a.clientAdapter = providerAdapter.GetClientAdapter()
			return nil
		},
		ArgsUsage: "<app-name> <key>",
		Action:    a.getConfig,
		BashComplete: func(c *cli.Context) {
			switch len(c.Args()) {
			case 0:
				BashCompleteApps(c)
			case 1:
				providerAdapter, err := client.CurrentProviderAdapter()
				if err != nil {
					return
				}
				app, err := providerAdapter.GetClientAdapter().GetAppsClient().GetApp(c.Args()[0])
				if err != nil {
					return
				}
				for key := range app.Config {
					fmt.Println(key)
				}
			}
		},
	}
}

// UnsetConfig for app command
func UnsetConfig() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:        "app",
		Usage:       "Remove a configuration key for this application.",
		Description: "This command removes a configuration for an application.",
		Category:    "MANAGEMENT COMMAND",
		Aliases:     []string{"apps", "a"},
		Before: func(c *cli.Context) error {
			providerAdapter, err := client.CurrentProviderAdapter()
			if err != nil {
				return err
			}
			a.clientAdapter = providerAdapter.GetClientAdapter()
			return nil
		},
		ArgsUsage: "<app-name> <key>",
		Action:    a.unsetConfig,
		BashComplete: func(c *cli.Context) {
			switch len(c.Args()) {
			case 0:
				BashCompleteApps(c)
			case 1:
				providerAdapter, err := client.CurrentProviderAdapter()
				if err != nil {
					return
				}
				app, err := providerAdapter.GetClientAdapter().GetAppsClient().GetApp(c.Args()[0])
				if err != nil {
					return
				}
				for key := range app.Config {
					fmt.Println(key)
				}
			}
		},
	}
}
