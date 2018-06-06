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
		Usage:     "create a route in an application",
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
		ArgsUsage: "<app> </path> <image>",
		Action:    r.create,
		Flags:     RouteFlags,
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
		Before: func(c *cli.Context) error {
			var err error
			r.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			r.client = r.provider.APIClient()
			return nil
		},
		ArgsUsage: "<app>",
		Action:    r.list,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "cursor",
				Usage: "pagination cursor",
			},
			cli.Int64Flag{
				Name:  "n",
				Usage: "number of routes to return",
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
		Usage:     "delete a route from an application `app`",
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
		ArgsUsage: "<app> </path>",
		Action:    r.delete,
	}
}

// Inspect route command
func Inspect() cli.Command {
	r := routesCmd{}
	return cli.Command{
		Name:      "routes",
		ShortName: "route",
		Usage:     "retrieve one or all routes properties",
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
		ArgsUsage: "<app> </path> [property.[key]]",
		Action:    r.inspect,
	}
}

// Update route command
func Update() cli.Command {
	r := routesCmd{}
	return cli.Command{
		Name:      "routes",
		ShortName: "route",
		Usage:     "update a route in an `app`",
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
		ArgsUsage: "<app> </path>",
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
		Usage:     "inspect configuration key for this route",
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
		ArgsUsage: "<app> </path> <key>",
		Action:    r.getConfig,
	}
}

// SetConfig for route command
func SetConfig() cli.Command {
	r := routesCmd{}
	return cli.Command{
		Name:      "routes",
		ShortName: "route",
		Usage:     "store a configuration key for this route",
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
		ArgsUsage: "<app> </path> <key> <value>",
		Action:    r.setConfig,
	}
}

// ListConfig for route command
func ListConfig() cli.Command {
	r := routesCmd{}
	return cli.Command{
		Name:      "routes",
		ShortName: "route",
		Usage:     "list configuration key/value pairs for this route",
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
		ArgsUsage: "<app> </path>",
		Action:    r.listConfig,
	}
}

// UnsetConfig for route command
func UnsetConfig() cli.Command {
	r := routesCmd{}
	return cli.Command{
		Name:      "routes",
		ShortName: "route",
		Usage:     "remove a configuration key for this route",
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
		ArgsUsage: "<app> </path> <key>",
		Action:    r.unsetConfig,
	}
}
