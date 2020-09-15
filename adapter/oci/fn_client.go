package oci

import (
	"context"
	"errors"
	"fmt"
	"github.com/fnproject/cli/adapter"
	"github.com/go-openapi/strfmt"
	"github.com/oracle/oci-go-sdk/functions"
	"os"
	"time"
)

const (
	DefaultMemory uint64 = 128 // MB

	// AnnotationImageDigest contains the resolved docker image digest associated with a function
	AnnotationImageDigest = "oracle.com/oci/imageDigest"

	// FnInvokeEndpointAnnotation is the annotation that exposes the fn invoke endpoint For want of a better place to put this it's here
	FnInvokeEndpointAnnotation = "fnproject.io/fn/invokeEndpoint"

	// AnnotationCompartmentID Represents OCI Providers compartment
	AnnotationCompartmentID = "oracle.com/oci/compartmentId"

	DefaultIdleTimeout int32 = 30 // seconds

	InvokeEndpointFmtString string = "%s/20181201/functions/%s/actions/invoke"
)

type FnClient struct {
	client *functions.FunctionsManagementClient
}

func (f FnClient) CreateFn(fn *adapter.Fn) (*adapter.Fn, error) {
	memory := int64(fn.Memory)
	if memory == 0 {
		memory = int64(DefaultMemory)
	}

	digest, err := parseDigestAnnotation(fn)
	if err != nil {
		return nil, err
	}

	body := functions.CreateFunctionDetails{
		ApplicationId:    &fn.AppID,
		MemoryInMBs:      &memory,
		Image:            &fn.Image,
		TimeoutInSeconds: parseTimeout(fn.Timeout),
		Config:           fn.Config,
		DisplayName:      &fn.Name,
		ImageDigest:      digest,
	}
	req := functions.CreateFunctionRequest{CreateFunctionDetails: body,}

	res, err := f.client.CreateFunction(context.Background(), req)

	if err != nil {
		return nil, err
	}

	return convertOCIFnToAdapterFn(&res.Function)
}

func (f FnClient) GetFn(appID string, fnName string) (*adapter.Fn, error) {
	req := functions.ListFunctionsRequest{ApplicationId: &appID, DisplayName: &fnName}
	resp, err := f.client.ListFunctions(context.Background(), req)
	if err != nil {
		return nil, err
	}

	if len(resp.Items) > 0 {
		return f.GetFnByFnID(*resp.Items[0].Id)
	} else {
		return nil, adapter.FunctionNameNotFoundError{Name: fnName}
	}
}

func (f FnClient) GetFnByFnID(fnID string) (*adapter.Fn, error) {
	resp, err := f.getFnByFnIDRaw(fnID)

	if err != nil {
		return nil, err
	}

	return convertOCIFnToAdapterFn(&resp.Function)
}

func (f FnClient) getFnByFnIDRaw(fnID string) (functions.GetFunctionResponse, error) {
	req := functions.GetFunctionRequest{FunctionId: &fnID}
	return f.client.GetFunction(context.Background(), req)
}

