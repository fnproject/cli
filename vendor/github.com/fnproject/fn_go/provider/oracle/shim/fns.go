package shim

import (
	"fmt"
	"github.com/fnproject/fn_go/clientv2/fns"
	"github.com/fnproject/fn_go/modelsv2"
	"github.com/fnproject/fn_go/provider/oracle/shim/client"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/oracle/oci-go-sdk/v48/functions"
)

const (
	defaultMemory int64 = 128 // MB

	annotationImageDigest    = "oracle.com/oci/imageDigest"
	annotationInvokeEndpoint = "fnproject.io/fn/invokeEndpoint"

	invokeEndpointFmtString = "%s/20181201/functions/%s/actions/invoke"
)

type fnsShim struct {
	ociClient client.FunctionsManagementClient
}

var _ fns.ClientService = &fnsShim{}

func NewFnsShim(ociClient client.FunctionsManagementClient) fns.ClientService {
	return &fnsShim{ociClient: ociClient}
}

func (s *fnsShim) CreateFn(params *fns.CreateFnParams) (*fns.CreateFnOK, error) {
	memory := int64(params.Body.Memory)
	if memory == 0 {
		memory = defaultMemory
	}

	digest, err := parseDigestAnnotation(params.Body.Annotations)
	if err != nil {
		return nil, err
	}

	details := functions.CreateFunctionDetails{
		DisplayName:      &params.Body.Name,
		ApplicationId:    &params.Body.AppID,
		Image:            &params.Body.Image,
		MemoryInMBs:      &memory,
		ImageDigest:      digest,
		Config:           params.Body.Config,
		TimeoutInSeconds: parseTimeout(params.Body.Timeout),
	}

	req := functions.CreateFunctionRequest{CreateFunctionDetails: details}

	res, err := s.ociClient.CreateFunction(ctxOrBackground(params.Context), req)
	if err != nil {
		return nil, err
	}

	return &fns.CreateFnOK{
		Payload: ociFnToV2(res.Function),
	}, nil
}

func (s *fnsShim) DeleteFn(params *fns.DeleteFnParams) (*fns.DeleteFnNoContent, error) {
	req := functions.DeleteFunctionRequest{FunctionId: &params.FnID}

	_, err := s.ociClient.DeleteFunction(ctxOrBackground(params.Context), req)
	if err != nil {
		return nil, err
	}

	return &fns.DeleteFnNoContent{}, nil
}

func (s *fnsShim) GetFn(params *fns.GetFnParams) (*fns.GetFnOK, error) {
	req := functions.GetFunctionRequest{FunctionId: &params.FnID}

	res, err := s.ociClient.GetFunction(ctxOrBackground(params.Context), req)
	if err != nil {
		return nil, err
	}

	return &fns.GetFnOK{
		Payload: ociFnToV2(res.Function),
	}, nil
}

func (s *fnsShim) ListFns(params *fns.ListFnsParams) (*fns.ListFnsOK, error) {
	var limit *int
	if params.PerPage != nil {
		ppInt := int(*params.PerPage)
		limit = &ppInt
	}

	req := functions.ListFunctionsRequest{
		ApplicationId: params.AppID,
		Limit:         limit,
		Page:          params.Cursor,
		DisplayName:   params.Name,
	}

	var functionSummaries []functions.FunctionSummary

	for {
		res, err := s.ociClient.ListFunctions(ctxOrBackground(params.Context), req)
		if err != nil {
			return nil, err
		}

		functionSummaries = append(functionSummaries, res.Items...)

		if res.OpcNextPage != nil {
			req.Page = res.OpcNextPage
		} else {
			break
		}
	}

	var items []*modelsv2.Fn

	// Consumers such as Fn CLI expect to get 'config' when doing a filter-by-name
	// Given FunctionSummary doesn't have these fields, we do a follow-up GetFn to get the full Function entity
	// We could possibly optimise Fn CLI usage of this somehow so it's only used where necessary (variable in ctx?)
	if params.Name != nil && len(functionSummaries) == 1 {
		getFnOK, err := s.GetFn(&fns.GetFnParams{
			FnID:    *functionSummaries[0].Id,
			Context: ctxOrBackground(params.Context),
		})
		if err != nil {
			return nil, err
		}

		items = append(items, getFnOK.Payload)
	} else {
		for _, f := range functionSummaries {
			items = append(items, ociFnSummaryToV2(f))
		}
	}

	return &fns.ListFnsOK{
		Payload: &modelsv2.FnList{
			Items: items,
		},
	}, nil
}

