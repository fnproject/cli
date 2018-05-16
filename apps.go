package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"context"
	"strings"

	apiapps "github.com/fnproject/fn_go/client/apps"
	"github.com/fnproject/fn_go/models"
	"github.com/jmoiron/jsonq"
	"github.com/urfave/cli"
)

type appsCommands struct {
	create  string "create"
	delete  string
	list    string
	inspect string
	update  string
}

const (
	appsCreate  = "create"
	appsDelete  = "delete"
	appsList    = "list"
	appsInspect = "inspect"
	appsUpdate  = "update"
)

func (a *clientCmd) appsCommand(command string) cli.Command {

	switch command {
	case appsCreate:
		return a.getCreateAppsCommand()
	case appsList:
		return a.getListAppsCommand()
	case appsDelete:
		return a.getDeleteAppsCommand()
	case appsInspect:
		return a.getInspectAppsCommand()
	case appsUpdate:
		return a.getUpdateAppsCommand()
	}

	return cli.Command{}
}

func (a *clientCmd) list(c *cli.Context) error {
	params := &apiapps.GetAppsParams{Context: context.Background()}
	var resApps []*models.App
	for {
		resp, err := a.client.Apps.GetApps(params)
		if err != nil {
			switch e := err.(type) {
			case *apiapps.GetAppsAppNotFound:
				return fmt.Errorf("%v", e.Payload.Error.Message)
			default:
				return err
			}
		}

		resApps = append(resApps, resp.Payload.Apps...)

		n := c.Int64("n")
		if n < 0 {
			return errors.New("number of calls: negative value not allowed")
		}

		howManyMore := n - int64(len(resApps)+len(resp.Payload.Apps))
		if howManyMore <= 0 || resp.Payload.NextCursor == "" {
			break
		}

		params.Cursor = &resp.Payload.NextCursor
	}

	if len(resApps) == 0 {
		fmt.Println("no apps found")
		return nil
	}

	for _, app := range resApps {
		fmt.Println(app.Name)
	}

	return nil
}

func (a *clientCmd) create(c *cli.Context) error {
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
		default:
			return err
		}
	}

	fmt.Println("Successfully created app: ", resp.Payload.App.Name)
	return nil
}

func (a *clientCmd) update(c *cli.Context) error {
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

func (a *clientCmd) configSet(c *cli.Context) error {
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

func (a *clientCmd) configGet(c *cli.Context) error {
	appName := c.Args().Get(0)
	key := c.Args().Get(1)

	resp, err := a.client.Apps.GetAppsApp(&apiapps.GetAppsAppParams{
		App:     appName,
		Context: context.Background(),
	})

	if err != nil {
		return err
	}

	val, ok := resp.Payload.App.Config[key]
	if !ok {
		return fmt.Errorf("config key does not exist")
	}

	fmt.Println(val)

	return nil
}

func (a *clientCmd) configList(c *cli.Context) error {
	appName := c.Args().Get(0)

	resp, err := a.client.Apps.GetAppsApp(&apiapps.GetAppsAppParams{
		App:     appName,
		Context: context.Background(),
	})

	if err != nil {
		return err
	}

	for key, val := range resp.Payload.App.Config {
		fmt.Printf("%s=%s\n", key, val)
	}

	return nil
}

func (a *clientCmd) configUnset(c *cli.Context) error {
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

func (a *clientCmd) patchApp(appName string, app *models.App) error {
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
		default:
			return err
		}
	}

	return nil
}

func (a *clientCmd) inspect(c *cli.Context) error {
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
		default:
			return err
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

func (a *clientCmd) delete(c *cli.Context) error {
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
		}
		return err
	}

	fmt.Println("App", appName, "deleted")
	return nil
}

func (a *clientCmd) getCreateAppsCommand() cli.Command {
	return cli.Command{
		Name:      "apps",
		Usage:     "create a new app",
		ArgsUsage: "<app>",
		Action:    a.create,
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "config",
				Usage: "application configuration",
			},
		},
	}
}

func (a *clientCmd) getListAppsCommand() cli.Command {
	return cli.Command{
		Name:   "apps",
		Usage:  "list all apps",
		Action: a.list,
		Flags: []cli.Flag{
			cli.StringFlag{
				Name:  "cursor",
				Usage: "pagination cursor",
			},
			cli.Int64Flag{
				Name:  "n",
				Usage: "number of apps to return",
				Value: int64(100),
			},
		},
	}
}

func (a *clientCmd) getDeleteAppsCommand() cli.Command {
	return cli.Command{
		Name:    "delete",
		Aliases: []string{"d"},
		Usage:   "delete an app",
		Action:  a.delete,
	}
}

func (a *clientCmd) getInspectAppsCommand() cli.Command {
	return cli.Command{
		Name:      "inspect",
		Usage:     "retrieve one or all apps properties",
		ArgsUsage: "<app> [property.[key]]",
		Action:    a.inspect,
	}
}

func (a *clientCmd) getUpdateAppsCommand() cli.Command {
	return cli.Command{
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
	}
}
