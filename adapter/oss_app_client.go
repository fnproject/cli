package adapter

import (
	"context"
	"fmt"
	"github.com/fnproject/cli/common"
	fnclient "github.com/fnproject/fn_go/clientv2"
	oss "github.com/fnproject/fn_go/clientv2"
	apiapps "github.com/fnproject/fn_go/clientv2/apps"
	"github.com/fnproject/fn_go/modelsv2"
	"github.com/urfave/cli"
	"os"
)

type OSSAppClient struct {
	client *oss.Fn
}

func (a *OSSAppClient) CreateApp(c *cli.Context) (*App, error) {
	app := &modelsv2.App{
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
			err = fmt.Errorf("%v", e.Payload.Message)
		case *apiapps.CreateAppConflict:
			err = fmt.Errorf("%v", e.Payload.Message)
		}
		return nil, err
	}

	fmt.Println("Successfully created app: ", resp.Payload.Name)

	return convertV2AppToAdapterApp(resp.Payload), nil
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

func (a *OSSAppClient) GetApp(c *cli.Context) (*App, error) {
	appName := c.Args().Get(0)
	appsResp, err := a.client.Apps.ListApps(&apiapps.ListAppsParams{
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

	return convertV2AppToAdapterApp(app), nil
}

func (a *OSSAppClient) UpdateApp(c *cli.Context) error {
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

func (a *OSSAppClient) DeleteApp(c *cli.Context) error {
	//TODO: call OSS client
	return nil
}

//ListApp returns an array of all apps in the given context and client
func (a *OSSAppClient) ListApp(c *cli.Context) ([]*App, error) {
	params := &apiapps.ListAppsParams{Context: context.Background()}
	var resApps []*App
	for {
		resp, err := a.client.Apps.ListApps(params)
		if err != nil {
			return nil, err
		}

		adapterApps := convertV2AppsToAdapterApps(resp.Payload.Items)
		resApps = append(resApps, adapterApps...)

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

func convertV2AppsToAdapterApps(v2Apps []*modelsv2.App) []*App {
	var resApps []*App

	for _, v2App := range v2Apps {
		resApps = append(resApps, convertV2AppToAdapterApp(v2App))
	}

	return resApps
}

func convertV2AppToAdapterApp(v2App *modelsv2.App) *App {
	resApps := App{Name: v2App.Name, ID: v2App.ID, Annotations: v2App.Annotations, Config: v2App.Config, SyslogURL: v2App.SyslogURL}
	return &resApps
}
