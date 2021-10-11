package client

import (
	"context"
	"fmt"
	"github.com/golang/mock/gomock"
	"github.com/oracle/oci-go-sdk/v48/common"
	"github.com/oracle/oci-go-sdk/v48/functions"
	"time"
)

func NewMockFunctionsManagementClientBasic(ctrl *gomock.Controller) FunctionsManagementClient {
	m := NewMockFunctionsManagementClient(ctrl)

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
				compartment := "GetApplicationCompartment"
				displayName := "GetApplicationDisplayName"
				syslogUrl := "GetApplicationSyslogUrl"
				return functions.GetApplicationResponse{
					Application: functions.Application{
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
					},
				}, nil
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
				id := "UpdateApplicationId"
				compartment := "UpdateApplicationCompartment"
				displayName := "UpdateApplicationDisplayName"
				config := map[string]string{
					"UpdateApplicationKey1": "UpdateApplicationValue1",
					"UpdateApplicationKey2": "UpdateApplicationValue2",
				}
				if request.Config != nil {
					config = request.Config
				}
				syslogUrl := "OriginalApplicationSyslogUrl"
				if request.SyslogUrl != nil {
					syslogUrl = *request.SyslogUrl
				}
				return functions.UpdateApplicationResponse{
					Application: functions.Application{
						Id:             &id,
						CompartmentId:  &compartment,
						DisplayName:    &displayName,
						LifecycleState: functions.ApplicationLifecycleStateActive,
						Config:         config,
						SubnetIds:      []string{"UpdateApplicationSubnet"},
						SyslogUrl:      &syslogUrl,
						FreeformTags:   nil,
						DefinedTags:    nil,
						TimeCreated:    &common.SDKTime{Time: time.Now()},
						TimeUpdated:    &common.SDKTime{Time: time.Now()},
					},
				}, nil
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
				return functions.CreateFunctionResponse{
					Function: functions.Function{
						Id:               &id,
						ApplicationId:    request.ApplicationId,
						CompartmentId:    &compartment,
						DisplayName:      request.DisplayName,
						LifecycleState:   functions.FunctionLifecycleStateActive,
						Image:            request.Image,
						ImageDigest:      request.ImageDigest,
						MemoryInMBs:      request.MemoryInMBs,
						TimeoutInSeconds: request.TimeoutInSeconds,
						InvokeEndpoint:   &invokeEndpoint,
						Config:           request.Config,
						FreeformTags:     request.FreeformTags,
						DefinedTags:      request.DefinedTags,
						TimeCreated:      &common.SDKTime{Time: time.Now()},
						TimeUpdated:      &common.SDKTime{Time: time.Now()},
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
				application := "GetFunctionApplication"
				compartment := "GetFunctionCompartment"
				displayName := "GetFunctionDisplayName"
				image := "GetFunctionImage"
				digest := "GetFunctionDigest"
				memory := int64(128)
				timeout := 30
				invokeEndpoint := "GetFunctionInvokeEndpoint"
				return functions.GetFunctionResponse{
					Function: functions.Function{
						Id:               request.FunctionId,
						ApplicationId:    &application,
						CompartmentId:    &compartment,
						DisplayName:      &displayName,
						LifecycleState:   functions.FunctionLifecycleStateActive,
						Image:            &image,
						ImageDigest:      &digest,
						MemoryInMBs:      &memory,
						TimeoutInSeconds: &timeout,
						InvokeEndpoint:   &invokeEndpoint,
						Config: map[string]string{
							"GetFunctionKey1": "GetFunctionValue1",
							"GetFunctionKey2": "GetFunctionValue2",
						},
						FreeformTags: nil,
						DefinedTags:  nil,
						TimeCreated:  &common.SDKTime{Time: time.Now()},
						TimeUpdated:  &common.SDKTime{Time: time.Now()},
					},
				}, nil
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
				id := "UpdateFunctionId"
				application := "UpdateFunctionApplication"
				compartment := "UpdateFunctionCompartment"
				displayName := "UpdateFunctionDisplayName"
				invokeEndpoint := "UpdateFunctionInvokeEndpoint"
				config := map[string]string{
					"UpdateFunctionKey1": "UpdateFunctionValue1",
					"UpdateFunctionKey2": "UpdateFunctionValue2",
				}
				if request.Config != nil {
					config = request.Config
				}
				image := "OriginalFunctionImage"
				if request.Image != nil {
					image = *request.Image
				}
				digest := "OriginalFunctionDigest"
				if request.ImageDigest != nil {
					digest = *request.ImageDigest
					if digest == "" {
						return functions.UpdateFunctionResponse{}, fmt.Errorf("invalid image digest")
					}
				}
				memory := int64(128)
				if request.MemoryInMBs != nil {
					memory = *request.MemoryInMBs
				}
				timeout := 30
				if request.TimeoutInSeconds != nil {
					timeout = *request.TimeoutInSeconds
				}
				return functions.UpdateFunctionResponse{
					Function: functions.Function{
						Id:               &id,
						ApplicationId:    &application,
						CompartmentId:    &compartment,
						DisplayName:      &displayName,
						LifecycleState:   functions.FunctionLifecycleStateActive,
						Image:            &image,
						ImageDigest:      &digest,
						MemoryInMBs:      &memory,
						TimeoutInSeconds: &timeout,
						InvokeEndpoint:   &invokeEndpoint,
						Config:           config,
						FreeformTags:     nil,
						DefinedTags:      nil,
						TimeCreated:      &common.SDKTime{Time: time.Now()},
						TimeUpdated:      &common.SDKTime{Time: time.Now()},
					},
				}, nil
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
	invokeEndpoint := "FunctionSummaryInvokeEndpoint"
	return functions.FunctionSummary{
		Id:               &id,
		CompartmentId:    &compartment,
		ApplicationId:    application,
		DisplayName:      &displayName,
		LifecycleState:   functions.FunctionLifecycleStateActive,
		Image:            &image,
		ImageDigest:      &digest,
		MemoryInMBs:      &memory,
		TimeoutInSeconds: &timeout,
		InvokeEndpoint:   &invokeEndpoint,
		FreeformTags:     nil,
		DefinedTags:      nil,
		TimeCreated:      &common.SDKTime{Time: time.Now()},
		TimeUpdated:      &common.SDKTime{Time: time.Now()},
	}
}
