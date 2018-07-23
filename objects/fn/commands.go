package fn

import (
	client "github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

// Create function command
func Create() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:      "function",
		ShortName: "func",
		Aliases:   []string{"f", "fn"},
		Usage:     "Create a function within an application",
		Before: func(c *cli.Context) error {
			var err error
			f.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			f.client = f.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app-name> <function-name> <image>",
		Action:    f.create,
		Flags:     FnFlags,
	}
}

// List functions command
func List() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:      "functions",
		ShortName: "funcs",
		Aliases:   []string{"fs", "fns"},
		Usage:     "List functions for an application",
		Before: func(c *cli.Context) error {
			var err error
			f.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			f.client = f.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app-name>",
		Action:    f.list,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "cursor",
				Usage: "pagination cursor",
			},
			cli.Int64Flag{
				Name:  "n",
				Usage: "number of functions to return",
				Value: int64(100),
			},
		},
	}
}

// Delete function command
func Delete() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:      "function",
		ShortName: "func",
		Aliases:   []string{"f", "fn"},
		Usage:     "Delete a function from an application",
		Before: func(c *cli.Context) error {
			var err error
			f.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			f.client = f.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app-name> <function-name>",
		Action:    f.delete,
	}
}

// Inspect function command
func Inspect() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:      "function",
		ShortName: "func",
		Aliases:   []string{"f", "fn"},
		Usage:     "Retrieve one or all properties for a function",
		Before: func(c *cli.Context) error {
			var err error
			f.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			f.client = f.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app-name> <function-name> [property.[key]]",
		Action:    f.inspect,
	}
}

// Update function command
func Update() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:      "function",
		ShortName: "func",
		Aliases:   []string{"f", "fn"},
		Usage:     "Update a function in application",
		Before: func(c *cli.Context) error {
			var err error
			f.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			f.client = f.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app-name>",
		Action:    f.update,
		Flags:     updateFnFlags,
	}
}

// GetConfig for function command
func GetConfig() cli.Command {
	r := fnsCmd{}
	return cli.Command{
		Name:      "function",
		ShortName: "func",
		Aliases:   []string{"f", "fn"},
		Usage:     "Inspect configuration key for this function",
		Before: func(c *cli.Context) error {
			var err error
			r.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			r.client = r.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app-name> <funct> <key>",
		Action:    r.getConfig,
	}
}

// SetConfig for function command
func SetConfig() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:      "function",
		ShortName: "func",
		Aliases:   []string{"f", "fn"},
		Usage:     "Store a configuration key for this function",
		Before: func(c *cli.Context) error {
			var err error
			f.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			f.client = f.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app-name> <function-name> <key> <value>",
		Action:    f.setConfig,
	}
}

// ListConfig for function command
func ListConfig() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:      "function",
		ShortName: "func",
		Aliases:   []string{"f", "fn"},
		Usage:     "List configuration key/value pairs for this function",
		Before: func(c *cli.Context) error {
			var err error
			f.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			f.client = f.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app-name> <function-name>",
		Action:    f.listConfig,
	}
}

// UnsetConfig for function command
func UnsetConfig() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:      "function",
		ShortName: "func",
		Aliases:   []string{"f", "fn"},
		Usage:     "Remove a configuration key for this function",
		Before: func(c *cli.Context) error {
			var err error
			f.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			f.client = f.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app-name> </path> <key>",
		Action:    f.unsetConfig,
	}
}
