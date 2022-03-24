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

package trigger

import (
	"encoding/json"
	"fmt"

	"github.com/fnproject/cli/client"
	"github.com/fnproject/cli/objects/app"
	"github.com/fnproject/cli/objects/fn"
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
		BashComplete: func(c *cli.Context) {
			switch len(c.Args()) {
			case 0:
				app.BashCompleteApps(c)
			case 1:
				fn.BashCompleteFns(c)
			}
		},
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
		BashComplete: func(c *cli.Context) {
			switch len(c.Args()) {
			case 0:
				app.BashCompleteApps(c)
			case 1:
				fn.BashCompleteFns(c)
			}
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
		BashComplete: func(c *cli.Context) {
			switch len(c.Args()) {
			case 0:
				app.BashCompleteApps(c)
			case 1:
				fn.BashCompleteFns(c)
			case 2:
				BashCompleteTriggers(c)
			}
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
		BashComplete: func(c *cli.Context) {
			switch len(c.Args()) {
			case 0:
				app.BashCompleteApps(c)
			case 1:
				fn.BashCompleteFns(c)
			case 2:
				BashCompleteTriggers(c)
			}
		},
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
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "endpoint",
				Usage: "Output the trigger HTTP endpoint if set",
			},
		},
		Before: func(ctx *cli.Context) error {
			var err error
			t.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			t.client = t.provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app-name> <function-name> <trigger-name> [property[.key]]",
		Action:    t.inspect,
		BashComplete: func(c *cli.Context) {
			switch len(c.Args()) {
			case 0:
				app.BashCompleteApps(c)
			case 1:
				fn.BashCompleteFns(c)
			case 2:
				BashCompleteTriggers(c)
			case 3:
				trigg, err := GetTriggerByAppFnAndTriggerNames(c.Args()[0], c.Args()[1], c.Args()[2])
				if err != nil {
					return
				}
				data, err := json.Marshal(trigg)
				if err != nil {
					return
				}
				var inspect map[string]interface{}
				err = json.Unmarshal(data, &inspect)
				if err != nil {
					return
				}
				for key := range inspect {
					fmt.Println(key)
				}
			}
		},
	}
}
