package oss

import (
	"context"
	"errors"
	"fmt"
	"github.com/fnproject/cli/adapter"
	oss "github.com/fnproject/fn_go/clientv2"
	apiapps "github.com/fnproject/fn_go/clientv2/apps"
	"github.com/fnproject/fn_go/modelsv2"
	"os"
)

type AppClient struct {
	client *oss.Fn
}

func (a AppClient) CreateApp(app *adapter.App) (*adapter.App, error) {
	resp, err := a.client.Apps.CreateApp(&apiapps.CreateAppParams{
		Context: context.Background(),
		Body:    convertAdapterAppToV2App(app),
	})

	if err != nil {
		return nil, err
	}
	return convertV2AppToAdapterApp(resp.Payload), err
}

func (a AppClient) GetApp(appName string) (*adapter.App, *string, error) {
	appsResp, err := a.client.Apps.ListApps(&apiapps.ListAppsParams{
		Context: context.Background(),
		Name:    &appName,
	})
	if err != nil {
		return nil, nil, err
	}
	var app *modelsv2.App
	if len(appsResp.Payload.Items) > 0 {
		app = appsResp.Payload.Items[0]
	} else {
		return nil, nil, adapter.AppNameNotFoundError{ Name: appName}
	}
	return convertV2AppToAdapterApp(app), nil, nil
}

func (a AppClient) UpdateApp(app *adapter.App, lock *string) (*adapter.App, error) {
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

func (a AppClient) HandleRetry(atomicOperation func() (*adapter.App, error)) (*adapter.App, error) {
	// No retry handling in the Open Source client
	return atomicOperation()
}

func (a AppClient) DeleteApp(appID string) error {
	_, err := a.client.Apps.DeleteApp(&apiapps.DeleteAppParams{
		Context: context.Background(),
		AppID:   appID,
	})

	if err != nil {
		switch e := err.(type) {
		case *apiapps.DeleteAppNotFound:
			return errors.New(e.Payload.Message)
		}
		return err
	}

	return nil
}

//ListApp returns an array of all apps in the given context and client
func (a AppClient) ListApp(limit int64) ([]*adapter.App, error) {
	params := &apiapps.ListAppsParams{Context: context.Background()}
	var resApps []*adapter.App
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

func convertV2AppsToAdapterApps(v2Apps []*modelsv2.App) []*adapter.App {
	var resApps []*adapter.App
	for _, v2App := range v2Apps {
		resApps = append(resApps, convertV2AppToAdapterApp(v2App))
	}
	return resApps
}

func convertV2AppToAdapterApp(v2App *modelsv2.App) *adapter.App {
	resApps := adapter.App{Name: v2App.Name, ID: v2App.ID, Annotations: v2App.Annotations, Config: v2App.Config, SyslogURL: v2App.SyslogURL}
	return &resApps
}

func convertAdapterAppToV2App(app *adapter.App) *modelsv2.App {
	resApps := modelsv2.App{Name: app.Name, ID: app.ID, Annotations: app.Annotations, Config: app.Config, SyslogURL: app.SyslogURL}
	return &resApps
}
