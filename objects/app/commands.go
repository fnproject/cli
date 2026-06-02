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

package app

import (
	"encoding/json"
	"fmt"

	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

// Create app command
func Create() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:        "app",
		Usage:       "Create a new application",
		Category:    "MANAGEMENT COMMAND",
		Description: "This command creates a new application.\n	Fn supports grouping functions into a set that defines an application (or API), making it easy to organize and deploy.\n	Applications define a namespace to organize functions and can contain configuration values that are shared across all functions in that application.",
		Aliases:     []string{"apps", "a"},
		Before: func(c *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			a.provider = provider
			a.client = provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app-name>",
		Action:    a.create,
		Flags: append([]cli.Flag{
			cli.StringSliceFlag{
				Name:  "config",
				Usage: "Application configuration",
			},
			cli.StringSliceFlag{
				Name:  "annotation",
				Usage: "Application annotations",
			},
			cli.StringSliceFlag{
				Name:  "tag",
				Usage: "Freeform tag in key=value form (can be specified multiple times)",
			},
			cli.StringSliceFlag{
				Name:  "defined-tag",
				Usage: "Defined tag in namespace.key=value form (can be specified multiple times)",
			},
			cli.StringFlag{
				Name:  "syslog-url",
				Usage: "Syslog URL to send application logs to",
			},
			cli.StringFlag{
				Name:  "shape",
				Usage: "Valid values are GENERIC_X86, GENERIC_ARM and GENERIC_X86_ARM. Default is GENERIC_X86. Setting this to GENERIC_X86, will run the functions in the application on X86 processor architecture.\n Setting this to GENERIC_ARM, will run the functions in the application on ARM processor architecture.\n When set to 'GENERIC_X86_ARM', functions in the application are run on either X86 or ARM processor architecture.\n Accepted values are:\n GENERIC_X86, GENERIC_ARM, GENERIC_X86_ARM",
			},
			cli.StringSliceFlag{
				Name:  "subnet-id",
				Usage: "Subnet OCID for OCI Functions applications (can be specified multiple times; maps to oracle.com/oci/subnetIds)",
			},
			cli.StringFlag{Name: "from-json", Usage: "Provide operation input as inline JSON or file://path"},
			cli.StringFlag{Name: "wait-for-state", Usage: "Wait until the application reaches the given lifecycle state"},
			cli.IntFlag{Name: "max-wait-seconds", Usage: "Maximum seconds to wait for the requested lifecycle state"},
			cli.IntFlag{Name: "wait-interval-seconds", Usage: "Polling interval in seconds while waiting for lifecycle state"},
		}, GeneratedOCIParityCreateUpdateAppFlags()...),
	}
}

// List apps command
func List() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:        "apps",
		Usage:       "List all created applications",
		Category:    "MANAGEMENT COMMAND",
		Description: "This command provides a list of defined applications.",
		Aliases:     []string{"app", "a"},
		Before: func(c *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			a.provider = provider
			a.client = provider.APIClientv2()
			return nil
		},
		Action: a.list,
		Flags: append([]cli.Flag{
			cli.StringFlag{
				Name:  "cursor",
				Usage: "Pagination cursor",
			},
			cli.Int64Flag{
				Name:  "n",
				Usage: "Number of apps to return",
				Value: int64(100),
			},
			cli.StringFlag{
				Name:  "output",
				Usage: "Output format (json)",
			},
			cli.StringFlag{
				Name:  "from-json",
				Usage: "Provide operation input as inline JSON or file://path",
			},
		}, GeneratedOCIParityListAppFlags()...),
	}
}

