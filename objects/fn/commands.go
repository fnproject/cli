package fn

import (
	"encoding/json"
	"fmt"

	client "github.com/fnproject/cli/client"
	"github.com/fnproject/cli/objects/app"
	"github.com/urfave/cli"
)

// Create function command
func Create() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:        "function",
		ShortName:   "func",
		Aliases:     []string{"f", "fn"},
		Category:    "MANAGEMENT COMMAND",
		Usage:       "Create a function within an application",
		Description: "This command creates a new function within an application.",
		Before: func(c *cli.Context) error {
			var err error
			providerAdapter, err := client.CurrentProviderAdapter()
			if err != nil {
				return err
			}
			f.apiClientAdapter = providerAdapter.APIClient()
			return nil
		},
		ArgsUsage: "<app-name> <function-name> <image>",
		Action:    f.create,
		Flags:     FnFlags,
		BashComplete: func(c *cli.Context) {
			if len(c.Args()) == 0 {
				app.BashCompleteApps(c)
			}
		},
	}
}

// List functions command
func List() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:        "functions",
		ShortName:   "funcs",
		Aliases:     []string{"f", "fn"},
		Usage:       "List functions for an application",
		Category:    "MANAGEMENT COMMAND",
		Description: "This command returns a list of functions for a created application.",
		Before: func(c *cli.Context) error {
			var err error
			providerAdapter, err := client.CurrentProviderAdapter()
			if err != nil {
				return err
			}
			f.apiClientAdapter = providerAdapter.APIClient()
			return nil
		},
		ArgsUsage: "<app-name>",
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
			cli.StringFlag{
				Name:  "output",
				Usage: "Output format (json)",
				Value: "",
			},
		},
		BashComplete: func(c *cli.Context) {
			switch len(c.Args()) {
			case 0:
				app.BashCompleteApps(c)
			}
		},
	}
}

// Delete function command
func Delete() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:        "function",
		ShortName:   "func",
		Aliases:     []string{"f", "fn"},
		Category:    "MANAGEMENT COMMAND",
		Description: "This command deletes an existing function from an application.",
		Usage:       "Delete a function from an application",
		Before: func(c *cli.Context) error {
			var err error
			f.providerAdapter, err = client.CurrentProviderAdapter()
			if err != nil {
				return err
			}
			f.apiClientAdapter = f.providerAdapter.APIClient()
			return nil
		},
		ArgsUsage: "<app-name> <function-name>",
		Action:    f.delete,
		BashComplete: func(c *cli.Context) {
			switch len(c.Args()) {
			case 0:
				app.BashCompleteApps(c)
			case 1:
				BashCompleteFns(c)
			}
		},
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "force, f",
				Usage: "Forces this delete (you will not be asked if you wish to continue with the delete)",
			},
			cli.BoolFlag{
				Name:  "recursive, r",
				Usage: "Delete this function and all associated resources (can fail part way through execution after deleting some resources without the ability to undo)",
			},
		},
	}
}

// Inspect function command
func Inspect() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:        "function",
		ShortName:   "func",
		Aliases:     []string{"f", "fn"},
		Category:    "MANAGEMENT COMMAND",
		Usage:       "Retrieve one or all properties for a function",
		Description: "This command inspects properties of a function.",
		Before: func(c *cli.Context) error {
			var err error
			providerAdapter, err := client.CurrentProviderAdapter()
			if err != nil {
				return err
			}
			f.apiClientAdapter = providerAdapter.APIClient()
			return nil
		},
		Flags: []cli.Flag{
			cli.BoolFlag{
				Name:  "endpoint",
				Usage: "Output the function invoke endpoint if set",
			},
		},
		ArgsUsage: "<app-name> <function-name> [property[.key]]",
		Action:    f.inspect,
		BashComplete: func(c *cli.Context) {
			switch len(c.Args()) {
			case 0:
				app.BashCompleteApps(c)
			case 1:
				BashCompleteFns(c)
			case 2:
				fn, err := getFnByAppAndFnName(c.Args()[0], c.Args()[1])
				if err != nil {
					return
				}
				data, err := json.Marshal(fn)
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

// Update function command
func Update() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:        "function",
		ShortName:   "func",
		Aliases:     []string{"f", "fn"},
		Category:    "MANAGEMENT COMMAND",
		Usage:       "Update a function in application",
		Description: "This command updates a function in an application.",
		Before: func(c *cli.Context) error {
			var err error
			providerAdapter, err := client.CurrentProviderAdapter()
			if err != nil {
				return err
			}
			f.apiClientAdapter = providerAdapter.APIClient()
			return nil
		},
		ArgsUsage: "<app-name> <function-name>",
		Action:    f.update,
		Flags:     updateFnFlags,
		BashComplete: func(c *cli.Context) {
			switch len(c.Args()) {
			case 0:
				app.BashCompleteApps(c)
			case 1:
				BashCompleteFns(c)
			}
		},
	}
}

