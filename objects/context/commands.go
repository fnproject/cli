/*
 * Copyright (c) 2019, 2020 Oracle and/or its affiliates. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package context

import (
	"fmt"

	"github.com/urfave/cli"
)

// Create context command
func Create() cli.Command {
	return cli.Command{
		Name:        "context",
		Usage:       "Create a new context",
		Aliases:     []string{"ctx"},
		ArgsUsage:   "<context-name>",
		Category:    "MANAGEMENT COMMAND",
		Description: "This command creates a new context for a created application.",
		Action:      createCtx,
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

// List contexts command
func List() cli.Command {
	return cli.Command{
		Name:        "contexts",
		Usage:       "List contexts",
		Aliases:     []string{"context", "ctx"},
		Category:    "MANAGEMENT COMMAND",
		Description: "This command returns a list of contexts.",
		Action:      listCtx,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "output",
				Usage: "Output format (json)",
				Value: "",
			},
		},
	}
}

// Delete context command
func Delete() cli.Command {
	return cli.Command{
		Name:        "context",
		Usage:       "Delete a context",
		Aliases:     []string{"ctx"},
		ArgsUsage:   "<context-name>",
		Description: "This command deletes a context.",
		Category:    "MANAGEMENT COMMAND",
		Action:      deleteCtx,
		BashComplete: func(c *cli.Context) {
			switch len(c.Args()) {
			case 0:
				contexts, err := getAvailableContexts()
				if err != nil {
					return
				}
				for _, c := range contexts {
					if c.Name != "default" {
						fmt.Println(c.Name)
					}
				}
			}
		},
	}
}

// Inspect context command
func Inspect() cli.Command {
	return cli.Command{
		Name:     "context",
		Usage:    "Inspect the contents of a context, if no context is specified the current-context will be used.",
		Aliases:  []string{"ctx"},
		Category: "MANAGEMENT COMMAND",
		Action:   inspectCtx,
		BashComplete: func(c *cli.Context) {
			switch len(c.Args()) {
			case 0:
				contexts, err := getAvailableContexts()
				if err != nil {
					return
				}
				for _, c := range contexts {
					fmt.Println(c.Name)
				}
			}
		},
	}
}

// Update context command
func Update() cli.Command {
	ctxMap := ContextMap{}
	return cli.Command{
		Name:        "context",
		Usage:       "Update context files",
		Aliases:     []string{"ctx"},
		ArgsUsage:   "<key> [value]",
		Category:    "MANAGEMENT COMMAND",
		Description: "This command updates the current context file.",
		Action:      ctxMap.updateCtx,
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "delete",
				Usage: "Delete key=value pair from context file.",
			},
		},
	}
}

// Use context command
func Use() cli.Command {
	return cli.Command{
		Name:        "context",
		Usage:       "Use context for future invocations",
		Aliases:     []string{"ctx"},
		ArgsUsage:   "<context-name>",
		Category:    "MANAGEMENT COMMAND",
		Description: "This command uses context for future invocations.",
		Action:      useCtx,
		BashComplete: func(c *cli.Context) {
			switch len(c.Args()) {
			case 0:
				contexts, err := getAvailableContexts()
				if err != nil {
					return
				}
				for _, c := range contexts {
					fmt.Println(c.Name)
				}
			}
		},
	}
}

// Unset context command
func Unset() cli.Command {
	return cli.Command{
		Name:        "context",
		Usage:       "Unset current-context",
		Aliases:     []string{"ctx"},
		Category:    "MANAGEMENT COMMAND",
		Description: "This command unsets the current context in use.",
		Action:      unsetCtx,
	}
}