// Delete app command
func Delete() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:        "app",
		Usage:       "Delete an application",
		Category:    "MANAGEMENT COMMAND",
		Description: "This command deletes a created application.",
		ArgsUsage:   "<app-name>",
		Aliases:     []string{"apps", "a"},
		Before: func(c *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			a.provider = provider
			a.client = provider.APIClientv2()
			return nil
		},
		Action: a.delete,
		BashComplete: func(c *cli.Context) {
			args := c.Args()
			if len(args) == 0 {
				BashCompleteApps(c)
			}
		},
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "force, f",
				Usage: "Forces this delete (you will not be asked if you wish to continue with the delete)",
			},
			cli.BoolFlag{
				Name:  "recursive, r",
				Usage: "Delete this app and all associated resources (can fail part way through execution after deleting some resources without the ability to undo)",
			},
			cli.StringFlag{Name: "from-json", Usage: "Provide operation input as inline JSON or file://path"},
			cli.StringFlag{Name: "if-match", Usage: "Apply optimistic concurrency control using the provided etag"},
			cli.StringFlag{Name: "wait-for-state", Usage: "Wait until the application reaches the given lifecycle state"},
			cli.IntFlag{Name: "max-wait-seconds", Usage: "Maximum seconds to wait for the requested lifecycle state"},
			cli.IntFlag{Name: "wait-interval-seconds", Usage: "Polling interval in seconds while waiting for lifecycle state"},
		},
	}
}

// ChangeCompartment app command
func ChangeCompartment() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:        "app",
		Usage:       "Move an OCI Functions application to another compartment",
		Category:    "MANAGEMENT COMMAND",
		Description: "This command moves an OCI Functions application to another compartment in the same tenancy.",
		Aliases:     []string{"apps", "a"},
		Before: func(c *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			a.provider = provider
			a.client = provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app-name>",
		Action:    a.changeCompartment,
		Flags: []cli.Flag{
			cli.StringFlag{Name: "compartment-id", Usage: "Target compartment OCID"},
			cli.StringFlag{Name: "from-json", Usage: "Provide operation input as inline JSON or file://path"},
			cli.StringFlag{Name: "if-match", Usage: "Apply optimistic concurrency control using the provided etag"},
			cli.StringFlag{Name: "wait-for-state", Usage: "Wait until the application reaches the given lifecycle state"},
			cli.IntFlag{Name: "max-wait-seconds", Usage: "Maximum seconds to wait for the requested lifecycle state"},
			cli.IntFlag{Name: "wait-interval-seconds", Usage: "Polling interval in seconds while waiting for lifecycle state"},
		},
		BashComplete: func(c *cli.Context) {
			if len(c.Args()) == 0 {
				BashCompleteApps(c)
			}
		},
	}
}

