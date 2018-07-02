package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"context"
	"strings"

	"github.com/fnproject/cli/common"
	"github.com/fnproject/fn_go/clientv2"
	apiapps "github.com/fnproject/fn_go/clientv2/apps"
	models "github.com/fnproject/fn_go/modelsv2"
	"github.com/fnproject/fn_go/provider"
	"github.com/jmoiron/jsonq"
	"github.com/urfave/cli"
)

type appsCmd struct {
	provider provider.Provider
	client   *clientv2.Fn
}

func (a *appsCmd) list(c *cli.Context) error {
	params := &apiapps.ListAppsParams{Context: context.Background()}
	var resApps []*models.App

	for {
		resp, err := a.client.Apps.ListApps(params)
		if err != nil {
			return err
		}

		if len(resp.Payload.Items) == 0 {
			break
		}

		n := c.Int64("n")
		if n < 0 {
			return errors.New("Number of calls: negative value not allowed")
		}

		resApps = append(resApps, resp.Payload.Items...)
		howManyMore := n - int64(len(resApps)+len(resp.Payload.Items))
		if howManyMore <= 0 || resp.Payload.NextCursor == "" {
			break
		}

		params.Cursor = &resp.Payload.NextCursor
	}

	if len(resApps) == 0 {
		fmt.Fprint(os.Stderr, "No apps found\n")
		return nil
	}

	for _, app := range resApps {
		fmt.Println(app.Name)
	}

	return nil
}

func appWithFlags(c *cli.Context, app *models.App) {
	if len(app.Config) == 0 {
		app.Config = common.ExtractEnvConfig(c.StringSlice("config"))
	}
	if len(app.Annotations) == 0 {
		if len(c.StringSlice("annotation")) > 0 {
			app.Annotations = common.ExtractAnnotations(c)
		}
	}
}

func (a *appsCmd) create(c *cli.Context) error {
	app := &models.App{
		Name: c.Args().Get(0),
	}

	appWithFlags(c, app)

	resp, err := a.client.Apps.CreateApp(&apiapps.CreateAppParams{
		Context: context.Background(),
		Body:    app,
	})

	if err != nil {
		switch e := err.(type) {
		case *apiapps.CreateAppBadRequest:
			return fmt.Errorf("%v", e.Payload.Message)

		default:
			return err
		}
	}

	fmt.Println("Successfully created app: ", resp.Payload.Name)
	return nil
}

func (a *appsCmd) update(c *cli.Context) error {
	appName := c.Args().First()

	app, err := GetAppByName(a.client, appName)
	if err != nil {
		return err
	}

	updatedApp := &models.App{}

	appWithFlags(c, updatedApp)

	err = a.putApp(updatedApp, app.ID)
	if err != nil {
		return err
	}

	fmt.Println("app", appName, "updated")
	return nil
}

func GetAppByName(client *clientv2.Fn, name string) (*models.App, error) {
	appsResp, err := client.Apps.ListApps(&apiapps.ListAppsParams{
		Context: context.Background(),
	})
	if err != nil {
		return nil, err
	}

	var app *models.App
	for i := 0; i < len(appsResp.Payload.Items); i++ {
		if appsResp.Payload.Items[i].Name == name {
			app = appsResp.Payload.Items[i]
		}
	}
	if app == nil {
		return nil, fmt.Errorf("app %s not found", name)
	}

	appResp, err := client.Apps.GetApp(&apiapps.GetAppParams{
		AppID:   app.ID,
		Context: context.Background(),
	})
	if err != nil {
		return nil, err
	}

	return appResp.Payload, nil
}

func (a *appsCmd) setConfig(c *cli.Context) error {
	appName := c.Args().Get(0)
	key := c.Args().Get(1)
	value := c.Args().Get(2)

	app, err := GetAppByName(a.client, appName)
	if err != nil {
		return err
	}

	app.Config = make(map[string]string)
	app.Config[key] = value

	if err := a.putApp(app, app.ID); err != nil {
		return fmt.Errorf("Error updating app configuration: %v", err)
	}

	fmt.Println(appName, "updated", key, "with", value)
	return nil
}

func (a *appsCmd) getConfig(c *cli.Context) error {
	appName := c.Args().Get(0)
	key := c.Args().Get(1)

	app, err := GetAppByName(a.client, appName)
	if err != nil {
		return err
	}

	val, ok := app.Config[key]

	if !ok {
		return fmt.Errorf("Config key does not exist")
	}

	fmt.Println(val)

	return nil
}

func (a *appsCmd) listConfig(c *cli.Context) error {
	appName := c.Args().Get(0)
	app, err := GetAppByName(a.client, appName)

	if err != nil {
		return err
	}

	for key, val := range app.Config {
		fmt.Printf("%s=%s\n", key, val)
	}

	return nil
}

func (a *appsCmd) unsetConfig(c *cli.Context) error {
	appName := c.Args().Get(0)
	key := c.Args().Get(1)

	app, err := GetAppByName(a.client, appName)
	if err != nil {
		return err
	}

	if app.Config[key] == "" {
		return nil
	}
	app.Config[key] = ""

	if err := a.putApp(app, app.ID); err != nil {
		return fmt.Errorf("Error updating app configuration: %v", err)
	}

	fmt.Printf("removed key '%s' from app '%s' \n", key, appName)
	return nil
}

func (a *appsCmd) putApp(app *models.App, appID string) error {
	_, err := a.client.Apps.UpdateApp(&apiapps.UpdateAppParams{
		Context: context.Background(),
		AppID:   appID,
		Body:    app,
	})

	if err != nil {
		switch e := err.(type) {
		case *apiapps.UpdateAppBadRequest:
			return errors.New(e.Payload.Message)
		case *apiapps.UpdateAppNotFound:
			return errors.New(e.Payload.Message)
		default:
			return err
		}
	}

	return nil
}

func (a *appsCmd) inspect(c *cli.Context) error {
	if c.Args().Get(0) == "" {
		return errors.New("Missing app name after the inspect command")
	}

	appName := c.Args().First()
	prop := c.Args().Get(1)
	app, err := GetAppByName(a.client, appName)
	if err != nil {
		return err
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "\t")

	if prop == "" {
		enc.Encode(app)
		return nil
	}

	// TODO: we really need to marshal it here just to
	// unmarshal as map[string]interface{}?
	data, err := json.Marshal(app)
	if err != nil {
		return fmt.Errorf("Could not marshal app: %v", err)
	}
	var inspect map[string]interface{}
	err = json.Unmarshal(data, &inspect)
	if err != nil {
		return fmt.Errorf("Could not unmarshal data: %v", err)
	}

	jq := jsonq.NewQuery(inspect)
	field, err := jq.Interface(strings.Split(prop, ".")...)
	if err != nil {
		return fmt.Errorf("Failed to inspect field %v", prop)
	}
	enc.Encode(field)

	return nil
}

func (a *appsCmd) delete(c *cli.Context) error {
	appName := c.Args().First()
	if appName == "" {
		return errors.New("App name required to delete")
	}

	app, err := GetAppByName(a.client, appName)
	if err != nil {
		return err
	}

	_, err = a.client.Apps.DeleteApp(&apiapps.DeleteAppParams{
		Context: context.Background(),
		AppID:   app.ID,
	})

	if err != nil {
		switch e := err.(type) {
		case *apiapps.DeleteAppNotFound:
			return errors.New(e.Payload.Message)
		}
		return err
	}

	fmt.Println("App", appName, "deleted")
	return nil
}
