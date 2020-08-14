package oci

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/fnproject/cli/adapter"
	"github.com/go-openapi/strfmt"
	"github.com/oracle/oci-go-sdk/functions"
	"github.com/spf13/viper"
	"os"
	"time"
)

const (
	// AnnotationSubnet - Subnet used to indicate the placement of the function runtime
	AnnotationSubnet = "oracle.com/oci/subnetIds"
)

type AppClient struct {
	client *functions.FunctionsManagementClient
}

func (a AppClient) CreateApp(app *adapter.App) (*adapter.App, error) {
	compartmentId := viper.GetString("oracle.compartment-id")

	subnetIds, err := parseSubnetIds(app.Annotations)

	if err != nil {
		return nil, err
	}
	body := functions.CreateApplicationDetails{
		CompartmentId: &compartmentId,
		Config:        app.Config,
		DisplayName:   &app.Name,
		SubnetIds:     subnetIds,
	}
	req := functions.CreateApplicationRequest{CreateApplicationDetails: body,}

	res, err := a.client.CreateApplication(context.Background(), req)

	if err != nil {
		return nil, err
	}

	adapterApp := convertOCIAppToAdapterApp(&res.Application)
	return adapterApp, nil
}

func parseSubnetIds(annotations map[string]interface{}) ([]string, error) {
	if annotations == nil || len(annotations) == 0 {
		return nil, errors.New("Missing subnets annotation")
	}

	var subnets []string
	subnetsInterface, ok := annotations[AnnotationSubnet]
	if ok {
		// Typecast to byte
		subnetsBytes := subnetsInterface.([]byte)

		err := json.NewDecoder(bytes.NewReader(subnetsBytes)).Decode(&subnets)
		if err != nil {
			return nil, errors.New("Invalid subnets annotation")
		}
	} else {
		return nil, errors.New("Missing subnets annotation")
	}

	return subnets, nil
}

func (a AppClient) GetApp(appName string) (*adapter.App, error) {
	compartmentId := viper.GetString("oracle.compartment-id")
	req := functions.ListApplicationsRequest{CompartmentId: &compartmentId, DisplayName: &appName}
	resp, err := a.client.ListApplications(context.Background(), req)
	if err != nil {
		return nil, err
	}

	if len(resp.Items) > 0 {
		adapterApp := convertOCIAppSummaryToAdapterApp(&resp.Items[0])
		return adapterApp, nil
	} else {
		return nil, adapter.AppNameNotFoundError{Name: appName}
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

	adapterApp := convertOCIAppToAdapterApp(&res.Application)
	return adapterApp, nil
}

func (a AppClient) DeleteApp(appID string) error {
	req := functions.DeleteApplicationRequest{ApplicationId: &appID,}
	_, err := a.client.DeleteApplication(context.Background(), req)
	return err
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
	createdAt, _ := strfmt.ParseDateTime(ociApp.TimeCreated.Format(time.RFC3339Nano))
	updatedAt, _ := strfmt.ParseDateTime(ociApp.TimeUpdated.Format(time.RFC3339Nano))

	annotationMap := make(map[string]interface{})
	annotationMap[AnnotationSubnet] = ociApp.SubnetIds
	annotationMap[AnnotationCompartmentID] = ociApp.CompartmentId

	return &adapter.App{
		Name:        *ociApp.DisplayName,
		ID:          *ociApp.Id,
		Annotations: annotationMap,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}
}

func convertOCIAppToAdapterApp(ociApp *functions.Application) *adapter.App {
	createAt, _ := strfmt.ParseDateTime(ociApp.TimeCreated.Format(time.RFC3339Nano))
	updatedAt, _ := strfmt.ParseDateTime(ociApp.TimeUpdated.Format(time.RFC3339Nano))

	annotationMap := make(map[string]interface{})
	annotationMap[AnnotationSubnet] = ociApp.SubnetIds
	annotationMap[AnnotationCompartmentID] = ociApp.CompartmentId

	return &adapter.App{
		ID:          *ociApp.Id,
		Name:        *ociApp.DisplayName,
		CreatedAt:   createAt,
		UpdatedAt:   updatedAt,
		Annotations: annotationMap,
		Config:      ociApp.Config,
	}
}
