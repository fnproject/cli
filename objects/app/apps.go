package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fnproject/cli/adapter"
	"os"
	"text/tabwriter"

	"context"
	"strings"

	"github.com/fnproject/cli/client"
	"github.com/fnproject/cli/common"

	fnclient "github.com/fnproject/fn_go/clientv2"
	apiapps "github.com/fnproject/fn_go/clientv2/apps"
	"github.com/fnproject/fn_go/modelsv2"
	"github.com/fnproject/fn_go/provider"
	"github.com/jmoiron/jsonq"
	"github.com/urfave/cli"
)

type appsCmd struct {
	provider        provider.Provider
	providerAdapter adapter.ProviderAdapter
	client          *fnclient.Fn
	clientAdapter   adapter.ClientAdapter
}

func printApps(c *cli.Context, apps []*adapter.App) error {
	outputFormat := strings.ToLower(c.String("output"))
	if outputFormat == "json" {
		var allApps []interface{}
		for _, app := range apps {
			a := struct {
				Name string `json:"name"`
				ID   string `json:"id"`
			}{app.Name,
				app.ID,
			}
			allApps = append(allApps, a)
		}
		b, err := json.MarshalIndent(allApps, "", "    ")
		if err != nil {
			return err
		}
		fmt.Fprint(os.Stdout, string(b))
	} else {
		w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
		fmt.Fprint(w, "NAME", "\t", "ID", "\t", "\n")
		for _, app := range apps {
			fmt.Fprint(w, app.Name, "\t", app.ID, "\t", "\n")

		}
		if err := w.Flush(); err != nil {
			return err
		}
	}
	return nil
}

func (a *appsCmd) list(c *cli.Context) error {
	resApps, err := a.clientAdapter.GetAppsClient().ListApp(c)
	if err != nil {
		return err
	}
	return printApps(c, resApps)
}


// BashCompleteApps can be called from a BashComplete function
// to provide app completion suggestions (Does not check if the
// current context already contains an app name as an argument.
// This should be checked before calling this)
func BashCompleteApps(c *cli.Context) {
	providerAdapter, err := client.CurrentProviderAdapter()
	if err != nil {
		return
	}
	resp, err := providerAdapter.GetClientAdapter().GetAppsClient().ListApp(c)
	if err != nil {
		return
	}
	for _, r := range resp {
		fmt.Println(r.Name)
	}
}


func (a *appsCmd) create(c *cli.Context) error {
	_, err := a.clientAdapter.GetAppsClient().CreateApp(c)
	return err
}

// CreateApp creates a new app using the given client
// This is used in deploy. To be removed
func CreateApp(a *fnclient.Fn, app *modelsv2.App) (*modelsv2.App, error) {
	resp, err := a.Apps.CreateApp(&apiapps.CreateAppParams{
		Context: context.Background(),
		Body:    app,
	})

	if err != nil {
		switch e := err.(type) {
		case *apiapps.CreateAppBadRequest:
			err = fmt.Errorf("%v", e.Payload.Message)
		case *apiapps.CreateAppConflict:
			err = fmt.Errorf("%v", e.Payload.Message)
		}
		return nil, err
	}

	fmt.Println("Successfully created app: ", resp.Payload.Name)
	return resp.Payload, nil
}

func (a *appsCmd) update(c *cli.Context) error {
	a.clientAdapter.GetAppsClient().UpdateApp(c)
	return nil
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

	if _, err = PutApp(a.client, app.ID, app); err != nil {
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
		return fmt.Errorf("config key does not exist")
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

	if len(app.Config) == 0 {
		fmt.Fprintf(os.Stderr, "No config found for app: %s\n", appName)
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprint(w, "KEY", "\t", "VALUE", "\n")
	for key, val := range app.Config {
		fmt.Fprint(w, key, "\t", val, "\n")
	}
	w.Flush()

	return nil
}

func (a *appsCmd) unsetConfig(c *cli.Context) error {
	appName := c.Args().Get(0)
	key := c.Args().Get(1)

	app, err := GetAppByName(a.client, appName)
	if err != nil {
		return err
	}
	_, ok := app.Config[key]
	if !ok {
		fmt.Printf("Config key '%s' does not exist. Nothing to do.\n", key)
		return nil
	}
	app.Config[key] = ""

	_, err = PutApp(a.client, app.ID, app)
	if err != nil {
		return err
	}

	fmt.Printf("Removed key '%s' from app '%s' \n", key, appName)
	return nil
}

func (a *appsCmd) inspect(c *cli.Context) error {
	if c.Args().Get(0) == "" {
		return errors.New("Missing app name after the inspect command")
	}

	prop := c.Args().Get(1)

	app, err := a.clientAdapter.GetAppsClient().GetApp(c)
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
		//return errors.New("App name required to delete")
	}

	app, err := GetAppByName(a.client, appName)
	if err != nil {
		return err
	}

	//recursive delete of sub-objects
	if c.Bool("recursive") {
		fns, triggers, err := common.ListFnsAndTriggersInApp(c, a.client, app)
		if err != nil {
			return fmt.Errorf("Failed to get associated objects: %s", err)
		}

		//Forced deletion
		var shouldContinue bool
		if c.Bool("force") {
			shouldContinue = true
		} else {
			shouldContinue = common.UserConfirmedMultiResourceDeletion([]*modelsv2.App{app}, fns, triggers)
		}

		if shouldContinue {
			err := common.DeleteTriggers(c, a.client, triggers)
			if err != nil {
				return fmt.Errorf("Failed to delete associated objects: %s", err)
			}
			err = common.DeleteFunctions(c, a.client, fns)
			if err != nil {
				return fmt.Errorf("Failed to delete associated objects: %s", err)
			}
		} else {
			return nil
		}
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

// PutApp updates the app with the given ID using the content of the provided app
func PutApp(a *fnclient.Fn, appID string, app *modelsv2.App) (*modelsv2.App, error) {
	resp, err := a.Apps.UpdateApp(&apiapps.UpdateAppParams{
		Context: context.Background(),
		AppID:   appID,
		Body:    app,
	})

	if err != nil {
		switch e := err.(type) {
		case *apiapps.UpdateAppBadRequest:
			err = fmt.Errorf("%s", e.Payload.Message)
		}
		return nil, err
	}

	return resp.Payload, nil
}

// GetAppByName looks up an app by name using the given client
// Used in many other places. To be removed
func GetAppByName(client *fnclient.Fn, appName string) (*modelsv2.App, error) {
	appsResp, err := client.Apps.ListApps(&apiapps.ListAppsParams{
		Context: context.Background(),
		Name:    &appName,
	})
	if err != nil {
		return nil, err
	}

	var app *modelsv2.App
	if len(appsResp.Payload.Items) > 0 {
		app = appsResp.Payload.Items[0]
	} else {
		return nil, adapter.NameNotFoundError{appName}
	}

	return app, nil
}
