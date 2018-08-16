package trigger

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

// Create trigger command
func Create() cli.Command {
	t := triggersCmd{}
	return cli.Command{
		Name:        "trigger",
		ShortName:   "trig",
		Category:    "MANAGEMENT COMMAND",
		Aliases:     []string{"t", "tr"},
		Usage:       "Create a new trigger",
		Description: "This command creates a new trigger.",
		Before: func(ctx *cli.Context) error {
			var err error
			t.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			t.client = t.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app-name> <function-name> <trigger-name>",
		Action:    t.create,
		Flags:     TriggerFlags,
	}
}

// List trigger command
func List() cli.Command {
	t := triggersCmd{}
	return cli.Command{
		Name:        "triggers",
		ShortName:   "trigs",
		Category:    "MANAGEMENT COMMAND",
		Description: "This command returns a list of all created triggers for an 'app' or for a specific 'function' of an application.",
		Aliases:     []string{"t", "tr"},
		Usage:       "List all triggers",
		Before: func(ctx *cli.Context) error {
			var err error
			t.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			t.client = t.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app-name> [function-name]",
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
		Name:        "trigger",
		ShortName:   "trig",
		Category:    "MANAGEMENT COMMAND",
		Description: "This command updates a created trigger.",
		Aliases:     []string{"t", "tr"},
		Usage:       "Update a trigger",
		Before: func(ctx *cli.Context) error {
			var err error
			t.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			t.client = t.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app-name> <function-name> <trigger-name>",
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
		Name:        "trigger",
		ShortName:   "trig",
		Category:    "MANAGEMENT COMMAND",
		Description: "This command deletes a created trigger.",
		Aliases:     []string{"t", "tr"},
		Usage:       "Delete a trigger",
		Before: func(ctx *cli.Context) error {
			var err error
			t.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			t.client = t.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app-name> <function-name> <trigger-name>",
		Action:    t.delete,
	}
}

// Inspect trigger command
func Inspect() cli.Command {
	t := triggersCmd{}
	return cli.Command{
		Name:        "trigger",
		ShortName:   "trig",
		Category:    "MANAGEMENT COMMAND",
		Aliases:     []string{"t", "tr"},
		Description: "This command gets one of all trigger properties.",
		Usage:       "Retrieve one or all trigger properties",
		Before: func(ctx *cli.Context) error {
			var err error
			t.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			t.client = t.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app-name> <function-name> <trigger-name>",
		Action:    t.inspect,
	}
}
