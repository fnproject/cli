package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"context"
	"strings"

	"github.com/fnproject/cli/common"
	fnclient "github.com/fnproject/fn_go/clientv2"
	apiapps "github.com/fnproject/fn_go/clientv2/apps"
	"github.com/fnproject/fn_go/modelsv2"
	"github.com/fnproject/fn_go/provider"
	"github.com/jmoiron/jsonq"
	"github.com/urfave/cli"
)

type appsCmd struct {
	provider provider.Provider
	client   *fnclient.Fn
}

func printApps(c *cli.Context, apps []*modelsv2.App) error {
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
	params := &apiapps.ListAppsParams{Context: context.Background()}
	var resApps []*modelsv2.App
	for {
		resp, err := a.client.Apps.ListApps(params)
		if err != nil {
			return err
		}

		resApps = append(resApps, resp.Payload.Items...)

		n := c.Int64("n")
		if n < 0 {
			return errors.New("Number of calls: negative value not allowed")
		}

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

	if err := printApps(c, resApps); err != nil {
		return err
	}
	return nil
}

func appWithFlags(c *cli.Context, app *modelsv2.App) {
	if len(c.String("syslog-url")) > 0 {
		app.SyslogURL = c.String("syslog-url")
	}
	if len(c.StringSlice("config")) > 0 {
		app.Config = common.ExtractEnvConfig(c.StringSlice("config"))
	}
	if len(c.StringSlice("annotation")) > 0 {
		app.Annotations = common.ExtractAnnotations(c)
	}
}

func (a *appsCmd) create(c *cli.Context) error {
	app := &modelsv2.App{
		Name: c.Args().Get(0),
	}

	appWithFlags(c, app)

	return CreateApp(a.client, app)
}

// CreateApp creates a new app using the given client
func CreateApp(a *fnclient.Fn, app *modelsv2.App) error {
	resp, err := a.Apps.CreateApp(&apiapps.CreateAppParams{
		Context: context.Background(),
		Body:    app,
	})

	if err != nil {
		switch e := err.(type) {
		case *apiapps.CreateAppBadRequest:
			return fmt.Errorf("%v", e.Payload.Message)
		case *apiapps.CreateAppConflict:
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

	appWithFlags(c, app)

	if err = PutApp(a.client, app.ID, app); err != nil {
		return err
	}

	fmt.Println("app", appName, "updated")
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

	if err = PutApp(a.client, app.ID, app); err != nil {
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
	app.Config[key] = ""

	err = PutApp(a.client, app.ID, app)
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
		//return errors.New("App name required to delete")
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

// PutApp updates the app with the given ID using the content of the provided app
func PutApp(a *fnclient.Fn, appID string, app *modelsv2.App) error {
	_, err := a.Apps.UpdateApp(&apiapps.UpdateAppParams{
		Context: context.Background(),
		AppID:   appID,
		Body:    app,
	})

	if err != nil {
		switch e := err.(type) {
		case *apiapps.UpdateAppBadRequest:
			return fmt.Errorf("%s", e.Payload.Message)

		default:
			return err
		}
	}

	return nil
}

// NameNotFoundError error for app not found when looked up by name
type NameNotFoundError struct {
	Name string
}

func (n NameNotFoundError) Error() string {
	return fmt.Sprintf("app %s not found", n.Name)
}

// GetAppByName looks up an app by name using the given client
func GetAppByName(client *fnclient.Fn, appName string) (*modelsv2.App, error) {
	appsResp, err := client.Apps.ListApps(&apiapps.ListAppsParams{
		Context: context.Background(),
	})
	if err != nil {
		return nil, err
	}

	var app *modelsv2.App
	for i := 0; i < len(appsResp.Payload.Items); i++ {
		if appsResp.Payload.Items[i].Name == appName {
			app = appsResp.Payload.Items[i]
		}
	}
	if app == nil {
		return nil, NameNotFoundError{appName}
	}

	return app, nil
}
