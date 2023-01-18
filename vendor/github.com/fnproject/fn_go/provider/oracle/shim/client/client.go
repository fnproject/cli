package client

import (
	"context"
	"github.com/oracle/oci-go-sdk/v65/functions"
)

// Interface extracted from Go SDK FunctionsManagementClient for mockability
type FunctionsManagementClient interface {
	CreateApplication(ctx context.Context, request functions.CreateApplicationRequest) (response functions.CreateApplicationResponse, err error)
	CreateFunction(ctx context.Context, request functions.CreateFunctionRequest) (response functions.CreateFunctionResponse, err error)
	DeleteApplication(ctx context.Context, request functions.DeleteApplicationRequest) (response functions.DeleteApplicationResponse, err error)
	DeleteFunction(ctx context.Context, request functions.DeleteFunctionRequest) (response functions.DeleteFunctionResponse, err error)
	GetApplication(ctx context.Context, request functions.GetApplicationRequest) (response functions.GetApplicationResponse, err error)
	GetFunction(ctx context.Context, request functions.GetFunctionRequest) (response functions.GetFunctionResponse, err error)
	ListApplications(ctx context.Context, request functions.ListApplicationsRequest) (response functions.ListApplicationsResponse, err error)
	ListFunctions(ctx context.Context, request functions.ListFunctionsRequest) (response functions.ListFunctionsResponse, err error)
	UpdateApplication(ctx context.Context, request functions.UpdateApplicationRequest) (response functions.UpdateApplicationResponse, err error)
	UpdateFunction(ctx context.Context, request functions.UpdateFunctionRequest) (response functions.UpdateFunctionResponse, err error)
}
