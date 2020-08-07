package oci

import (
	"context"
	"fmt"
	"github.com/fnproject/cli/adapter"
	"github.com/go-openapi/strfmt"
	"github.com/oracle/oci-go-sdk/functions"
	"os"
)

type FnClient struct {
	client *functions.FunctionsManagementClient
}

func (f FnClient) CreateFn(fn *adapter.Fn) (*adapter.Fn, error) {
	memory := int64(fn.Memory)
	timeout := int(*fn.Timeout)

	body := functions.CreateFunctionDetails{
		ApplicationId: &fn.AppID,
		MemoryInMBs: &memory,
		Image: &fn.Image,
		TimeoutInSeconds: &timeout,
		Config: fn.Config,
		DisplayName: &fn.Name,
	}
	req := functions.CreateFunctionRequest{CreateFunctionDetails: body,}

	res,err := f.client.CreateFunction(context.Background(), req)

	if err != nil {
		return nil, err
	}

	adapterFn := convertOCIFnToAdapterFn(&res.Function)
	return adapterFn, nil
}

func (f FnClient) GetFn(appID string, fnName string) (*adapter.Fn, error) {
	req := functions.ListFunctionsRequest{ApplicationId: &appID, DisplayName: &fnName}
	resp, err := f.client.ListFunctions(context.Background(), req)
	if err != nil {
		return nil, err
	}

	if len(resp.Items) > 0 {
		adapterFn := convertOCIFnSummaryToAdapterFn(&resp.Items[0])
		return adapterFn, nil
	} else {
		return nil, adapter.FunctionNameNotFoundError{Name: fnName}
	}
}

func (f FnClient) GetFnByFnID(fnID string) (*adapter.Fn, error) {
	req := functions.GetFunctionRequest{FunctionId: &fnID}
	resp, err := f.client.GetFunction(context.Background(), req)

	if err != nil {
		return nil, err
	}

	return convertOCIFnToAdapterFn(&resp.Function), nil
}

func (f FnClient) UpdateFn(fn *adapter.Fn) (*adapter.Fn, error) {
	timeout := int(*fn.Timeout)
	memory := int64(fn.Memory)

	body := functions.UpdateFunctionDetails{
		Image: &fn.Image,
		TimeoutInSeconds: &timeout,
		MemoryInMBs: &memory,
		Config: fn.Config,
	}

	req := functions.UpdateFunctionRequest{UpdateFunctionDetails: body,}
	res, err := f.client.UpdateFunction(context.Background(), req)

	if err != nil {
		return nil, err
	}

	adapterFn := convertOCIFnToAdapterFn(&res.Function)
	return adapterFn, nil
}

func (f FnClient) DeleteFn(fnID string) error {
	req := functions.DeleteFunctionRequest{FunctionId: &fnID,}
	_, err := f.client.DeleteFunction(context.Background(), req)
	return err
}

func (f FnClient) ListFn(appID string, limit int64) ([]*adapter.Fn, error) {
	var resFns []*adapter.Fn
	req := functions.ListFunctionsRequest{ApplicationId: &appID}

	for {
		resp, err := f.client.ListFunctions(context.Background(), req)
		if err != nil {
			return nil, err
		}

		adapterFns := convertOCIFnsToAdapterFns(&resp.Items)
		resFns = append(resFns, adapterFns...)
		howManyMore := limit - int64(len(resFns)+len(resp.Items))

		if howManyMore <= 0 || resp.OpcNextPage == nil {
			break
		}

		req.Page = resp.OpcNextPage
	}

	if len(resFns) == 0 {
		fmt.Fprint(os.Stderr, "No apps found\n")
		return nil, nil
	}

	return resFns, nil
}

func convertOCIFnsToAdapterFns(ociFns *[]functions.FunctionSummary) []*adapter.Fn {
	var resFns []*adapter.Fn
	for _, ociFn := range *ociFns {
		fn := convertOCIFnSummaryToAdapterFn(&ociFn)
		resFns = append(resFns, fn)
	}
	return resFns
}

func convertOCIFnSummaryToAdapterFn(ociFn *functions.FunctionSummary) *adapter.Fn {
	createdAt,_ := strfmt.ParseDateTime(ociFn.TimeCreated.String())
	updatedAt,_ := strfmt.ParseDateTime(ociFn.TimeUpdated.String())
	timeout := int32(*ociFn.TimeoutInSeconds)
	return &adapter.Fn{
		Name: *ociFn.DisplayName,
		ID: *ociFn.Id,
		AppID: *ociFn.ApplicationId,
		Timeout: &timeout,
		Image: *ociFn.Image,
		Memory: uint64(*ociFn.MemoryInMBs),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}
}

func convertOCIFnToAdapterFn(ociFn *functions.Function) *adapter.Fn {
	createdAt,_ := strfmt.ParseDateTime(ociFn.TimeCreated.String())
	updatedAt,_ := strfmt.ParseDateTime(ociFn.TimeUpdated.String())
	timeout := int32(*ociFn.TimeoutInSeconds)
	return &adapter.Fn{
		Name: *ociFn.DisplayName,
		ID: *ociFn.Id,
		AppID: *ociFn.ApplicationId,
		Timeout: &timeout,
		Image: *ociFn.Image,
		Memory: uint64(*ociFn.MemoryInMBs),
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
		Config: ociFn.Config,
	}
}