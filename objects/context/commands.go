package context

import "github.com/urfave/cli"

func Create() cli.Command {
	return cli.Command{
		Name:        "context",
		Usage:       "Create a new context",
		Aliases:     []string{"ctx"},
		ArgsUsage:   "<context>",
		Category:    "MANAGEMENT COMMAND",
		Description: "This command creates a new context for a created application.",
		Action:      create,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "provider",
				Usage: "Context provider",
			},
			cli.StringFlag{
				Name:  "api-url",
				Usage: "Context api url",
			},
			cli.StringFlag{
				Name:  "registry",
				Usage: "Context registry",
			},
		},
	}
}

func List() cli.Command {
	return cli.Command{
		Name:     "contexts",
		Usage:    "List contexts",
		Aliases:  []string{"context", "ctx"},
		Category: "MANAGEMENT COMMAND",
		Action:   list,
	}
}

func Delete() cli.Command {
	return cli.Command{
		Name:      "context",
		Usage:     "Delete a context",
		Aliases:   []string{"ctx"},
		ArgsUsage: "<context>",
		Category:  "MANAGEMENT COMMAND",
		Action:    delete,
	}
}

func Inspect() cli.Command {
	return cli.Command{
		Name:     "context",
		Usage:    "Inspect the contents of a context, if no context is specified the current-context will be used.",
		Aliases:  []string{"ctx"},
		Category: "MANAGEMENT COMMAND",
		Action:   inspect,
	}
}

func Update() cli.Command {
	ctxMap := ContextMap{}
	return cli.Command{
		Name:      "context",
		Usage:     "Update context files",
		Aliases:   []string{"ctx"},
		ArgsUsage: "<key> <value>",
		Category:  "MANAGEMENT COMMAND",
		Action:    ctxMap.update,
	}
}

func Use() cli.Command {
	return cli.Command{
		Name:      "context",
		Usage:     "Use context for future invocations",
		Aliases:   []string{"ctx"},
		ArgsUsage: "<context>",
		Category:  "MANAGEMENT COMMAND",
		Action:    use,
	}
}

func Unset() cli.Command {
	return cli.Command{
		Name:     "context",
		Usage:    "Unset current-context",
		Aliases:  []string{"ctx"},
		Category: "MANAGEMENT COMMAND",
		Action:   unset,
	}
}
