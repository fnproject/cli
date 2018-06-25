package trigger

import (
	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

func Create() cli.Command {
	t := triggersCmd{}
	return cli.Command{
		Name:      "triggers",
		ShortName: "trigger",
		Usage:     "create a new trigger",
		Aliases:   []string{"t", "tr", "trig"},
		Before: func(ctx *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			t.client = provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app> <function> <trigger>",
		Action:    t.create,
		Flags:     TriggerFlags,
	}
}

func List() cli.Command {
	t := triggersCmd{}
	return cli.Command{
		Name:      "triggers",
		ShortName: "trigger",
		Usage:     "list all triggers",
		Aliases:   []string{"t", "tr", "trig"},
		Before: func(ctx *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			t.client = provider.APIClientv2()
			return nil
		},
		Action: t.list,
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

func update() cli.Command {
	t := triggersCmd{}
	return cli.Command{
		Name:      "triggers",
		ShortName: "trigger",
		Usage:     "update a trigger",
		Aliases:   []string{"t", "tr", "trig"},
		Before: func(ctx *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			t.client = provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<trigger> ",
		Action:    t.update,
	}
}

func delete() cli.Command {
	t := triggersCmd{}
	return cli.Command{
		Name:      "triggers",
		ShortName: "trigger",
		Usage:     "delete a trigger",
		Aliases:   []string{"t", "tr", "trig"},
		Before: func(ctx *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			t.client = provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<trigger>",
		Action:    t.delete,
	}
}
