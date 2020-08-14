package oci

import (
	"bytes"
	"context"
	"encoding/json"
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
	memory := int64(fn.Memory)
	var memoryPtr *int64
	if memory != 0 {
		memoryPtr = &memory
	}

	digest, err := parseDigestAnnotation(fn)
	if err != nil {
		return nil, err
	}

	body := functions.UpdateFunctionDetails{
		Image:            &fn.Image,
		TimeoutInSeconds: parseTimeout(fn.Timeout),
		MemoryInMBs:      memoryPtr,
		Config:           fn.Config,
		ImageDigest:      digest,
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
	createdAt, _ := strfmt.ParseDateTime(ociFn.TimeCreated.Format(time.RFC3339Nano))
	updatedAt, _ := strfmt.ParseDateTime(ociFn.TimeUpdated.Format(time.RFC3339Nano))

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
	}
}

func convertOCIFnToAdapterFn(ociFn *functions.Function) *adapter.Fn {
	createdAt, _ := strfmt.ParseDateTime(ociFn.TimeCreated.Format(time.RFC3339Nano))
	updatedAt, _ := strfmt.ParseDateTime(ociFn.TimeUpdated.Format(time.RFC3339Nano))

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
	}
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

	var digest string
	if digestInterface, found := fn.Annotations[AnnotationImageDigest]; found {
		digestBytes := digestInterface.([]byte)
		if err := json.NewDecoder(bytes.NewReader(digestBytes)).Decode(&digest); err != nil {
			return nil, err
		}
	}
	return &digest, nil
}