func (f FnClient) UpdateFn(fn *adapter.Fn) (*adapter.Fn, error) {
	memory := int64(fn.Memory)
	var memoryPtr *int64
	if memory != 0 {
		memoryPtr = &memory
	}

	digest, err := parseDigestAnnotation(fn)
	if err != nil {
		return nil, err
	}

	var updateRes functions.UpdateFunctionResponse
	var updateErr error

	for i := 0; i < NoEtagMatchRetryCount; i++ {
		getRes, getErr := f.getFnByFnIDRaw(fn.ID)

		if getErr != nil {
			return nil, getErr
		}

		body := functions.UpdateFunctionDetails{
			Image:            &fn.Image,
			TimeoutInSeconds: parseTimeout(fn.Timeout),
			MemoryInMBs:      memoryPtr,
			Config:           mergeConfig(getRes.Config, fn.Config),
			ImageDigest:      digest,
		}

		req := functions.UpdateFunctionRequest{UpdateFunctionDetails: body, FunctionId: &fn.ID, IfMatch: getRes.Etag}
		updateRes, updateErr = f.client.UpdateFunction(context.Background(), req)

		if updateErr == nil || updateRes.HTTPResponse().StatusCode != NoEtagMatchStatusCode {
			// Break here and do not retry if there is no error or if error is not `NoEtagMatch`
			// See: https://docs.cloud.oracle.com/en-us/iaas/Content/API/References/apierrors.htm
			break
		}
	}

	if updateErr != nil {
		return nil, err
	}

	return convertOCIFnToAdapterFn(&updateRes.Function)
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

		adapterFns, err := convertOCIFnsToAdapterFns(&resp.Items)
		if err != nil {
			return nil, err
		}

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

func convertOCIFnsToAdapterFns(ociFns *[]functions.FunctionSummary) ([]*adapter.Fn, error) {
	var resFns []*adapter.Fn
	for _, ociFn := range *ociFns {
		fn, err := convertOCIFnSummaryToAdapterFn(&ociFn)
		if err != nil {
			return nil, err
		}
		resFns = append(resFns, fn)
	}
	return resFns, nil
}

func convertOCIFnSummaryToAdapterFn(ociFn *functions.FunctionSummary) (*adapter.Fn, error) {
	createdAt, err := strfmt.ParseDateTime(ociFn.TimeCreated.Format(time.RFC3339Nano))
	if err != nil {
		return nil, errors.New("missing or invalid TimeCreated in function")
	}

	updatedAt, err := strfmt.ParseDateTime(ociFn.TimeUpdated.Format(time.RFC3339Nano))
	if err != nil {
		return nil, errors.New("missing or invalid TimeUpdated in function")
	}

	var timeoutPtr *int32
	timeoutPtr = nil
	if ociFn.TimeoutInSeconds != nil {
		timeout := int32(*ociFn.TimeoutInSeconds)
		timeoutPtr = &timeout
	}

	annotationMap := make(map[string]interface{})
	invokeEndpoint := fmt.Sprintf(InvokeEndpointFmtString, *ociFn.InvokeEndpoint, *ociFn.Id)
	annotationMap[FnInvokeEndpointAnnotation] = invokeEndpoint
	annotationMap[AnnotationCompartmentID] = *ociFn.CompartmentId
	annotationMap[AnnotationImageDigest] = *ociFn.ImageDigest

	defaultIdleTimeout := DefaultIdleTimeout
	return &adapter.Fn{
		Name:        *ociFn.DisplayName,
		ID:          *ociFn.Id,
		AppID:       *ociFn.ApplicationId,
		Timeout:     timeoutPtr,
		Image:       *ociFn.Image,
		Memory:      uint64(*ociFn.MemoryInMBs),
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Annotations: annotationMap,
		IDLETimeout: &defaultIdleTimeout,
	}, nil
}

func convertOCIFnToAdapterFn(ociFn *functions.Function) (*adapter.Fn, error) {
	createdAt, err := strfmt.ParseDateTime(ociFn.TimeCreated.Format(time.RFC3339Nano))
	if err != nil {
		return nil, errors.New("missing or invalid TimeCreated in function")
	}

	updatedAt, err := strfmt.ParseDateTime(ociFn.TimeUpdated.Format(time.RFC3339Nano))
	if err != nil {
		return nil, errors.New("missing or invalid TimeUpdated in function")
	}


	var timeoutPtr *int32
	timeoutPtr = nil
	if ociFn.TimeoutInSeconds != nil {
		timeout := int32(*ociFn.TimeoutInSeconds)
		timeoutPtr = &timeout
	}

	annotationMap := make(map[string]interface{})
	invokeEndpoint := fmt.Sprintf(InvokeEndpointFmtString, *ociFn.InvokeEndpoint, *ociFn.Id)
	annotationMap[FnInvokeEndpointAnnotation] = invokeEndpoint
	annotationMap[AnnotationCompartmentID] = *ociFn.CompartmentId
	annotationMap[AnnotationImageDigest] = *ociFn.ImageDigest

	defaultIdleTimeout := DefaultIdleTimeout
	return &adapter.Fn{
		Name:        *ociFn.DisplayName,
		ID:          *ociFn.Id,
		AppID:       *ociFn.ApplicationId,
		Timeout:     timeoutPtr,
		Image:       *ociFn.Image,
		Memory:      uint64(*ociFn.MemoryInMBs),
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		Config:      ociFn.Config,
		Annotations: annotationMap,
		IDLETimeout: &defaultIdleTimeout,
	}, nil
}

func parseTimeout(timeout *int32) *int {
	if timeout == nil {
		return nil
	}
	result := int(*timeout)
	return &result
}

func parseDigestAnnotation(fn *adapter.Fn) (*string, error) {
	if fn.Annotations == nil {
		return nil, nil
	}

	digestInterface, ok := fn.Annotations[AnnotationImageDigest]
	if !ok {
		// Missing ImageDigest
		return nil, nil
	}

	// Typecast to string
	digest, success := digestInterface.(string)

	if !success {
		return nil, errors.New("Invalid image digest")
	}

	return &digest, nil
}