// Inspect app command
func Inspect() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:        "app",
		Usage:       "Retrieve one or all apps properties",
		Description: "This command inspects properties of an application.",
		Category:    "MANAGEMENT COMMAND",
		Aliases:     []string{"apps", "a"},
		Before: func(c *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			a.provider = provider
			a.client = provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app-name> [property.[key]]",
		Action:    a.inspect,
		BashComplete: func(c *cli.Context) {
			switch len(c.Args()) {
			case 0:
				BashCompleteApps(c)
			case 1:
				provider, err := client.CurrentProvider()
				if err != nil {
					return
				}
				app, err := GetAppByName(provider.APIClientv2(), c.Args()[0])
				if err != nil {
					return
				}
				data, err := json.Marshal(app)
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

// Update app command
func Update() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:        "app",
		Usage:       "Update an application",
		Category:    "MANAGEMENT COMMAND",
		Description: "This command updates a created application.",
		Aliases:     []string{"apps", "a"},
		Before: func(c *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			a.provider = provider
			a.client = provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app-name>",
		Action:    a.update,
		Flags: append([]cli.Flag{
			cli.StringSliceFlag{
				Name:  "config,c",
				Usage: "Application configuration",
			},
			cli.StringSliceFlag{
				Name:  "annotation",
				Usage: "Application annotations",
			},
			cli.StringSliceFlag{
				Name:  "tag",
				Usage: "Freeform tag in key=value form (can be specified multiple times)",
			},
			cli.StringSliceFlag{
				Name:  "defined-tag",
				Usage: "Defined tag in namespace.key=value form (can be specified multiple times)",
			},
			cli.StringSliceFlag{
				Name:  "remove-tag",
				Usage: "Remove a freeform tag by key (can be specified multiple times)",
			},
			cli.StringSliceFlag{
				Name:  "remove-defined-tag",
				Usage: "Remove a defined tag by namespace.key (can be specified multiple times)",
			},
			cli.BoolFlag{
				Name:  "clear-tags",
				Usage: "Clear all freeform and defined tags",
			},
			cli.BoolFlag{
				Name:  "clear-freeform-tags",
				Usage: "Clear all freeform tags",
			},
			cli.BoolFlag{
				Name:  "clear-defined-tags",
				Usage: "Clear all defined tags",
			},
			cli.StringFlag{
				Name:  "syslog-url",
				Usage: "Syslog URL to send application logs to",
			},
			cli.StringSliceFlag{
				Name:  "subnet-id",
				Usage: "Subnet OCID for OCI Functions applications (accepted for non-Oracle providers; Oracle-backed update currently not supported)",
			},
			cli.StringFlag{Name: "from-json", Usage: "Provide operation input as inline JSON or file://path"},
			cli.StringFlag{Name: "if-match", Usage: "Apply optimistic concurrency control using the provided etag"},
			cli.StringFlag{Name: "wait-for-state", Usage: "Wait until the application reaches the given lifecycle state"},
			cli.IntFlag{Name: "max-wait-seconds", Usage: "Maximum seconds to wait for the requested lifecycle state"},
			cli.IntFlag{Name: "wait-interval-seconds", Usage: "Polling interval in seconds while waiting for lifecycle state"},
		}, GeneratedOCIParityCreateUpdateAppFlags()...),
		BashComplete: func(c *cli.Context) {
			args := c.Args()
			if len(args) == 0 {
				BashCompleteApps(c)
			}
		},
	}
}

// SetConfig for function command
func SetConfig() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:        "app",
		Usage:       "Store a configuration key for this application",
		Description: "This command sets configurations for an application.",
		Category:    "MANAGEMENT COMMAND",
		Aliases:     []string{"apps", "a"},
		Before: func(c *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			a.client = provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app-name> <key> <value>",
		Action:    a.setConfig,
		BashComplete: func(c *cli.Context) {
			args := c.Args()
			if len(args) == 0 {
				BashCompleteApps(c)
			}
		},
	}
}

// ListConfig for app command
func ListConfig() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:        "app",
		Usage:       "List configuration key/value pairs for this application",
		Category:    "MANAGEMENT COMMAND",
		Description: "This command lists the configuration of an application.",
		Aliases:     []string{"apps", "a"},
		Before: func(c *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			a.client = provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app-name>",
		Action:    a.listConfig,
		BashComplete: func(c *cli.Context) {
			args := c.Args()
			if len(args) == 0 {
				BashCompleteApps(c)
			}
		},
	}
}

// GetConfig for function command
func GetConfig() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:        "app",
		Usage:       "Inspect configuration key for this application",
		Description: "This command gets the configuration of an application.",
		Category:    "MANAGEMENT COMMAND",
		Aliases:     []string{"apps", "a"},
		Before: func(c *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			a.client = provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app-name> <key>",
		Action:    a.getConfig,
		BashComplete: func(c *cli.Context) {
			switch len(c.Args()) {
			case 0:
				BashCompleteApps(c)
			case 1:
				provider, err := client.CurrentProvider()
				if err != nil {
					return
				}
				app, err := GetAppByName(provider.APIClientv2(), c.Args()[0])
				if err != nil {
					return
				}
				for key := range app.Config {
					fmt.Println(key)
				}
			}
		},
	}
}

// UnsetConfig for app command
func UnsetConfig() cli.Command {
	a := appsCmd{}
	return cli.Command{
		Name:        "app",
		Usage:       "Remove a configuration key for this application.",
		Description: "This command removes a configuration for an application.",
		Category:    "MANAGEMENT COMMAND",
		Aliases:     []string{"apps", "a"},
		Before: func(c *cli.Context) error {
			provider, err := client.CurrentProvider()
			if err != nil {
				return err
			}
			a.client = provider.APIClientv2()
			return nil
		},
		ArgsUsage: "<app-name> <key>",
		Action:    a.unsetConfig,
		BashComplete: func(c *cli.Context) {
			switch len(c.Args()) {
			case 0:
				BashCompleteApps(c)
			case 1:
				provider, err := client.CurrentProvider()
				if err != nil {
					return
				}
				app, err := GetAppByName(provider.APIClientv2(), c.Args()[0])
				if err != nil {
					return
				}
				for key := range app.Config {
					fmt.Println(key)
				}
			}
		},
	}
}
