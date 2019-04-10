package app

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"text/tabwriter"

	"context"
	"strings"

	"github.com/fnproject/cli/client"
	"github.com/fnproject/cli/common"
	fnclient "github.com/fnproject/fn_go/clientv2"
	apiapps "github.com/fnproject/fn_go/clientv2/apps"
	apifns "github.com/fnproject/fn_go/clientv2/fns"
	apitriggers "github.com/fnproject/fn_go/clientv2/triggers"
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
	resApps, err := getApps(c, a.client)
	if err != nil {
		return err
	}
	return printApps(c, resApps)
}

// getApps returns an array of all apps in the given context and client
func getApps(c *cli.Context, client *fnclient.Fn) ([]*modelsv2.App, error) {
	params := &apiapps.ListAppsParams{Context: context.Background()}
	var resApps []*modelsv2.App
	for {
		resp, err := client.Apps.ListApps(params)
		if err != nil {
			return nil, err
		}

		resApps = append(resApps, resp.Payload.Items...)

		n := c.Int64("n")

		howManyMore := n - int64(len(resApps)+len(resp.Payload.Items))
		if howManyMore <= 0 || resp.Payload.NextCursor == "" {
			break
		}

		params.Cursor = &resp.Payload.NextCursor
	}

	if len(resApps) == 0 {
		fmt.Fprint(os.Stderr, "No apps found\n")
		return nil, nil
	}
	return resApps, nil
}

// BashCompleteApps can be called from a BashComplete function
// to provide app completion suggestions (Does not check if the
// current context already contains an app name as an argument.
// This should be checked before calling this)
func BashCompleteApps(c *cli.Context) {
	provider, err := client.CurrentProvider()
	if err != nil {
		return
	}
	resp, err := getApps(c, provider.APIClientv2())
	if err != nil {
		return
	}
	for _, r := range resp {
		fmt.Println(r.Name)
	}
}

func appWithFlags(c *cli.Context, app *modelsv2.App) {
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
	app := &modelsv2.App{
		Name: c.Args().Get(0),
	}

	appWithFlags(c, app)

	_, err := CreateApp(a.client, app)
	return err
}

// CreateApp creates a new app using the given client
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
	appName := c.Args().First()

	app, err := GetAppByName(a.client, appName)
	if err != nil {
		return err
	}

	appWithFlags(c, app)

	if _, err = PutApp(a.client, app.ID, app); err != nil {
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

	if c.Bool("f") {

		fns, triggers, err := getAssociatedResources(c, a.client, app)
		if err != nil {
			return fmt.Errorf("Failed to get associated objects: %s", err)
		}
		if userConfirmedRecursiveAppDeleteion(fns, triggers) {
			deleteTriggers(c, a.client, triggers)
			deleteFunctions(c, a.client, fns)
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
		Name:    &appName,
	})
	if err != nil {
		return nil, err
	}

	var app *modelsv2.App
	if len(appsResp.Payload.Items) > 0 {
		app = appsResp.Payload.Items[0]
	} else {
		return nil, NameNotFoundError{appName}
	}

	return app, nil
}

func getAssociatedResources(c *cli.Context, client *fnclient.Fn, app *modelsv2.App) ([]*modelsv2.Fn, []*modelsv2.Trigger, error) {
	fns, err := listFnsInApp(c, client, app)
	if err != nil {
		return nil, nil, err
	}
	var trs []*modelsv2.Trigger
	for _, fn := range fns {
		t, err := listTriggersInFunc(c, client, fn)
		if err != nil {
			return nil, nil, err
		}
		trs = append(trs, t...)
	}
	return fns, trs, nil
}

func userConfirmedRecursiveAppDeleteion(fns []*modelsv2.Fn, trs []*modelsv2.Trigger) bool {
	if len(fns) > 0 {
		fmt.Println("You are about to delete the following resources:")
		fmt.Println("   Applications: 1")
		fmt.Println("   Functions:   ", len(fns))
		fmt.Println("   Triggers:    ", len(trs))
		fmt.Println("This operation cannot be undone")
		fmt.Printf("Do you wish to proceed? Y/N: ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		if strings.ToLower(input) == "y\n" {
			return true
		} else if strings.ToLower(input) == "n\n" {
			fmt.Println("Cancelling delete")
			return false
		} else {
			fmt.Println("Unrecognised input, should be Y/N")
			fmt.Println("Cancelling delete")
			return false
		}
	}
	return true
}

func deleteFunctions(c *cli.Context, client *fnclient.Fn, fns []*modelsv2.Fn) error {
	if fns == nil {
		return nil
	}
	for _, fn := range fns {
		params := apifns.NewDeleteFnParams()
		params.FnID = fn.ID
		_, err := client.Fns.DeleteFn(params)
		if err != nil {
			return fmt.Errorf("Failed to delete Function %s: %s", fn.Name, err)
		}
		fmt.Println("Function ", fn.Name, " deleted")
	}
	return nil
}

func deleteTriggers(c *cli.Context, client *fnclient.Fn, triggers []*modelsv2.Trigger) error {
	if triggers == nil {
		return nil
	}
	for _, t := range triggers {
		params := apitriggers.NewDeleteTriggerParams()
		params.TriggerID = t.ID
		_, err := client.Triggers.DeleteTrigger(params)
		if err != nil {
			return fmt.Errorf("Failed to Delete trigger %s: %s", t.Name, err)
		}
		fmt.Println("Trigger ", t.Name, " deleted")
	}
	return nil
}

func listFnsInApp(c *cli.Context, client *fnclient.Fn, app *modelsv2.App) ([]*modelsv2.Fn, error) {
	params := &apifns.ListFnsParams{
		Context: context.Background(),
		AppID:   &app.ID,
	}

	var resFns []*modelsv2.Fn
	for {
		resp, err := client.Fns.ListFns(params)

		if err != nil {
			return nil, fmt.Errorf("Could not list functions in App %s: %s", app.Name, err)
		}
		n := c.Int64("n")

		resFns = append(resFns, resp.Payload.Items...)
		howManyMore := n - int64(len(resFns)+len(resp.Payload.Items))
		if howManyMore <= 0 || resp.Payload.NextCursor == "" {
			break
		}

		params.Cursor = &resp.Payload.NextCursor
	}

	return resFns, nil
}

func listTriggersInFunc(c *cli.Context, client *fnclient.Fn, fn *modelsv2.Fn) ([]*modelsv2.Trigger, error) {
	params := &apitriggers.ListTriggersParams{
		Context: context.Background(),
		AppID:   &fn.AppID,
		FnID:    &fn.ID,
	}

	var resTriggers []*modelsv2.Trigger
	for {
		resp, err := client.Triggers.ListTriggers(params)
		if err != nil {
			return nil, fmt.Errorf("Could not list Triggers in Function %s: %s", fn.Name, err)
		}
		n := c.Int64("n")

		resTriggers = append(resTriggers, resp.Payload.Items...)
		howManyMore := n - int64(len(resTriggers)+len(resp.Payload.Items))
		if howManyMore <= 0 || resp.Payload.NextCursor == "" {
			break
		}
		params.Cursor = &resp.Payload.NextCursor
	}
	return resTriggers, nil
}
