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

package runtime

import (
	"github.com/fnproject/cli/client"
	"github.com/fnproject/fn_go/provider"
	"github.com/fnproject/fn_go/provider/oracle"
	"github.com/urfave/cli"
)

// List runtimes command
func ListRuntimes() cli.Command {
	cmd := runtimeCmd{}
	return cli.Command{
		Name:        "runtimes",
		Usage:       "List supported runtimes",
		Category:    "MANAGEMENT COMMAND",
		Description: "This command lists all supported runtimes.",
		Before: func(c *cli.Context) error {
			var err error
			cmd.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			cmd.providerName = c.String("provider")
			return nil
		},
		Action: cmd.listRuntimes,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "output",
				Usage: "Output format (json)",
			},
			cli.StringFlag{
				Name:  "provider",
				Usage: "Override provider name",
			},
		},
		BashComplete: func(c *cli.Context) {
			provider, err := client.CurrentProvider()
			if err != nil {
				return
			}
			if _, ok := provider.(*oracle.OracleProvider); !ok {
				return
			}
		},
	}
}

// List runtime versions command
func ListRuntimeVersions() cli.Command {
	cmd := runtimeCmd{}
	return cli.Command{
		Name:        "runtime-versions",
		Usage:       "List runtime versions for a runtime",
		Category:    "MANAGEMENT COMMAND",
		Description: "This command lists runtime versions for a runtime name.",
		Before: func(c *cli.Context) error {
			var err error
			cmd.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			cmd.providerName = c.String("provider")
			return nil
		},
		Action: cmd.listRuntimeVersions,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "runtime-name",
				Usage: "Runtime name (e.g. java21.ol10)",
			},
			cli.StringFlag{
				Name:  "output",
				Usage: "Output format (json)",
			},
			cli.StringFlag{
				Name:  "provider",
				Usage: "Override provider name",
			},
		},
		BashComplete: func(c *cli.Context) {
			if c.String("runtime-name") != "" {
				return
			}
			suggestRuntimeNames(c)
		},
	}
}

// Get latest runtime version command
func GetLatestRuntimeVersion() cli.Command {
	cmd := runtimeCmd{}
	return cli.Command{
		Name:        "latest-runtime-version",
		Usage:       "Get the latest runtime version for a runtime",
		Category:    "MANAGEMENT COMMAND",
		Description: "This command gets the latest runtime version for a runtime name.",
		Before: func(c *cli.Context) error {
			var err error
			cmd.provider, err = client.CurrentProvider()
			if err != nil {
				return err
			}
			cmd.providerName = c.String("provider")
			return nil
		},
		Action: cmd.getLatestRuntimeVersion,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "runtime-name",
				Usage: "Runtime name (e.g. java21.ol10)",
			},
			cli.StringFlag{
				Name:  "output",
				Usage: "Output format (json)",
			},
			cli.StringFlag{
				Name:  "provider",
				Usage: "Override provider name",
			},
		},
		BashComplete: func(c *cli.Context) {
			if c.String("runtime-name") != "" {
				return
			}
			suggestRuntimeNames(c)
		},
	}
}

type runtimeCmd struct {
	provider     provider.Provider
	providerName string
}