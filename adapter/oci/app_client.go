package oci

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/fnproject/cli/adapter"
	"github.com/go-openapi/strfmt"
	"github.com/oracle/oci-go-sdk/functions"
	"github.com/spf13/viper"
	"os"
)

type AppClient struct {
	client *functions.FunctionsManagementClient
}

func (a AppClient) CreateApp(app *adapter.App) (*adapter.App, error) {
	compartmentId := viper.GetString("oracle.compartment-id")

	body := functions.CreateApplicationDetails{
		CompartmentId: &compartmentId,
		Config: app.Config,
		DisplayName: &app.Name,
		SubnetIds: extractSubnetIds(app.Annotations),
	}
	req := functions.CreateApplicationRequest{CreateApplicationDetails: body,}

	res,err := a.client.CreateApplication(context.Background(), req)

	if err != nil {
		return nil, err
	}

	adapterApp := convertOCIAppTpAdapterApp(&res.Application)
	return adapterApp, nil
	}

func extractSubnetIds(Annotations map[string]interface{}) []string {
	if len(Annotations) == 0 {
		return nil
	}

	var subnets []string
	subnetsInterface, ok := Annotations["oracle.com/oci/subnetIds"]
	if ok {
		// Typecast to byte
		subnetsBytes := subnetsInterface.([]byte)

		err := json.NewDecoder(bytes.NewReader(subnetsBytes)).Decode(&subnets)
		if err != nil {
			return nil
		}
	}

	return subnets
}

func (a AppClient) GetApp(appName string) (*adapter.App, error) {
	compartmentId := viper.GetString("oracle.compartment-id")
	req := functions.ListApplicationsRequest{CompartmentId: &compartmentId,}
	resp, err := a.client.ListApplications(context.Background(), req)
	if err != nil {
		return nil, err
	}

	if len(resp.Items) > 0 {
		adapterApp := convertOCIAppSummaryToAdapterApp(&resp.Items[0])
		return adapterApp, nil
	} else {
		return nil, adapter.NameNotFoundError{ Name: appName}
	}
}

func (a AppClient) UpdateApp(app *adapter.App) (*adapter.App, error) {
	body := functions.UpdateApplicationDetails{
		Config: app.Config,
	}

	req := functions.UpdateApplicationRequest{UpdateApplicationDetails: body,}
	res, err := a.client.UpdateApplication(context.Background(), req)

	if err != nil {
		return nil, err
	}

	adapterApp := convertOCIAppTpAdapterApp(&res.Application)
	return adapterApp, nil
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
		app := convertOCIAppSummaryToAdapterApp(&ociApp)
		resApps = append(resApps, app)
	}

	return resApps
}

func convertOCIAppSummaryToAdapterApp(ociApp *functions.ApplicationSummary) *adapter.App {
	annotationMap := make(map[string]interface{})
	annotationMap["oracle.com/oci/subnetIds"] = ociApp.SubnetIds
	createdAt,_ := strfmt.ParseDateTime(ociApp.TimeCreated.String())
	updatedAt,_ := strfmt.ParseDateTime(ociApp.TimeUpdated.String())
	return &adapter.App{
		Name: *ociApp.DisplayName,
		ID: *ociApp.Id,
		Annotations: annotationMap,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

func convertOCIAppTpAdapterApp(ociApp *functions.Application) *adapter.App {
	createAt,_ := strfmt.ParseDateTime(ociApp.TimeCreated.String())
	updatedAt,_ := strfmt.ParseDateTime(ociApp.TimeUpdated.String())
	annotationMap := make(map[string]interface{})
	annotationMap["oracle.com/oci/subnetIds"] = ociApp.SubnetIds

	return &adapter.App{
		ID: *ociApp.Id,
		Name: *ociApp.DisplayName,
		CreatedAt: createAt,
		UpdatedAt: updatedAt,
		Annotations: annotationMap,
		Config: ociApp.Config,
	}
}