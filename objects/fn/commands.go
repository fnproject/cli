package fn

import (
	client "github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

// Create route command
func Create() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:      "functions",
		ShortName: "funcs",
		Usage:     "create a function in an application",
		Aliases:   []string{"f","fn"},
		Before: func(c *cli.Context) error {
			var err error
			f.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			f.client = f.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app> <fnname> <image>",
		Action:    f.create,
		Flags:     FnFlags,
	}
}

// List routes command
func List() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:      "functions",
		ShortName: "function",
		Aliases:   []string{"f","fn","fns","func","funcs"},
		Usage:     "list functions for `app`",
		Before: func(c *cli.Context) error {
			var err error
			f.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			f.client = f.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app>",
		Action:    f.list,
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
	f := fnsCmd{}
	return cli.Command{
		Name:      "functions",
		ShortName: "function",
		Aliases:   []string{"f","fn","fns","func","funcs"},

		Usage:     "delete a function from an application `app`",
		Before: func(c *cli.Context) error {
			var err error
			f.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			f.client = f.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app> </path>",
		Action:    f.delete,
	}
}

// Inspect route command
func Inspect() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:      "functions",
		ShortName: "function",
		Aliases:   []string{"f","fn","fns","func","funcs"},
		Usage:     "retrieve one or all routes properties",
		Before: func(c *cli.Context) error {
			var err error
			f.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			f.client = f.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app> <fn> [property.[key]]",
		Action:    f.inspect,
	}
}

// Update route command
func Update() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:      "functions",
		ShortName: "function",
		Aliases:   []string{"f","fn","fns","func","funcs"},
		Usage:     "update a function in an `app`",
		Before: func(c *cli.Context) error {
			var err error
			f.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			f.client = f.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app> </path>",
		Action:    f.update,
		Flags:     updateFnFlags,
	}
}


// GetConfig for route command
func GetConfig() cli.Command {
	r := fnsCmd{}
	return cli.Command{
		Name:      "functions",
		ShortName: "function",
		Aliases:   []string{"f","fn","fns","func","funcs"},
		Usage:     "inspect configuration key for this function",
		Before: func(c *cli.Context) error {
			var err error
			r.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			r.client = r.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app> <funct> <key>",
		Action:    r.getConfig,
	}
}

// SetConfig for route command
func SetConfig() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:      "functions",
		ShortName: "function",
		Aliases:   []string{"f","fn","fns","func","funcs"},
		Usage:     "store a configuration key for this function",

		Before: func(c *cli.Context) error {
			var err error
			f.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			f.client = f.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app> <fn> <key> <value>",
		Action:    f.setConfig,
	}
}

// ListConfig for route command
func ListConfig() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:      "functions",
		ShortName: "function",
		Aliases:   []string{"f","fn","fns","func","funcs"},

		Usage:     "list configuration key/value pairs for this function",
		Before: func(c *cli.Context) error {
			var err error
			f.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			f.client = f.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app> <fnname>",
		Action:    f.listConfig,
	}
}

// UnsetConfig for route command
func UnsetConfig() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:      "functions",
		ShortName: "function",
		Aliases:   []string{"f","fn","fns","func","funcs"},
		Usage:     "remove a configuration key for this route",
		Before: func(c *cli.Context) error {
			var err error
			f.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			f.client = f.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app> </path> <key>",
		Action:    f.unsetConfig,
	}
}