// GetConfig for function command
func GetConfig() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:        "function",
		ShortName:   "func",
		Aliases:     []string{"f", "fn"},
		Category:    "MANAGEMENT COMMAND",
		Usage:       "Inspect configuration key for this function",
		Description: "This command gets the configuration of a specific function for an application.",
		Before: func(c *cli.Context) error {
			var err error
			providerAdapter, err := client.CurrentProviderAdapter()
			if err != nil {
				return err
			}
			f.apiClientAdapter = providerAdapter.APIClient()
			return nil
		},
		ArgsUsage: "<app-name> <function-name> <key>",
		Action:    f.getConfig,
		BashComplete: func(c *cli.Context) {
			switch len(c.Args()) {
			case 0:
				app.BashCompleteApps(c)
			case 1:
				BashCompleteFns(c)
			case 2:
				fn, err := getFnByAppAndFnName(c.Args()[0], c.Args()[1])
				if err != nil {
					return
				}
				for key := range fn.Config {
					fmt.Println(key)
				}
			}
		},
	}
}

// SetConfig for function command
func SetConfig() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:        "function",
		ShortName:   "func",
		Aliases:     []string{"f", "fn"},
		Category:    "MANAGEMENT COMMAND",
		Usage:       "Store a configuration key for this function",
		Description: "This command sets the configuration of a specific function for an application.",
		Before: func(c *cli.Context) error {
			var err error
			providerAdapter, err := client.CurrentProviderAdapter()
			if err != nil {
				return err
			}
			f.apiClientAdapter = providerAdapter.APIClient()
			return nil
		},
		ArgsUsage: "<app-name> <function-name> <key> <value>",
		Action:    f.setConfig,
		BashComplete: func(c *cli.Context) {
			switch len(c.Args()) {
			case 0:
				app.BashCompleteApps(c)
			case 1:
				BashCompleteFns(c)
			}
		},
	}
}

// ListConfig for function command
func ListConfig() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:        "function",
		ShortName:   "func",
		Aliases:     []string{"f", "fn"},
		Category:    "MANAGEMENT COMMAND",
		Usage:       "List configuration key/value pairs for this function",
		Description: "This command returns a list of configurations for a specific function.",
		Before: func(c *cli.Context) error {
			var err error
			providerAdapter, err := client.CurrentProviderAdapter()
			if err != nil {
				return err
			}
			f.apiClientAdapter = providerAdapter.APIClient()
			return nil
		},
		ArgsUsage: "<app-name> <function-name>",
		Action:    f.listConfig,
		BashComplete: func(c *cli.Context) {
			switch len(c.Args()) {
			case 0:
				app.BashCompleteApps(c)
			case 1:
				BashCompleteFns(c)
			}
		},
	}
}

// UnsetConfig for function command
func UnsetConfig() cli.Command {
	f := fnsCmd{}
	return cli.Command{
		Name:        "function",
		ShortName:   "func",
		Aliases:     []string{"f", "fn"},
		Category:    "MANAGEMENT COMMAND",
		Usage:       "Remove a configuration key for this function",
		Description: "This command removes a configuration of a specific function.",
		Before: func(c *cli.Context) error {
			var err error
			providerAdapter, err := client.CurrentProviderAdapter()
			if err != nil {
				return err
			}
			f.apiClientAdapter = providerAdapter.APIClient()
			return nil
		},
		ArgsUsage: "<app-name> <function-name> <key>",
		Action:    f.unsetConfig,
		BashComplete: func(c *cli.Context) {
			switch len(c.Args()) {
			case 0:
				app.BashCompleteApps(c)
			case 1:
				BashCompleteFns(c)
			case 2:
				fn, err := getFnByAppAndFnName(c.Args()[0], c.Args()[1])
				if err != nil {
					return
				}
				for key := range fn.Config {
					fmt.Println(key)
				}
			}
		},
	}
}
