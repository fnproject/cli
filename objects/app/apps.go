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
	provider         provider.Provider
	providerAdapter  adapter.Provider
	client           *fnclient.Fn
	apiClientAdapter adapter.APIClient
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
	n := c.Int64("n")
	resApps, err := a.apiClientAdapter.AppClient().ListApp(n)
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
	n := c.Int64("n")
	resp, err := providerAdapter.APIClient().AppClient().ListApp(n)
	if err != nil {
		return
	}
	for _, r := range resp {
		fmt.Println(r.Name)
	}
}


func appWithFlags(c *cli.Context, app *adapter.App) {
	if c.IsSet("syslog-url") {
		str := c.String("syslog-url")
		app.SyslogURL = &str
	}
	if len(c.StringSlice("config")) > 0 {
		app.Config = common.ExtractConfig(c.StringSlice("config"))
	}
	if len(c.StringSlice("annotation")) > 0 {
		app.Annotations = common.ExtractAnnotations(c)
	}
}

func (a *appsCmd) create(c *cli.Context) error {
	app := &adapter.App{
		Name: c.Args().Get(0),
	}

	appWithFlags(c, app)

	_, err := CreateApp(a.apiClientAdapter.AppClient(), app)
	return err
}

// CreateApp creates a new app using the given client
func CreateApp(a adapter.AppClient, app *adapter.App) (*adapter.App, error) {
	resp, err := a.CreateApp(app)

	if err != nil {
		switch e := err.(type) {
		case *apiapps.CreateAppBadRequest:
			err = fmt.Errorf("%v", e.Payload.Message)
		case *apiapps.CreateAppConflict:
			err = fmt.Errorf("%v", e.Payload.Message)
		}
		return nil, err
	}

	fmt.Println("Successfully created app: ", resp.Name)
	return resp, nil
}

func (a *appsCmd) update(c *cli.Context) error {
	appName := c.Args().First()

	app, err := a.apiClientAdapter.AppClient().GetApp(appName)
	if err != nil {
		return err
	}

	appWithFlags(c, app)

	if _, err = a.apiClientAdapter.AppClient().UpdateApp(app); err != nil {
		return err
	}

	fmt.Println("app", appName, "updated")
	return nil
}

func (a *appsCmd) setConfig(c *cli.Context) error {
	appName := c.Args().Get(0)
	key := c.Args().Get(1)
	value := c.Args().Get(2)

	app, err := a.apiClientAdapter.AppClient().GetApp(appName)
	if err != nil {
		return err
	}

	app.Config = make(map[string]string)
	app.Config[key] = value

	if _, err = a.apiClientAdapter.AppClient().UpdateApp(app); err != nil {
		return fmt.Errorf("Error updating app configuration: %v", err)
	}

	fmt.Println(appName, "updated", key, "with", value)
	return nil
}

func (a *appsCmd) getConfig(c *cli.Context) error {
	appName := c.Args().Get(0)
	key := c.Args().Get(1)

	app, err := a.apiClientAdapter.AppClient().GetApp(appName)
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

	app, err := a.apiClientAdapter.AppClient().GetApp(appName)
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

	app, err := a.apiClientAdapter.AppClient().GetApp(appName)
	if err != nil {
		return err
	}
	_, ok := app.Config[key]
	if !ok {
		fmt.Printf("Config key '%s' does not exist. Nothing to do.\n", key)
		return nil
	}
	app.Config[key] = ""

	_, err = a.apiClientAdapter.AppClient().UpdateApp(app)
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

	appName := c.Args().First()
	prop := c.Args().Get(1)

	app, err := a.apiClientAdapter.AppClient().GetApp(appName)
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
