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
		Usage:     "create a function in an application",
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

// List functions command
func List() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:      "functions",
		ShortName: "funcs",
		Aliases:   []string{"f", "fn"},
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

// Inspect function command
func Inspect() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:      "function",
		ShortName: "func",
		Aliases:   []string{"f", "fn"},
		Usage:     "retrieve one or all functions properties",
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

// Update function command
func Update() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:      "function",
		ShortName: "func",
		Aliases:   []string{"f", "fn"},
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

// GetConfig for function command
func GetConfig() cli.Command {
	r := fnsCmd{}
	return cli.Command{
		Name:      "function",
		ShortName: "func",
		Aliases:   []string{"f", "fn"},
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

// SetConfig for function command
func SetConfig() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:      "function",
		ShortName: "func",
		Aliases:   []string{"f", "fn"},
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

// ListConfig for function command
func ListConfig() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:      "function",
		ShortName: "func",
		Aliases:   []string{"f", "fn"},
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

// UnsetConfig for function command
func UnsetConfig() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:      "function",
		ShortName: "func",
		Aliases:   []string{"f", "fn"},
		Usage:     "remove a configuration key for this function",
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
