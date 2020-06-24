package adapter

import (
	"context"
	"fmt"
	oss "github.com/fnproject/fn_go/clientv2"
	apiapps "github.com/fnproject/fn_go/clientv2/apps"
	"github.com/fnproject/fn_go/modelsv2"
	"os"
)

type OSSAppClient struct {
	client *oss.Fn
}

func (a OSSAppClient) CreateApp(app *App) (*App, error) {

	resp, err := a.client.Apps.CreateApp(&apiapps.CreateAppParams{
		Context: context.Background(),
		Body:    convertAdapterAppToV2App(app),
	})

	return convertV2AppToAdapterApp(resp.Payload), err
}

func (a OSSAppClient) GetApp(appName string) (*App, error) {
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

func (a OSSAppClient) UpdateApp(app *App) (*App, error) {

	resp, err := a.client.Apps.UpdateApp(&apiapps.UpdateAppParams{
		Context: context.Background(),
		AppID:   app.ID,
		Body:    convertAdapterAppToV2App(app),
	})

	if err != nil {
		switch e := err.(type) {
		case *apiapps.UpdateAppBadRequest:
			err = fmt.Errorf("%s", e.Payload.Message)
		}
		return nil, err
	}

	return convertV2AppToAdapterApp(resp.Payload), nil
}

func (a OSSAppClient) DeleteApp(appID string) error {
	//TODO: call OSS client
	return nil
}

//ListApp returns an array of all apps in the given context and client
func (a OSSAppClient) ListApp(limit int64) ([]*App, error) {
	params := &apiapps.ListAppsParams{Context: context.Background()}
	var resApps []*App
	for {
		resp, err := a.client.Apps.ListApps(params)
		if err != nil {
			return nil, err
		}

		adapterApps := convertV2AppsToAdapterApps(resp.Payload.Items)
		resApps = append(resApps, adapterApps...)

		howManyMore := limit - int64(len(resApps)+len(resp.Payload.Items))
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

func convertAdapterAppToV2App(app *App) *modelsv2.App {
	resApps := modelsv2.App{Name: app.Name, ID: app.ID, Annotations: app.Annotations, Config: app.Config, SyslogURL: app.SyslogURL}
	return &resApps
}
