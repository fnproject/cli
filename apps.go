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

func (client *fnClient) apps(command string) cli.Command {
	var aCmd cli.Command

	switch command {
	case CreateCmd:
		aCmd = client.getCreateAppCommand()
	case ListCmd:
		aCmd = client.getListAppsCommand()
	case DeleteCmd:
		aCmd = client.getDeleteAppCommand()
	case InspectCmd:
		aCmd = client.getInspectAppsCommand()
	case UpdateCmd:
		aCmd = client.getUpdateAppCommand()
	case ConfigCmd:
		aCmd = client.getConfigAppsCommand()
	}

	return aCmd
}

func (client *fnClient) getCreateAppCommand() cli.Command {
	return cli.Command{
		Name:      "app",
		Usage:     "Create a new application",
		ArgsUsage: "<app>",
		Action:    client.createApp,
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "config",
				Usage: "application configuration",
			},
		},
	}
}

func (client *fnClient) getListAppsCommand() cli.Command {
	return cli.Command{
		Name:   "apps",
		Usage:  "List all applications ",
		Action: client.listApps,
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

func (client *fnClient) getDeleteAppCommand() cli.Command {
	return cli.Command{
		Name:   "app",
		Usage:  "Delete an application",
		Action: client.deleteApps,
	}
}

func (client *fnClient) getInspectAppsCommand() cli.Command {
	return cli.Command{
		Name:      "apps",
		Usage:     "retrieve one or all apps properties",
		ArgsUsage: "<app> [property.[key]]",
		Action:    client.inspectApps,
	}
}

func (client *fnClient) getUpdateAppCommand() cli.Command {
	return cli.Command{
		Name:      "app",
		Usage:     "update an application",
		ArgsUsage: "<app>",
		Action:    client.updateApps,
		Flags: []cli.Flag{
			cli.StringSliceFlag{
				Name:  "config,c",
				Usage: "route configuration",
			},
		},
	}
}

func (client *fnClient) getConfigAppsCommand() cli.Command {
	return cli.Command{
		Name:  "apps",
		Usage: "manage apps configs",
		Subcommands: []cli.Command{
			{
				Name:      "set",
				Aliases:   []string{"s"},
				Usage:     "store a configuration key for this application",
				ArgsUsage: "<app> <key> <value>",
				Action:    client.configSetApps,
			},
			{
				Name:      "get",
				Aliases:   []string{"g"},
				Usage:     "inspect configuration key for this application",
				ArgsUsage: "<app> <key>",
				Action:    client.configGetApps,
			},
			{
				Name:      "list",
				Aliases:   []string{"l"},
				Usage:     "list configuration key/value pairs for this application",
				ArgsUsage: "<app>",
				Action:    client.configListApps,
			},
			{
				Name:      "unset",
				Aliases:   []string{"u"},
				Usage:     "remove a configuration key for this application",
				ArgsUsage: "<app> <key>",
				Action:    client.configUnsetApps,
			},
		},
	}
}

func (client *fnClient) listApps(c *cli.Context) error {
	params := &apiapps.GetAppsParams{Context: context.Background()}
	var resApps []*models.App
	for {
		resp, err := client.client.Apps.GetApps(params)
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

func (client *fnClient) createApp(c *cli.Context) error {
	body := &models.AppWrapper{App: &models.App{
		Name:   c.Args().Get(0),
		Config: extractEnvConfig(c.StringSlice("config")),
	}}

	resp, err := client.client.Apps.PostApps(&apiapps.PostAppsParams{
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

func (client *fnClient) updateApps(c *cli.Context) error {
	appName := c.Args().First()

	patchedApp := &models.App{
		Config: extractEnvConfig(c.StringSlice("config")),
	}

	err := client.patchApp(appName, patchedApp)
	if err != nil {
		return err
	}

	fmt.Println("app", appName, "updated")
	return nil
}

func (client *fnClient) configSetApps(c *cli.Context) error {
	appName := c.Args().Get(0)
	key := c.Args().Get(1)
	value := c.Args().Get(2)

	app := &models.App{
		Config: make(map[string]string),
	}

	app.Config[key] = value

	if err := client.patchApp(appName, app); err != nil {
		return fmt.Errorf("error updating app configuration: %v", err)
	}

	fmt.Println(appName, "updated", key, "with", value)
	return nil
}

func (client *fnClient) configGetApps(c *cli.Context) error {
	appName := c.Args().Get(0)
	key := c.Args().Get(1)

	resp, err := client.client.Apps.GetAppsApp(&apiapps.GetAppsAppParams{
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

func (client *fnClient) configListApps(c *cli.Context) error {
	appName := c.Args().Get(0)

	resp, err := client.client.Apps.GetAppsApp(&apiapps.GetAppsAppParams{
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

func (client *fnClient) configUnsetApps(c *cli.Context) error {
	appName := c.Args().Get(0)
	key := c.Args().Get(1)

	app := &models.App{
		Config: make(map[string]string),
	}

	app.Config[key] = ""

	if err := client.patchApp(appName, app); err != nil {
		return fmt.Errorf("error updating app configuration: %v", err)
	}

	fmt.Printf("removed key '%s' from app '%s' \n", key, appName)
	return nil
}

func (client *fnClient) patchApp(appName string, app *models.App) error {
	_, err := client.client.Apps.PatchAppsApp(&apiapps.PatchAppsAppParams{
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

func (client *fnClient) inspectApps(c *cli.Context) error {
	if c.Args().Get(0) == "" {
		return errors.New("missing app name after the inspect command")
	}

	appName := c.Args().First()
	prop := c.Args().Get(1)

	resp, err := client.client.Apps.GetAppsApp(&apiapps.GetAppsAppParams{
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

func (client *fnClient) deleteApps(c *cli.Context) error {
	appName := c.Args().First()
	if appName == "" {
		return errors.New("app name required to delete")
	}

	_, err := client.client.Apps.DeleteAppsApp(&apiapps.DeleteAppsAppParams{
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
