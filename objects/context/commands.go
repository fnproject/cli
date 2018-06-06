package context

import "github.com/urfave/cli"

func Create() cli.Command {
	return cli.Command{
		Name:      "context",
		Usage:     "create a new context",
		Aliases:   []string{"ctx"},
		ArgsUsage: "<context>",
		Action:    create,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "provider",
				Usage: "context provider",
			},
			cli.StringFlag{
				Name:  "api-url",
				Usage: "context api url",
			},
			cli.StringFlag{
				Name:  "registry",
				Usage: "context registry",
			},
		},
	}
}

func List() cli.Command {
	return cli.Command{
		Name:    "context",
		Usage:   "list contexts",
		Aliases: []string{"ctx"},
		Action:  list,
	}
}

func Delete() cli.Command {
	return cli.Command{
		Name:      "context",
		Usage:     "delete a context",
		Aliases:   []string{"ctx"},
		ArgsUsage: "<context>",
		Action:    delete,
	}
}

func Update() cli.Command {
	ctxMap := ContextMap{}
	return cli.Command{
		Name:      "context",
		Usage:     "update context files",
		Aliases:   []string{"ctx"},
		ArgsUsage: "<key> <value>",
		Action:    ctxMap.update,
	}
}

func Use() cli.Command {
	return cli.Command{
		Name:      "context",
		Usage:     "use context for future invocations",
		Aliases:   []string{"ctx"},
		ArgsUsage: "<context>",
		Action:    use,
	}
}

func Unset() cli.Command {
	return cli.Command{
		Name:    "context",
		Usage:   "unset current-context",
		Aliases: []string{"ctx"},
		Action:  unset,
	}
}
