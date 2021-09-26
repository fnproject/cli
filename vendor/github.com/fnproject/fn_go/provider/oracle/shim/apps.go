package shim

import (
	"fmt"
	"github.com/fnproject/fn_go/clientv2/apps"
	"github.com/fnproject/fn_go/modelsv2"
	"github.com/fnproject/fn_go/provider/oracle/shim/client"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/oracle/oci-go-sdk/v48/functions"
)

const (
	annotationSubnet = "oracle.com/oci/subnetIds"
)

type appsShim struct {
	ociClient     client.FunctionsManagementClient
	compartmentId string
}

var _ apps.ClientService = &appsShim{}

func NewAppsShim(ociClient client.FunctionsManagementClient, compartmentId string) apps.ClientService {
	return &appsShim{ociClient: ociClient, compartmentId: compartmentId}
}

func (s *appsShim) CreateApp(params *apps.CreateAppParams) (*apps.CreateAppOK, error) {
	subnetIds, err := parseSubnetIds(params.Body.Annotations)
	if err != nil {
		return nil, err
	}

	details := functions.CreateApplicationDetails{
		CompartmentId: &s.compartmentId,
		DisplayName:   &params.Body.Name,
		SubnetIds:     subnetIds,
		Config:        params.Body.Config,
		SyslogUrl:     params.Body.SyslogURL,
	}

	req := functions.CreateApplicationRequest{CreateApplicationDetails: details}

	res, err := s.ociClient.CreateApplication(ctxOrBackground(params.Context), req)
	if err != nil {
		return nil, err
	}

	return &apps.CreateAppOK{
		Payload: ociAppToV2(res.Application),
	}, nil
}

func (s *appsShim) DeleteApp(params *apps.DeleteAppParams) (*apps.DeleteAppNoContent, error) {
	req := functions.DeleteApplicationRequest{ApplicationId: &params.AppID}

	_, err := s.ociClient.DeleteApplication(ctxOrBackground(params.Context), req)
	if err != nil {
		return nil, err
	}

	return &apps.DeleteAppNoContent{}, nil
}

func (s *appsShim) GetApp(params *apps.GetAppParams) (*apps.GetAppOK, error) {
	req := functions.GetApplicationRequest{ApplicationId: &params.AppID}

	res, err := s.ociClient.GetApplication(ctxOrBackground(params.Context), req)
	if err != nil {
		return nil, err
	}

	return &apps.GetAppOK{
		Payload: ociAppToV2(res.Application),
	}, nil
}

func (s *appsShim) ListApps(params *apps.ListAppsParams) (*apps.ListAppsOK, error) {
	var limit *int
	if params.PerPage != nil {
		ppInt := int(*params.PerPage)
		limit = &ppInt
	}

	req := functions.ListApplicationsRequest{
		CompartmentId: &s.compartmentId,
		Limit:         limit,
		Page:          params.Cursor,
		DisplayName:   params.Name,
	}

	var applicationSummaries []functions.ApplicationSummary

	for {
		res, err := s.ociClient.ListApplications(ctxOrBackground(params.Context), req)
		if err != nil {
			return nil, err
		}

		applicationSummaries = append(applicationSummaries, res.Items...)

		if res.OpcNextPage != nil {
			req.Page = res.OpcNextPage
		} else {
			break
		}
	}

	var items []*modelsv2.App

	// Consumers such as Fn CLI expect to get 'config' and 'syslogUrl' when doing a filter-by-name
	// Given ApplicationSummary doesn't have these fields, we do a follow-up GetApp to get the full Application entity
	// We could possibly optimise Fn CLI usage of this somehow so it's only used where necessary (variable in ctx?)
	if params.Name != nil && len(applicationSummaries) == 1 {
		getAppOK, err := s.GetApp(&apps.GetAppParams{
			AppID:   *applicationSummaries[0].Id,
			Context: ctxOrBackground(params.Context),
		})
		if err != nil {
			return nil, err
		}

		items = append(items, getAppOK.Payload)
	} else {
		for _, a := range applicationSummaries {
			items = append(items, ociAppSummaryToV2(a))
		}
	}

	return &apps.ListAppsOK{
		Payload: &modelsv2.AppList{
			Items: items,
		},
	}, nil
}

