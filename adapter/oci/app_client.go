package oci

import (
	"context"
	"errors"
	"fmt"
	"github.com/fnproject/cli/adapter"
	"github.com/go-openapi/strfmt"
	"github.com/oracle/oci-go-sdk/functions"
	"github.com/spf13/viper"
	"os"
	"strings"
	"time"
)

const (
	// AnnotationSubnet - Subnet used to indicate the placement of the function runtime
	AnnotationSubnet = "oracle.com/oci/subnetIds"

	// Number of retries when optimistic concurrency fails
	NoEtagMatchRetryCount = 3

	// Error string for No Etag Match
	// See: https://docs.cloud.oracle.com/en-us/iaas/Content/API/References/apierrors.htm
	NoEtagMatchErrorString = "Service error:NoEtagMatch"
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

	return convertOCIAppToAdapterApp(&res.Application)
}

func parseSubnetIds(annotations map[string]interface{}) ([]string, error) {
	if annotations == nil || len(annotations) == 0 {
		return nil, errors.New("Missing subnets annotation")
	}

	var subnets []string
	subnetsInterface, ok := annotations[AnnotationSubnet]
	if !ok {
		return nil, errors.New("Missing subnets annotation")
	}

	// Typecast to []interface{}
	subnetsArray, success := subnetsInterface.([]interface{})

	if !success {
		return nil, errors.New("Invalid subnets annotation")
	}

	for _,s := range subnetsArray {
		// Typecast to string
		subnetString, secondSuccess := s.(string)
		if !secondSuccess {
			return nil, errors.New("Invalid subnets annotation")
		}
		subnets = append(subnets, subnetString)
	}

	return subnets, nil
}

func (a AppClient) GetApp(appName string) (*adapter.App, *string, error) {
	compartmentId := viper.GetString("oracle.compartment-id")
	req := functions.ListApplicationsRequest{CompartmentId: &compartmentId, DisplayName: &appName}
	resp, err := a.client.ListApplications(context.Background(), req)
	if err != nil {
		return nil, nil, err
	}

	if len(resp.Items) > 0 {
		getreq := functions.GetApplicationRequest{ApplicationId: resp.Items[0].Id}
		getres,geterr := a.client.GetApplication(context.Background(), getreq)

		if geterr != nil {
			return nil, nil, geterr
		}

		app, err := convertOCIAppToAdapterApp(&getres.Application)
		return app, getres.Etag, err
	} else {
		return nil, nil, adapter.AppNameNotFoundError{Name: appName}
	}
}

func mergeConfig(config map[string]string) {
	for k, v := range config {
		if v == "" {
			delete(config, k)
		} else {
			config[k] = v
		}
	}
}

func (a AppClient) UpdateApp(app *adapter.App, lock *string) (*adapter.App, error) {
	mergeConfig(app.Config)

	body := functions.UpdateApplicationDetails{
		Config: app.Config,
	}

	req := functions.UpdateApplicationRequest{UpdateApplicationDetails: body, ApplicationId: &app.ID, IfMatch: lock}
	res, err := a.client.UpdateApplication(context.Background(), req)

	if err != nil {
		return nil, err
	}

	return convertOCIAppToAdapterApp(&res.Application)
}

func (a AppClient) HandleRetry(atomicOperation func() (*adapter.App, error)) (*adapter.App, error) {
	var app *adapter.App
	var err error

	for i:= 0; i < NoEtagMatchRetryCount; i++ {
		app, err = atomicOperation()

		if err == nil || !strings.Contains(err.Error(), NoEtagMatchErrorString) {
			// Break here and do not retry if there is no error or if error is not `NoEtagMatch`
			// See: https://docs.cloud.oracle.com/en-us/iaas/Content/API/References/apierrors.htm
			break
		}
	}

	if err != nil {
		return nil, err
	}

	return app, nil
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

		adapterApps, err := convertOCIAppsToAdapterApps(&resp.Items)
		if err != nil {
			return nil, err
		}

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

func convertOCIAppsToAdapterApps(ociApps *[]functions.ApplicationSummary) ([]*adapter.App, error) {
	var resApps []*adapter.App
	for _, ociApp := range *ociApps {
		app, err := convertOCIAppSummaryToAdapterApp(&ociApp)
		if err != nil {
			return nil, err
		}
		resApps = append(resApps, app)
	}

	return resApps, nil
}

func convertOCIAppSummaryToAdapterApp(ociApp *functions.ApplicationSummary) (*adapter.App, error) {
	createdAt, err := strfmt.ParseDateTime(ociApp.TimeCreated.Format(time.RFC3339Nano))
	if err != nil {
		return nil, errors.New("missing or invalid TimeCreated in application")
	}

	updatedAt, err := strfmt.ParseDateTime(ociApp.TimeUpdated.Format(time.RFC3339Nano))
	if err != nil {
		return nil, errors.New("missing or invalid TimeUpdated in application")
	}

	annotationMap := make(map[string]interface{})
	annotationMap[AnnotationSubnet] = ociApp.SubnetIds
	annotationMap[AnnotationCompartmentID] = ociApp.CompartmentId

	return &adapter.App{
		Name:        *ociApp.DisplayName,
		ID:          *ociApp.Id,
		Annotations: annotationMap,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Config:		 ociApp.FreeformTags,
	}, nil
}

func convertOCIAppToAdapterApp(ociApp *functions.Application) (*adapter.App, error) {
	createAt, err := strfmt.ParseDateTime(ociApp.TimeCreated.Format(time.RFC3339Nano))
	if err != nil {
		return nil, errors.New("missing or invalid TimeCreated in application")
	}

	updatedAt, err := strfmt.ParseDateTime(ociApp.TimeUpdated.Format(time.RFC3339Nano))
	if err != nil {
		return nil, errors.New("missing or invalid TimeUpdated in application")
	}

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
	}, nil
}
