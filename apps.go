package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"context"
	"strings"

	client "github.com/fnproject/cli/client"
	fnclient "github.com/fnproject/fn_go/client"
	apiapps "github.com/fnproject/fn_go/client/apps"
	"github.com/fnproject/fn_go/models"
	"github.com/jmoiron/jsonq"
	"github.com/urfave/cli"
)

type appsCmd struct {
	client *fnclient.Functions
}

func apps() cli.Command {
	a := appsCmd{client: client.APIClient()}

	return cli.Command{
		Name:  "apps",
		Usage: "manage applications",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Aliases:   []string{"c"},
				Usage:     "create a new app",
				ArgsUsage: "<app>",
				Action:    a.create,
				Flags: []cli.Flag{
					cli.StringSliceFlag{
						Name:  "config",
						Usage: "application configuration",
					},
				},
			},
			{
				Name:      "inspect",
				Aliases:   []string{"i"},
				Usage:     "retrieve one or all apps properties",
				ArgsUsage: "<app> [property.[key]]",
				Action:    a.inspect,
			},
			{
				Name:      "update",
				Aliases:   []string{"u"},
				Usage:     "update an `app`",
				ArgsUsage: "<app>",
				Action:    a.update,
				Flags: []cli.Flag{
					cli.StringSliceFlag{
						Name:  "config,c",
						Usage: "route configuration",
					},
				},
			},
			{
				Name:  "config",
				Usage: "manage your apps's function configs",
				Subcommands: []cli.Command{
					{
						Name:      "set",
						Aliases:   []string{"s"},
						Usage:     "store a configuration key for this application",
						ArgsUsage: "<app> <key> <value>",
						Action:    a.configSet,
					},
					{
						Name:      "unset",
						Aliases:   []string{"u"},
						Usage:     "remove a configuration key for this application",
						ArgsUsage: "<app> <key>",
						Action:    a.configUnset,
					},
				},
			},
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "list all apps",
				Action:  a.list,
			},
			{
				Name:    "delete",
				Aliases: []string{"d"},
				Usage:   "delete an app",
				Action:  a.delete,
			},
		},
	}
}

func (a *appsCmd) list(c *cli.Context) error {
	resp, err := a.client.Apps.GetApps(&apiapps.GetAppsParams{
		Context: context.Background(),
	})

	if err != nil {
		switch e := err.(type) {
		case *apiapps.GetAppsAppNotFound:
			return fmt.Errorf("%v", e.Payload.Error.Message)
		case *apiapps.GetAppsAppDefault:
			return fmt.Errorf("%v", e.Payload.Error.Message)
		case *apiapps.GetAppsDefault:
			// this is the one getting called, not sure what the one above is?
			return fmt.Errorf("%v", e.Payload.Error.Message)
		default:
			return fmt.Errorf("%v", err)
		}
	}

	if len(resp.Payload.Apps) == 0 {
		fmt.Println("no apps found")
		return nil
	}

	for _, app := range resp.Payload.Apps {
		fmt.Println(app.Name)
	}

	return nil
}

func (a *appsCmd) create(c *cli.Context) error {
	body := &models.AppWrapper{App: &models.App{
		Name:   c.Args().Get(0),
		Config: extractEnvConfig(c.StringSlice("config")),
	}}

	resp, err := a.client.Apps.PostApps(&apiapps.PostAppsParams{
		Context: context.Background(),
		Body:    body,
	})

	if err != nil {
		switch e := err.(type) {
		case *apiapps.PostAppsBadRequest:
			return fmt.Errorf("%v", e.Payload.Error.Message)
		case *apiapps.PostAppsConflict:
			return fmt.Errorf("%v", e.Payload.Error.Message)
		case *apiapps.PostAppsDefault:
			return fmt.Errorf("%v", e.Payload.Error.Message)
		default:
			return fmt.Errorf("%v", err)
		}
	}

	fmt.Println("Successfully created app: ", resp.Payload.App.Name)
	return nil
}

