package client

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/oracle/oci-go-sdk/v65/common"
	"github.com/oracle/oci-go-sdk/v65/functions"
	"time"
)

func NewMockFunctionsManagementClientBasic(ctrl *gomock.Controller) FunctionsManagementClient {
	m := NewMockFunctionsManagementClient(ctrl)
	var storedApplication *functions.Application
	var storedFunction *functions.Function

	// CreateApplication
	m.EXPECT().
		CreateApplication(
			gomock.Any(),
			gomock.AssignableToTypeOf(functions.CreateApplicationRequest{}),
		).
		DoAndReturn(
			func(ctx context.Context, request functions.CreateApplicationRequest) (functions.CreateApplicationResponse, error) {
				id := "CreateApplicationId"
				return functions.CreateApplicationResponse{
					Application: functions.Application{
						Id:             &id,
						CompartmentId:  request.CompartmentId,
						DisplayName:    request.DisplayName,
						LifecycleState: functions.ApplicationLifecycleStateActive,
						Config:         request.Config,
						SubnetIds:      request.SubnetIds,
						SyslogUrl:      request.SyslogUrl,
						FreeformTags:   request.FreeformTags,
						DefinedTags:    request.DefinedTags,
						TimeCreated:    &common.SDKTime{Time: time.Now()},
						TimeUpdated:    &common.SDKTime{Time: time.Now()},
					},
				}, nil
			},
		).
		AnyTimes()

	// DeleteApplication
	m.EXPECT().
		DeleteApplication(
			gomock.Any(),
			gomock.AssignableToTypeOf(functions.DeleteApplicationRequest{}),
		).
		Return(functions.DeleteApplicationResponse{}, nil).
		AnyTimes()

	// GetApplication
	m.EXPECT().
		GetApplication(
			gomock.Any(),
			gomock.AssignableToTypeOf(functions.GetApplicationRequest{}),
		).
		DoAndReturn(
			func(ctx context.Context, request functions.GetApplicationRequest) (functions.GetApplicationResponse, error) {
				if storedApplication != nil {
					return functions.GetApplicationResponse{Application: *storedApplication}, nil
				}
				compartment := "GetApplicationCompartment"
				displayName := "GetApplicationDisplayName"
				syslogUrl := "GetApplicationSyslogUrl"
				storedApplication = &functions.Application{
					Id:             request.ApplicationId,
					CompartmentId:  &compartment,
					DisplayName:    &displayName,
					LifecycleState: functions.ApplicationLifecycleStateActive,
					Config: map[string]string{
						"GetApplicationKey1": "GetApplicationValue1",
						"GetApplicationKey2": "GetApplicationValue2",
					},
					SubnetIds:    []string{"GetApplicationSubnet"},
					SyslogUrl:    &syslogUrl,
					FreeformTags: nil,
					DefinedTags:  nil,
					TimeCreated:  &common.SDKTime{Time: time.Now()},
					TimeUpdated:  &common.SDKTime{Time: time.Now()},
				}
				return functions.GetApplicationResponse{Application: *storedApplication}, nil
			},
		).
		AnyTimes()

	// ListApplications
	m.EXPECT().
		ListApplications(
			gomock.Any(),
			gomock.AssignableToTypeOf(functions.ListApplicationsRequest{}),
		).
		DoAndReturn(
			func(ctx context.Context, request functions.ListApplicationsRequest) (functions.ListApplicationsResponse, error) {
				if request.DisplayName != nil && *request.DisplayName != "" {
					return functions.ListApplicationsResponse{
						Items: []functions.ApplicationSummary{
							newBasicApplicationSummary(0, request.CompartmentId),
						},
					}, nil
				}

				page0 := []functions.ApplicationSummary{
					newBasicApplicationSummary(0, request.CompartmentId),
					newBasicApplicationSummary(1, request.CompartmentId),
					newBasicApplicationSummary(2, request.CompartmentId),
				}
				page1 := []functions.ApplicationSummary{
					newBasicApplicationSummary(3, request.CompartmentId),
					newBasicApplicationSummary(4, request.CompartmentId),
					newBasicApplicationSummary(5, request.CompartmentId),
				}
				page2 := []functions.ApplicationSummary{
					newBasicApplicationSummary(6, request.CompartmentId),
					newBasicApplicationSummary(7, request.CompartmentId),
					newBasicApplicationSummary(8, request.CompartmentId),
				}

				var response functions.ListApplicationsResponse
				if request.Page == nil {
					opcNextPage := "1"
					response = functions.ListApplicationsResponse{Items: page0, OpcNextPage: &opcNextPage}
				} else if *request.Page == "1" {
					opcNextPage := "2"
					response = functions.ListApplicationsResponse{Items: page1, OpcNextPage: &opcNextPage}
				} else if *request.Page == "2" {
					response = functions.ListApplicationsResponse{Items: page2}
				}
				return response, nil
			},
		).
		AnyTimes()

	// UpdateApplication
	m.EXPECT().
		UpdateApplication(
			gomock.Any(),
			gomock.AssignableToTypeOf(functions.UpdateApplicationRequest{}),
		).
		DoAndReturn(
			func(ctx context.Context, request functions.UpdateApplicationRequest) (functions.UpdateApplicationResponse, error) {
				if storedApplication == nil {
					compartment := "UpdateApplicationCompartment"
					displayName := "UpdateApplicationDisplayName"
					syslogURL := "OriginalApplicationSyslogUrl"
					storedApplication = &functions.Application{
						Id: request.ApplicationId, CompartmentId: &compartment, DisplayName: &displayName,
						LifecycleState: functions.ApplicationLifecycleStateActive,
						Config:         map[string]string{"UpdateApplicationKey1": "UpdateApplicationValue1", "UpdateApplicationKey2": "UpdateApplicationValue2"},
						SubnetIds:      []string{"UpdateApplicationSubnet"}, SyslogUrl: &syslogURL,
						TimeCreated: &common.SDKTime{Time: time.Now()}, TimeUpdated: &common.SDKTime{Time: time.Now()},
					}
				}
				if request.Config != nil {
					storedApplication.Config = request.Config
				}
				if request.SyslogUrl != nil {
					storedApplication.SyslogUrl = request.SyslogUrl
				}
				if request.FreeformTags != nil {
					storedApplication.FreeformTags = request.FreeformTags
				}
				if request.DefinedTags != nil {
					storedApplication.DefinedTags = request.DefinedTags
				}
				return functions.UpdateApplicationResponse{}, nil
			},
		).
		AnyTimes()

	// CreateFunction
	m.EXPECT().
		CreateFunction(
			gomock.Any(),
			gomock.AssignableToTypeOf(functions.CreateFunctionRequest{}),
		).
		DoAndReturn(
			func(ctx context.Context, request functions.CreateFunctionRequest) (functions.CreateFunctionResponse, error) {
				id := "CreateFunctionId"
				compartment := "CreateFunctionCompartment"
				invokeEndpoint := "CreateFunctionInvokeEndpoint"
				var image *string
				var digest *string
				if src, ok := request.SourceDetails.(functions.CreateContainerImageFunctionSourceDetails); ok {
					image = src.Image
					digest = src.ImageDigest
				}
				return functions.CreateFunctionResponse{
					Function: functions.Function{
						Id:                           &id,
						ApplicationId:                request.ApplicationId,
						CompartmentId:                &compartment,
						DisplayName:                  request.DisplayName,
						LifecycleState:               functions.FunctionLifecycleStateActive,
						SourceDetails:                functions.ContainerImageFunctionSourceDetails{Image: image, ImageDigest: digest},
						MemoryInMBs:                  request.MemoryInMBs,
						TimeoutInSeconds:             request.TimeoutInSeconds,
						InvokeEndpoint:               &invokeEndpoint,
						Config:                       request.Config,
						ProvisionedConcurrencyConfig: functions.NoneProvisionedConcurrencyConfig{},
						FreeformTags:                 request.FreeformTags,
						DefinedTags:                  request.DefinedTags,
						TimeCreated:                  &common.SDKTime{Time: time.Now()},
						TimeUpdated:                  &common.SDKTime{Time: time.Now()},
					},
				}, nil
			},
		).
		AnyTimes()

	// DeleteFunction
	m.EXPECT().
		DeleteFunction(
			gomock.Any(),
			gomock.AssignableToTypeOf(functions.DeleteFunctionRequest{}),
		).
		Return(functions.DeleteFunctionResponse{}, nil).
		AnyTimes()

	// GetFunction
	m.EXPECT().
		GetFunction(
			gomock.Any(),
			gomock.AssignableToTypeOf(functions.GetFunctionRequest{}),
		).
		DoAndReturn(
			func(ctx context.Context, request functions.GetFunctionRequest) (functions.GetFunctionResponse, error) {
				if storedFunction != nil {
					return functions.GetFunctionResponse{Function: *storedFunction}, nil
				}
				application := "GetFunctionApplication"
				compartment := "GetFunctionCompartment"
				displayName := "GetFunctionDisplayName"
				image := "GetFunctionImage"
				digest := "GetFunctionDigest"
				memory := int64(128)
				timeout := 30
				pcCount := 5
				invokeEndpoint := "GetFunctionInvokeEndpoint"
				storedFunction = &functions.Function{
					Id:                           request.FunctionId,
					ApplicationId:                &application,
					CompartmentId:                &compartment,
					DisplayName:                  &displayName,
					LifecycleState:               functions.FunctionLifecycleStateActive,
					SourceDetails:                functions.ContainerImageFunctionSourceDetails{Image: &image, ImageDigest: &digest},
					MemoryInMBs:                  &memory,
					TimeoutInSeconds:             &timeout,
					ProvisionedConcurrencyConfig: functions.ConstantProvisionedConcurrencyConfig{Count: &pcCount},
					InvokeEndpoint:               &invokeEndpoint,
					Config: map[string]string{
						"GetFunctionKey1": "GetFunctionValue1",
						"GetFunctionKey2": "GetFunctionValue2",
					},
					FreeformTags: nil,
					DefinedTags:  nil,
					TimeCreated:  &common.SDKTime{Time: time.Now()},
					TimeUpdated:  &common.SDKTime{Time: time.Now()},
				}
				return functions.GetFunctionResponse{Function: *storedFunction}, nil
			},
		).
		AnyTimes()

	// ListFunctions
	m.EXPECT().
		ListFunctions(
			gomock.Any(),
			gomock.AssignableToTypeOf(functions.ListFunctionsRequest{}),
		).
		DoAndReturn(
			func(ctx context.Context, request functions.ListFunctionsRequest) (functions.ListFunctionsResponse, error) {
				if request.DisplayName != nil && *request.DisplayName != "" {
					return functions.ListFunctionsResponse{
						Items: []functions.FunctionSummary{
							newBasicFunctionSummary(0, request.ApplicationId),
						},
					}, nil
				}

				page0 := []functions.FunctionSummary{
					newBasicFunctionSummary(0, request.ApplicationId),
					newBasicFunctionSummary(1, request.ApplicationId),
					newBasicFunctionSummary(2, request.ApplicationId),
				}
				page1 := []functions.FunctionSummary{
					newBasicFunctionSummary(3, request.ApplicationId),
					newBasicFunctionSummary(4, request.ApplicationId),
					newBasicFunctionSummary(5, request.ApplicationId),
				}
				page2 := []functions.FunctionSummary{
					newBasicFunctionSummary(6, request.ApplicationId),
					newBasicFunctionSummary(7, request.ApplicationId),
					newBasicFunctionSummary(8, request.ApplicationId),
				}

				var response functions.ListFunctionsResponse
				if request.Page == nil {
					opcNextPage := "1"
					response = functions.ListFunctionsResponse{Items: page0, OpcNextPage: &opcNextPage}
				} else if *request.Page == "1" {
					opcNextPage := "2"
					response = functions.ListFunctionsResponse{Items: page1, OpcNextPage: &opcNextPage}
				} else if *request.Page == "2" {
					response = functions.ListFunctionsResponse{Items: page2}
				}
				return response, nil
			},
		).
		AnyTimes()

	// UpdateFunction
	m.EXPECT().
		UpdateFunction(
			gomock.Any(),
			gomock.AssignableToTypeOf(functions.UpdateFunctionRequest{}),
		).
		DoAndReturn(
			func(ctx context.Context, request functions.UpdateFunctionRequest) (functions.UpdateFunctionResponse, error) {
				if storedFunction == nil {
					application, compartment, displayName := "UpdateFunctionApplication", "UpdateFunctionCompartment", "UpdateFunctionDisplayName"
					image, digest, invokeEndpoint := "OriginalFunctionImage", "OriginalFunctionDigest", "UpdateFunctionInvokeEndpoint"
					memory, timeout := int64(128), 30
					storedFunction = &functions.Function{
						Id: request.FunctionId, ApplicationId: &application, CompartmentId: &compartment, DisplayName: &displayName,
						LifecycleState: functions.FunctionLifecycleStateActive,
						SourceDetails:  functions.ContainerImageFunctionSourceDetails{Image: &image, ImageDigest: &digest},
						MemoryInMBs:    &memory, TimeoutInSeconds: &timeout, InvokeEndpoint: &invokeEndpoint,
						Config:      map[string]string{"UpdateFunctionKey1": "UpdateFunctionValue1", "UpdateFunctionKey2": "UpdateFunctionValue2"},
						TimeCreated: &common.SDKTime{Time: time.Now()}, TimeUpdated: &common.SDKTime{Time: time.Now()},
					}
				}
				if src, ok := request.SourceDetails.(functions.UpdateContainerImageFunctionSourceDetails); ok && src.ImageDigest != nil && *src.ImageDigest == "" {
					return functions.UpdateFunctionResponse{}, fmt.Errorf("invalid image digest")
				}
				if request.Config != nil {
					storedFunction.Config = request.Config
				}
				if request.MemoryInMBs != nil {
					storedFunction.MemoryInMBs = request.MemoryInMBs
				}
				if request.TimeoutInSeconds != nil {
					storedFunction.TimeoutInSeconds = request.TimeoutInSeconds
				}
				if src, ok := request.SourceDetails.(functions.UpdateContainerImageFunctionSourceDetails); ok {
					current, _ := storedFunction.SourceDetails.(functions.ContainerImageFunctionSourceDetails)
					if src.Image != nil {
						current.Image = src.Image
					}
					if src.ImageDigest != nil {
						current.ImageDigest = src.ImageDigest
					}
					storedFunction.SourceDetails = current
				}
				if request.FreeformTags != nil {
					storedFunction.FreeformTags = request.FreeformTags
				}
				if request.DefinedTags != nil {
					storedFunction.DefinedTags = request.DefinedTags
				}
				return functions.UpdateFunctionResponse{}, nil
			},
		).
		AnyTimes()

	return m
}