func (s *fnsShim) UpdateFn(params *fns.UpdateFnParams) (*fns.UpdateFnOK, error) {
	var etag *string

	// We can respect 'omitempty' here - only do get-and-merge on config if present
	if params.Body.Config != nil && len(params.Body.Config) != 0 {
		// Get the current version of the Fn so that we can merge config
		req := functions.GetFunctionRequest{FunctionId: &params.FnID}

		res, err := s.ociClient.GetFunction(ctxOrBackground(params.Context), req)
		if err != nil {
			return nil, err
		}

		params.Body.Config = mergeConfig(res.Config, params.Body.Config)

		etag = res.Etag
	}

	memory := int64(params.Body.Memory)
	var memoryPtr *int64
	if memory != 0 {
		memoryPtr = &memory
	}

	var imagePtr *string
	if params.Body.Image != "" {
		imagePtr = &params.Body.Image
	}

	digest, err := parseDigestAnnotation(params.Body.Annotations)
	if err != nil {
		return nil, err
	}

	details := functions.UpdateFunctionDetails{
		Image:            imagePtr,
		ImageDigest:      digest,
		MemoryInMBs:      memoryPtr,
		Config:           params.Body.Config,
		TimeoutInSeconds: parseTimeout(params.Body.Timeout),
	}

	req := functions.UpdateFunctionRequest{
		FunctionId:            &params.FnID,
		UpdateFunctionDetails: details,
		IfMatch:               etag,
	}

	res, err := s.ociClient.UpdateFunction(ctxOrBackground(params.Context), req)
	if err != nil {
		return nil, err
	}

	return &fns.UpdateFnOK{
		Payload: ociFnToV2(res.Function),
	}, nil
}

func (*fnsShim) SetTransport(runtime.ClientTransport) {}

func parseTimeout(timeout *int32) *int {
	if timeout == nil {
		return nil
	}
	result := int(*timeout)
	return &result
}

func parseDigestAnnotation(annotations map[string]interface{}) (*string, error) {
	if annotations == nil || len(annotations) == 0 {
		return nil, nil
	}

	digestInterface, ok := annotations[annotationImageDigest]
	if !ok {
		// Missing ImageDigest
		return nil, nil
	}

	// Typecast to string
	digest, success := digestInterface.(string)
	if !success {
		return nil, fmt.Errorf("invalid image digest")
	}

	if digest == "" {
		return nil, nil
	}

	return &digest, nil
}

func ociFnToV2(ociFn functions.Function) *modelsv2.Fn {
	annotations := make(map[string]interface{})
	invokeEndpoint := fmt.Sprintf(invokeEndpointFmtString, *ociFn.InvokeEndpoint, *ociFn.Id)
	annotations[annotationCompartmentId] = *ociFn.CompartmentId
	annotations[annotationImageDigest] = *ociFn.ImageDigest
	annotations[annotationInvokeEndpoint] = invokeEndpoint

	var timeoutPtr *int32
	if ociFn.TimeoutInSeconds != nil {
		timeout := int32(*ociFn.TimeoutInSeconds)
		timeoutPtr = &timeout
	}

	return &modelsv2.Fn{
		Annotations: annotations,
		AppID:       *ociFn.ApplicationId,
		Config:      ociFn.Config,
		CreatedAt:   strfmt.DateTime(ociFn.TimeCreated.Time),
		ID:          *ociFn.Id,
		Image:       *ociFn.Image,
		Memory:      uint64(*ociFn.MemoryInMBs),
		Name:        *ociFn.DisplayName,
		Timeout:     timeoutPtr,
		UpdatedAt:   strfmt.DateTime(ociFn.TimeUpdated.Time),
	}
}

func ociFnSummaryToV2(ociFnSummary functions.FunctionSummary) *modelsv2.Fn {
	annotations := make(map[string]interface{})
	invokeEndpoint := fmt.Sprintf(invokeEndpointFmtString, *ociFnSummary.InvokeEndpoint, *ociFnSummary.Id)
	annotations[annotationCompartmentId] = *ociFnSummary.CompartmentId
	annotations[annotationImageDigest] = *ociFnSummary.ImageDigest
	annotations[annotationInvokeEndpoint] = invokeEndpoint

	var timeoutPtr *int32
	if ociFnSummary.TimeoutInSeconds != nil {
		timeout := int32(*ociFnSummary.TimeoutInSeconds)
		timeoutPtr = &timeout
	}

	return &modelsv2.Fn{
		Annotations: annotations,
		AppID:       *ociFnSummary.ApplicationId,
		CreatedAt:   strfmt.DateTime(ociFnSummary.TimeCreated.Time),
		ID:          *ociFnSummary.Id,
		Image:       *ociFnSummary.Image,
		Memory:      uint64(*ociFnSummary.MemoryInMBs),
		Name:        *ociFnSummary.DisplayName,
		Timeout:     timeoutPtr,
		UpdatedAt:   strfmt.DateTime(ociFnSummary.TimeUpdated.Time),
	}
}