func (a *appsCmd) update(c *cli.Context) error {
	appName := c.Args().First()

	patchedApp := &models.App{
		Config: extractEnvConfig(c.StringSlice("config")),
	}

	err := a.patchApp(appName, patchedApp)
	if err != nil {
		return err
	}

	fmt.Println("app", appName, "updated")
	return nil
}

func (a *appsCmd) configSet(c *cli.Context) error {
	appName := c.Args().Get(0)
	key := c.Args().Get(1)
	value := c.Args().Get(2)

	app := &models.App{
		Config: make(map[string]string),
	}

	app.Config[key] = value

	if err := a.patchApp(appName, app); err != nil {
		return fmt.Errorf("error updating app configuration: %v", err)
	}

	fmt.Println(appName, "updated", key, "with", value)
	return nil
}

func (a *appsCmd) configUnset(c *cli.Context) error {
	appName := c.Args().Get(0)
	key := c.Args().Get(1)

	app := &models.App{
		Config: make(map[string]string),
	}

	app.Config[key] = ""

	if err := a.patchApp(appName, app); err != nil {
		return fmt.Errorf("error updating app configuration: %v", err)
	}

	fmt.Printf("removed key '%s' from app '%s' \n", key, appName)
	return nil
}

func (a *appsCmd) patchApp(appName string, app *models.App) error {
	_, err := a.client.Apps.PatchAppsApp(&apiapps.PatchAppsAppParams{
		Context: context.Background(),
		App:     appName,
		Body:    &models.AppWrapper{App: app},
	})

	if err != nil {
		switch e := err.(type) {
		case *apiapps.PatchAppsAppBadRequest:
			return errors.New(e.Payload.Error.Message)
		case *apiapps.PatchAppsAppNotFound:
			return errors.New(e.Payload.Error.Message)
		case *apiapps.PatchAppsAppDefault:
			return errors.New(e.Payload.Error.Message)
		default:
			return fmt.Errorf("%v", err)
		}
	}

	return nil
}

func (a *appsCmd) inspect(c *cli.Context) error {
	if c.Args().Get(0) == "" {
		return errors.New("missing app name after the inspect command")
	}

	appName := c.Args().First()
	prop := c.Args().Get(1)

	resp, err := a.client.Apps.GetAppsApp(&apiapps.GetAppsAppParams{
		Context: context.Background(),
		App:     appName,
	})

	if err != nil {
		switch e := err.(type) {
		case *apiapps.GetAppsAppNotFound:
			return fmt.Errorf("%v", e.Payload.Error.Message)
		case *apiapps.GetAppsAppDefault:
			return fmt.Errorf("%v", e.Payload.Error.Message)
		default:
			return fmt.Errorf("%v", err)
		}
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")

	if prop == "" {
		enc.Encode(resp.Payload.App)
		return nil
	}

	// TODO: we really need to marshal it here just to
	// unmarshal as map[string]interface{}?
	data, err := json.Marshal(resp.Payload.App)
	if err != nil {
		return fmt.Errorf("could not marshal app: %v", err)
	}
	var inspect map[string]interface{}
	err = json.Unmarshal(data, &inspect)
	if err != nil {
		return fmt.Errorf("could not unmarshal data: %v", err)
	}

	jq := jsonq.NewQuery(inspect)
	field, err := jq.Interface(strings.Split(prop, ".")...)
	if err != nil {
		return fmt.Errorf("failed to inspect field %v", prop)
	}
	enc.Encode(field)

	return nil
}

func (a *appsCmd) delete(c *cli.Context) error {
	appName := c.Args().First()
	if appName == "" {
		return errors.New("app name required to delete")
	}

	_, err := a.client.Apps.DeleteAppsApp(&apiapps.DeleteAppsAppParams{
		Context: context.Background(),
		App:     appName,
	})

	if err != nil {
		switch e := err.(type) {
		case *apiapps.DeleteAppsAppNotFound:
			return errors.New(e.Payload.Error.Message)
		case *apiapps.DeleteAppsAppDefault:
			return errors.New(e.Payload.Error.Message)
		}
		return fmt.Errorf("%v", err)
	}

	fmt.Println("App", appName, "deleted")
	return nil
}