func (s *appsShim) UpdateApp(params *apps.UpdateAppParams) (*apps.UpdateAppOK, error) {
	var etag *string

	// We can respect 'omitempty' here - only do get-and-merge on config if present
	if params.Body.Config != nil && len(params.Body.Config) != 0 {
		// Get the current version of the App so that we can merge config
		req := functions.GetApplicationRequest{ApplicationId: &params.AppID}

		res, err := s.ociClient.GetApplication(ctxOrBackground(params.Context), req)
		if err != nil {
			return nil, err
		}

		params.Body.Config = mergeConfig(res.Config, params.Body.Config)

		etag = res.Etag
	}

	details := functions.UpdateApplicationDetails{
		Config:    params.Body.Config,
		SyslogUrl: params.Body.SyslogURL,
	}

	req := functions.UpdateApplicationRequest{
		ApplicationId:            &params.AppID,
		UpdateApplicationDetails: details,
		IfMatch:                  etag,
	}

	res, err := s.ociClient.UpdateApplication(ctxOrBackground(params.Context), req)
	if err != nil {
		return nil, err
	}

	return &apps.UpdateAppOK{
		Payload: ociAppToV2(res.Application),
	}, nil
}

func (*appsShim) SetTransport(runtime.ClientTransport) {}

func parseSubnetIds(annotations map[string]interface{}) ([]string, error) {
	if annotations == nil || len(annotations) == 0 {
		return nil, fmt.Errorf("missing subnets annotation")
	}

	var subnets []string
	subnetsInterface, ok := annotations[annotationSubnet]
	if !ok {
		return nil, fmt.Errorf("missing subnets annotation")
	}

	// Typecast to []interface{}
	subnetsArray, success := subnetsInterface.([]interface{})

	if !success {
		return nil, fmt.Errorf("invalid subnets annotation")
	}

	for _, s := range subnetsArray {
		// Typecast to string
		subnetString, secondSuccess := s.(string)
		if !secondSuccess {
			return nil, fmt.Errorf("invalid subnets annotation")
		}
		subnets = append(subnets, subnetString)
	}

	return subnets, nil
}

func ociAppToV2(ociApp functions.Application) *modelsv2.App {
	annotations := make(map[string]interface{})
	annotations[annotationCompartmentId] = *ociApp.CompartmentId
	annotations[annotationSubnet] = ociSubnetsToAnnotationValue(ociApp.SubnetIds)

	return &modelsv2.App{
		Annotations: annotations,
		Config:      ociApp.Config,
		CreatedAt:   strfmt.DateTime(ociApp.TimeCreated.Time),
		ID:          *ociApp.Id,
		Name:        *ociApp.DisplayName,
		SyslogURL:   ociApp.SyslogUrl,
		UpdatedAt:   strfmt.DateTime(ociApp.TimeUpdated.Time),
	}
}

func ociAppSummaryToV2(ociAppSummary functions.ApplicationSummary) *modelsv2.App {
	annotations := make(map[string]interface{})
	annotations[annotationCompartmentId] = *ociAppSummary.CompartmentId
	annotations[annotationSubnet] = ociSubnetsToAnnotationValue(ociAppSummary.SubnetIds)

	return &modelsv2.App{
		Annotations: annotations,
		CreatedAt:   strfmt.DateTime(ociAppSummary.TimeCreated.Time),
		ID:          *ociAppSummary.Id,
		Name:        *ociAppSummary.DisplayName,
		UpdatedAt:   strfmt.DateTime(ociAppSummary.TimeUpdated.Time),
	}
}

// Behaviour of go-swagger v2 client is to return slice of interfaces
func ociSubnetsToAnnotationValue(subnets []string) []interface{} {
	ifs := make([]interface{}, len(subnets))
	for i, s := range subnets {
		ifs[i] = s
	}
	return ifs
}
