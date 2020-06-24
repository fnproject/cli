package adapter

import (
	"context"
	"fmt"
	"github.com/oracle/oci-go-sdk/functions"
	"github.com/spf13/viper"
	"os"
)

type OCIAppClient struct {
	client *functions.FunctionsManagementClient
}

func (a OCIAppClient) CreateApp(app *App) (*App, error) {
	//TODO: call OCI client
	return nil, nil
}

func (a OCIAppClient) GetApp(appName string) (*App, error) {
	//TODO: call OCI client
	return nil, nil
}

func (a OCIAppClient) UpdateApp(app *App) (*App, error) {
	//TODO: call OCI client
	return nil, nil
}

func (a OCIAppClient) DeleteApp(appID string) error {
	//TODO: call OCI client
	return nil
}

func (a OCIAppClient) ListApp(limit int64) ([]*App, error) {
	compartmentId := viper.GetString("oracle.compartment-id")
	var resApps []*App
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

func convertOCIAppsToAdapterApps(ociApps *[]functions.ApplicationSummary) []*App {
	var resApps []*App

	for _, ociApp := range *ociApps {
		app := App{Name: *ociApp.DisplayName, ID: *ociApp.Id}
		resApps = append(resApps, &app)
	}

	return resApps
}