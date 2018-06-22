package trigger

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

// Create trigger command
func Create() cli.Command {
	t := triggersCmd{}
	return cli.Command{
		Name:      "trigger",
		ShortName: "trig",
		Aliases:   []string{"t", "tr"},
		Usage:     "create a new trigger",
		Before: func(ctx *cli.Context) error {
			var err error
			t.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			t.client = t.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app> <function> <trigger>",
		Action:    t.create,
		Flags:     TriggerFlags,
	}
}

// List trigger command
func List() cli.Command {
	t := triggersCmd{}
	return cli.Command{
		Name:      "triggers",
		ShortName: "trigs",
		Aliases:   []string{"t", "tr"},
		Usage:     "list all triggers",
		Before: func(ctx *cli.Context) error {
			var err error
			t.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			t.client = t.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app> <function>",
		Action:    t.list,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "cursor",
				Usage: "pagination cursor",
			},
			cli.Int64Flag{
				Name:  "n",
				Usage: "number of triggers to return",
				Value: int64(100),
			},
		},
	}
}

// Update trigger command
func Update() cli.Command {
	t := triggersCmd{}
	return cli.Command{
		Name:      "trigger",
		ShortName: "trig",
		Aliases:   []string{"t", "tr"},
		Usage:     "update a trigger",
		Before: func(ctx *cli.Context) error {
			var err error
			t.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			t.client = t.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app> <function> <trigger>",
		Action:    t.update,
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "annotation",
				Usage: "trigger annotations",
			},
		},
	}
}

// Delete trigger command
func Delete() cli.Command {
	t := triggersCmd{}
	return cli.Command{
		Name:      "trigger",
		ShortName: "trig",
		Aliases:   []string{"t", "tr"},
		Usage:     "delete a trigger",
		Before: func(ctx *cli.Context) error {
			var err error
			t.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			t.client = t.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app> <function> <trigger>",
		Action:    t.delete,
	}
}

// Inspect trigger command
func Inspect() cli.Command {
	t := triggersCmd{}
	return cli.Command{
		Name:      "trigger",
		ShortName: "trig",
		Aliases:   []string{"t", "tr"},
		Usage:     "inspect a trigger",
		Before: func(ctx *cli.Context) error {
			var err error
			t.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			t.client = t.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app> <function> <trigger>",
		Action:    t.inspect,
	}
}
