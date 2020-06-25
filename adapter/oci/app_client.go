package oci

import (
	"context"
	"fmt"
	"github.com/fnproject/cli/adapter"
	"github.com/oracle/oci-go-sdk/functions"
	"github.com/spf13/viper"
	"os"
)

type AppClient struct {
	client *functions.FunctionsManagementClient
}

func (a AppClient) CreateApp(app *adapter.App) (*adapter.App, error) {
	//TODO: call OCI client
	return nil, nil
}

func (a AppClient) GetApp(appName string) (*adapter.App, error) {
	//TODO: call OCI client
	return nil, nil
}

func (a AppClient) UpdateApp(app *adapter.App) (*adapter.App, error) {
	//TODO: call OCI client
	return nil, nil
}

func (a AppClient) DeleteApp(appID string) error {
	//TODO: call OCI client
	return nil
}

func (a AppClient) ListApp(limit int64) ([]*adapter.App, error) {
	compartmentId := viper.GetString("oracle.compartment-id")
	var resApps []*adapter.App
	req := functions.ListApplicationsRequest{CompartmentId: &compartmentId,}

	for {
		resp, err := a.client.ListApplications(context.Background(), req)
		if err != nil {
			return nil, err
		}

		adapterApps := convertOCIAppsToAdapterApps(&resp.Items)
		resApps = append(resApps, adapterApps...)
		howManyMore := limit - int64(len(resApps)+len(resp.Items))

		if howManyMore <= 0 || resp.OpcNextPage == nil {
			break
		}

		req.Page = resp.OpcNextPage
	}

	if len(resApps) == 0 {
		fmt.Fprint(os.Stderr, "No apps found\n")
		return nil, nil
	}

	return resApps, nil
}

func convertOCIAppsToAdapterApps(ociApps *[]functions.ApplicationSummary) []*adapter.App {
	var resApps []*adapter.App
	for _, ociApp := range *ociApps {
		app := adapter.App{Name: *ociApp.DisplayName, ID: *ociApp.Id}
		resApps = append(resApps, &app)
	}

	return resApps
}