func newBasicApplicationSummary(n int, compartment *string) functions.ApplicationSummary {
	id := fmt.Sprintf("ApplicationSummaryId%d", n)
	displayName := fmt.Sprintf("ApplicationSummaryDisplayName%d", n)
	return functions.ApplicationSummary{
		Id:             &id,
		CompartmentId:  compartment,
		DisplayName:    &displayName,
		LifecycleState: functions.ApplicationLifecycleStateActive,
		SubnetIds:      []string{"ApplicationSummarySubnet"},
		FreeformTags:   nil,
		DefinedTags:    nil,
		TimeCreated:    &common.SDKTime{Time: time.Now()},
		TimeUpdated:    &common.SDKTime{Time: time.Now()},
	}
}

func newBasicFunctionSummary(n int, application *string) functions.FunctionSummary {
	id := fmt.Sprintf("FunctionSummaryId%d", n)
	compartment := "FunctionSummaryCompartment"
	displayName := fmt.Sprintf("FunctionSummaryDisplayName%d", n)
	image := "FunctionSummaryImage"
	digest := "FunctionSummaryDigest"
	memory := int64(128)
	timeout := 30
	pcCount := n + 1
	invokeEndpoint := "FunctionSummaryInvokeEndpoint"
	return functions.FunctionSummary{
		Id:                           &id,
		CompartmentId:                &compartment,
		ApplicationId:                application,
		DisplayName:                  &displayName,
		LifecycleState:               functions.FunctionLifecycleStateActive,
		SourceDetails:                functions.ContainerImageFunctionSourceDetails{Image: &image, ImageDigest: &digest},
		MemoryInMBs:                  &memory,
		TimeoutInSeconds:             &timeout,
		ProvisionedConcurrencyConfig: functions.ConstantProvisionedConcurrencyConfig{Count: &pcCount},
		InvokeEndpoint:               &invokeEndpoint,
		FreeformTags:                 nil,
		DefinedTags:                  nil,
		TimeCreated:                  &common.SDKTime{Time: time.Now()},
		TimeUpdated:                  &common.SDKTime{Time: time.Now()},
	}
}
