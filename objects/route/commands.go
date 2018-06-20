package route

import (
	client "github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

// Create route command
func Create() cli.Command {
	r := routesCmd{}
	return cli.Command{
		Name:      "routes",
		ShortName: "route",
		Usage:     "Create a route in an application",
		Category:  "MANAGEMENT COMMAND",
		Aliases:   []string{"r"},
		Before: func(c *cli.Context) error {
			var err error
			r.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			r.client = r.provider.APIClient()
			return nil
		},
		ArgsUsage:   "<app_name> </path> <image>",
		Description: "This command creates a new route for a created application.",
		Action:      r.create,
		Flags:       RouteFlags,
	}
}

// List routes command
func List() cli.Command {
	r := routesCmd{}
	return cli.Command{
		Name:      "routes",
		ShortName: "route",
		Usage:     "list routes for `app`",
		Aliases:   []string{"r"},
		Category:  "MANAGEMENT COMMAND",
		Before: func(c *cli.Context) error {
			var err error
			r.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			r.client = r.provider.APIClient()
			return nil
		},
		ArgsUsage: "<app_name>",
		Action:    r.list,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "cursor",
				Usage: "Pagination cursor",
			},
			cli.Int64Flag{
				Name:  "n",
				Usage: "Number of routes to return",
				Value: int64(100),
			},
		},
	}
}

// Delete route command
func Delete() cli.Command {
	r := routesCmd{}
	return cli.Command{
		Name:      "routes",
		ShortName: "route",
		Usage:     "Delete a route from an application `app`",
		Category:  "MANAGEMENT COMMAND",
		Aliases:   []string{"r"},
		Before: func(c *cli.Context) error {
			var err error
			r.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			r.client = r.provider.APIClient()
			return nil
		},
		ArgsUsage: "<app_name> </path>",
		Action:    r.delete,
	}
}

// Inspect route command
func Inspect() cli.Command {
	r := routesCmd{}
	return cli.Command{
		Name:      "routes",
		ShortName: "route",
		Usage:     "Retrieve one or all routes properties",
		Aliases:   []string{"r"},
		Category:  "MANAGEMENT COMMAND",
		Before: func(c *cli.Context) error {
			var err error
			r.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			r.client = r.provider.APIClient()
			return nil
		},
		ArgsUsage: "<app_name> </path> [property.[key]]",
		Action:    r.inspect,
	}
}

// Update route command
func Update() cli.Command {
	r := routesCmd{}
	return cli.Command{
		Name:      "routes",
		ShortName: "route",
		Usage:     "Update a route in an `app`",
		Aliases:   []string{"r"},
		Category:  "MANAGEMENT COMMAND",
		Before: func(c *cli.Context) error {
			var err error
			r.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			r.client = r.provider.APIClient()
			return nil
		},
		ArgsUsage: "<app_name> </path>",
		Action:    r.update,
		Flags:     updateRouteFlags,
	}
}

// GetConfig for route command
func GetConfig() cli.Command {
	r := routesCmd{}
	return cli.Command{
		Name:      "routes",
		ShortName: "route",
		Usage:     "Inspect configuration key for this route",
		Aliases:   []string{"r"},
		Category:  "MANAGEMENT COMMAND",
		Before: func(c *cli.Context) error {
			var err error
			r.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			r.client = r.provider.APIClient()
			return nil
		},
		ArgsUsage: "<app_name> </path> <key>",
		Action:    r.getConfig,
	}
}

// SetConfig for route command
func SetConfig() cli.Command {
	r := routesCmd{}
	return cli.Command{
		Name:      "routes",
		ShortName: "route",
		Usage:     "Store a configuration key for this route",
		Category:  "MANAGEMENT COMMAND",
		Aliases:   []string{"r"},
		Before: func(c *cli.Context) error {
			var err error
			r.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			r.client = r.provider.APIClient()
			return nil
		},
		ArgsUsage: "<app_name> </path> <key> <value>",
		Action:    r.setConfig,
	}
}

// ListConfig for route command
func ListConfig() cli.Command {
	r := routesCmd{}
	return cli.Command{
		Name:      "routes",
		ShortName: "route",
		Usage:     "List configuration key/value pairs for this route",
		Aliases:   []string{"r"},
		Category:  "MANAGEMENT COMMAND",
		Before: func(c *cli.Context) error {
			var err error
			r.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			r.client = r.provider.APIClient()
			return nil
		},
		ArgsUsage: "<app_name> </path>",
		Action:    r.listConfig,
	}
}

// UnsetConfig for route command
func UnsetConfig() cli.Command {
	r := routesCmd{}
	return cli.Command{
		Name:      "routes",
		ShortName: "route",
		Usage:     "Remove a configuration key for this route",
		Aliases:   []string{"r"},
		Category:  "MANAGEMENT COMMAND",
		Before: func(c *cli.Context) error {
			var err error
			r.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			r.client = r.provider.APIClient()
			return nil
		},
		ArgsUsage: "<app_name> </path> <key>",
		Action:    r.unsetConfig,
	}
}
