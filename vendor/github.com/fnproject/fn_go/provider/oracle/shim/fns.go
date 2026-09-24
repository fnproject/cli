package shim

import (
	"context"
	"fmt"
	"github.com/fnproject/fn_go/clientv2/fns"
	"github.com/fnproject/fn_go/modelsv2"
	"github.com/fnproject/fn_go/provider/oracle/shim/client"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/oracle/oci-go-sdk/v65/functions"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	defaultMemory int64 = 128 // MB

	annotationImageDigest            = "oracle.com/oci/imageDigest"
	annotationInvokeEndpoint         = "fnproject.io/fn/invokeEndpoint"
	annotationPCStrategy             = "oracle.com/oci/provisionedConcurrencyStrategy"
	annotationPCCount                = "oracle.com/oci/provisionedConcurrencyCount"
	annotationDetachedTimeoutSeconds = "oracle.com/oci/detachedModeTimeoutInSeconds"
	annotationSuccessDestinationKind = "oracle.com/oci/successDestinationKind"
	annotationSuccessDestinationOCID = "oracle.com/oci/successDestinationOcid"
	annotationFailureDestinationKind = "oracle.com/oci/failureDestinationKind"
	annotationFailureDestinationOCID = "oracle.com/oci/failureDestinationOcid"
	annotationSourceType             = "oracle.com/oci/sourceType"
	annotationPbfListingID           = "oracle.com/oci/pbfListingId"

	invokeEndpointFmtString = "%s/20181201/functions/%s/actions/invoke"
)

type fnsShim struct {
	ociClient                client.FunctionsManagementClient
	workRequestClientFactory func() (*functions.WorkRequestManagementClient, error)
}

var _ fns.ClientService = &fnsShim{}

func NewFnsShim(ociClient client.FunctionsManagementClient, workRequestClientFactory ...func() (*functions.WorkRequestManagementClient, error)) fns.ClientService {
	shim := &fnsShim{ociClient: ociClient}
	if len(workRequestClientFactory) > 0 {
		shim.workRequestClientFactory = workRequestClientFactory[0]
	}
	return shim
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
	pcConfig, err := parseProvisionedConcurrencyAnnotation(params.Body.Annotations)
	if err != nil {
		return nil, err
	}
	freeformTags, err := parseFreeformTagsAnnotation(params.Body.Annotations)
	if err != nil {
		return nil, err
	}
	definedTags, err := parseDefinedTagsAnnotation(params.Body.Annotations)
	if err != nil {
		return nil, err
	}

	sourceDetails, err := createFunctionSourceDetails(params.Body.Image, digest)
	if params.Body.CodeOnly {
		sourceDetails, err = createCodeOnlyFunctionSourceDetails(params.Body)
	}
	if err != nil {
		return nil, err
	}

	details := functions.CreateFunctionDetails{
		DisplayName:                  &params.Body.Name,
		ApplicationId:                &params.Body.AppID,
		MemoryInMBs:                  &memory,
		Config:                       params.Body.Config,
		TimeoutInSeconds:             parseTimeout(params.Body.Timeout),
		SourceDetails:                sourceDetails,
		ProvisionedConcurrencyConfig: pcConfig,
		FreeformTags:                 freeformTags,
		DefinedTags:                  definedTags,
	}
	if err := applyGeneratedOCIParityCreateFunctionDetails(&details, params.Body.Annotations); err != nil {
		return nil, err
	}
	if detachedTimeoutSeconds, err := parseDetachedTimeoutAnnotation(params.Body.Annotations); err != nil {
		return nil, err
	} else if detachedTimeoutSeconds != nil {
		details.DetachedModeTimeoutInSeconds = detachedTimeoutSeconds
	}
	if successDestination, failureDestination, err := parseDestinationAnnotations(params.Body.Annotations); err != nil {
		return nil, err
	} else {
		if successDestination != nil {
			details.SuccessDestination = successDestination
		}
		if failureDestination != nil {
			details.FailureDestination = failureDestination
		}
	}

	req := functions.CreateFunctionRequest{CreateFunctionDetails: details}

	res, err := s.ociClient.CreateFunction(ctxOrBackground(params.Context), req)
	if err != nil {
		return nil, err
	}
	if err := s.waitForFunctionLifecycleWorkRequest(ctxOrBackground(params.Context), res.OpcWorkRequestId); err != nil {
		return nil, err
	}

	return &fns.CreateFnOK{
		Payload: ociFnToV2(res.Function),
	}, nil
}

func (s *fnsShim) DeleteFn(params *fns.DeleteFnParams) (*fns.DeleteFnNoContent, error) {
	req := functions.DeleteFunctionRequest{FunctionId: &params.FnID}

	res, err := s.ociClient.DeleteFunction(ctxOrBackground(params.Context), req)
	if err != nil {
		return nil, err
	}
	if err := s.waitForFunctionLifecycleWorkRequest(ctxOrBackground(params.Context), res.OpcWorkRequestId); err != nil {
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
	freeformTags, err := parseFreeformTagsAnnotation(params.Body.Annotations)
	if err != nil {
		return nil, err
	}
	definedTags, err := parseDefinedTagsAnnotation(params.Body.Annotations)
	if err != nil {
		return nil, err
	}

	var updateSourceDetails functions.UpdateFunctionSourceDetails
	if params.Body.CodeOnly {
		updateSourceDetails, err = createCodeOnlyUpdateSourceDetails(params.Body)
		if err != nil {
			return nil, err
		}
	} else if imagePtr != nil || digest != nil {
		updateSourceDetails = functions.UpdateContainerImageFunctionSourceDetails{
			Image:       imagePtr,
			ImageDigest: digest,
		}
	}

	details := functions.UpdateFunctionDetails{
		MemoryInMBs:      memoryPtr,
		Config:           params.Body.Config,
		TimeoutInSeconds: parseTimeout(params.Body.Timeout),
		SourceDetails:    updateSourceDetails,
		FreeformTags:     freeformTags,
		DefinedTags:      definedTags,
	}
	if err := applyGeneratedOCIParityUpdateFunctionDetails(&details, params.Body.Annotations); err != nil {
		return nil, err
	}
	if detachedTimeoutSeconds, err := parseDetachedTimeoutAnnotation(params.Body.Annotations); err != nil {
		return nil, err
	} else if detachedTimeoutSeconds != nil {
		details.DetachedModeTimeoutInSeconds = detachedTimeoutSeconds
	}
	if successDestination, failureDestination, err := parseDestinationAnnotations(params.Body.Annotations); err != nil {
		return nil, err
	} else {
		if successDestination != nil {
			details.SuccessDestination = successDestination
		}
		if failureDestination != nil {
			details.FailureDestination = failureDestination
		}
	}

	req := functions.UpdateFunctionRequest{
		FunctionId:            &params.FnID,
		UpdateFunctionDetails: details,
		IfMatch:               etag,
	}

	updateRes, err := s.ociClient.UpdateFunction(ctxOrBackground(params.Context), req)
	if err != nil {
		return nil, err
	}
	if err := s.waitForFunctionLifecycleWorkRequest(ctxOrBackground(params.Context), updateRes.OpcWorkRequestId); err != nil {
		return nil, err
	}
	getRes, err := s.ociClient.GetFunction(ctxOrBackground(params.Context), functions.GetFunctionRequest{FunctionId: &params.FnID})
	if err != nil {
		return nil, err
	}

	return &fns.UpdateFnOK{
		Payload: ociFnToV2(getRes.Function),
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

func parseProvisionedConcurrencyAnnotation(annotations map[string]interface{}) (functions.FunctionProvisionedConcurrencyConfig, error) {
	if annotations == nil || len(annotations) == 0 {
		return nil, nil
	}
	strategyRaw, ok := annotations[annotationPCStrategy]
	if !ok {
		return nil, nil
	}
	strategy, ok := strategyRaw.(string)
	if !ok {
		return nil, fmt.Errorf("invalid provisioned concurrency strategy")
	}
	switch strings.ToUpper(strings.TrimSpace(strategy)) {
	case "NONE":
		return functions.NoneProvisionedConcurrencyConfig{}, nil
	case "CONSTANT":
		countRaw, ok := annotations[annotationPCCount]
		if !ok {
			return nil, fmt.Errorf("invalid provisioned concurrency count")
		}
		var count int
		switch typed := countRaw.(type) {
		case int:
			count = typed
		case int32:
			count = int(typed)
		case int64:
			count = int(typed)
		case float64:
			count = int(typed)
		default:
			return nil, fmt.Errorf("invalid provisioned concurrency count")
		}
		return functions.ConstantProvisionedConcurrencyConfig{Count: &count}, nil
	default:
		return nil, fmt.Errorf("invalid provisioned concurrency strategy")
	}
}

func parseDetachedTimeoutAnnotation(annotations map[string]interface{}) (*int, error) {
	if annotations == nil || len(annotations) == 0 {
		return nil, nil
	}
	raw, ok := annotations[annotationDetachedTimeoutSeconds]
	if !ok {
		return nil, nil
	}
	switch typed := raw.(type) {
	case int:
		return &typed, nil
	case int32:
		value := int(typed)
		return &value, nil
	case int64:
		value := int(typed)
		return &value, nil
	case float64:
		value := int(typed)
		return &value, nil
	case string:
		value, err := strconv.Atoi(typed)
		if err != nil {
			return nil, fmt.Errorf("invalid detached timeout annotation")
		}
		return &value, nil
	default:
		return nil, fmt.Errorf("invalid detached timeout annotation")
	}
}

func parseDestinationAnnotations(annotations map[string]interface{}) (functions.SuccessDestinationDetails, functions.FailureDestinationDetails, error) {
	if annotations == nil || len(annotations) == 0 {
		return nil, nil, nil
	}
	var success functions.SuccessDestinationDetails
	var failure functions.FailureDestinationDetails
	if kindRaw, ok := annotations[annotationSuccessDestinationKind]; ok {
		kind, ok := kindRaw.(string)
		if !ok {
			return nil, nil, fmt.Errorf("invalid success destination kind")
		}
		ocidRaw, ok := annotations[annotationSuccessDestinationOCID]
		if !ok {
			return nil, nil, fmt.Errorf("invalid success destination ocid")
		}
		ocid, ok := ocidRaw.(string)
		if !ok {
			return nil, nil, fmt.Errorf("invalid success destination ocid")
		}
		var err error
		success, err = parseSuccessDestination(strings.ToUpper(strings.TrimSpace(kind)), ocid)
		if err != nil {
			return nil, nil, err
		}
	}
	if kindRaw, ok := annotations[annotationFailureDestinationKind]; ok {
		kind, ok := kindRaw.(string)
		if !ok {
			return nil, nil, fmt.Errorf("invalid failure destination kind")
		}
		ocidRaw, ok := annotations[annotationFailureDestinationOCID]
		if !ok {
			return nil, nil, fmt.Errorf("invalid failure destination ocid")
		}
		ocid, ok := ocidRaw.(string)
		if !ok {
			return nil, nil, fmt.Errorf("invalid failure destination ocid")
		}
		var err error
		failure, err = parseFailureDestination(strings.ToUpper(strings.TrimSpace(kind)), ocid)
		if err != nil {
			return nil, nil, err
		}
	}
	return success, failure, nil
}

func parseSuccessDestination(kind, ocid string) (functions.SuccessDestinationDetails, error) {
	switch kind {
	case "STREAM":
		return functions.StreamSuccessDestinationDetails{StreamId: &ocid}, nil
	case "QUEUE":
		return functions.QueueSuccessDestinationDetails{QueueId: &ocid}, nil
	case "NOTIFICATIONS", "NOTIFICATION":
		return functions.NotificationSuccessDestinationDetails{TopicId: &ocid}, nil
	case "NONE":
		return functions.NoneSuccessDestinationDetails{}, nil
	default:
		return nil, fmt.Errorf("invalid success destination kind %q", kind)
	}
}

func parseFailureDestination(kind, ocid string) (functions.FailureDestinationDetails, error) {
	switch kind {
	case "STREAM":
		return functions.StreamFailureDestinationDetails{StreamId: &ocid}, nil
	case "QUEUE":
		return functions.QueueFailureDestinationDetails{QueueId: &ocid}, nil
	case "NOTIFICATIONS", "NOTIFICATION":
		return functions.NotificationFailureDestinationDetails{TopicId: &ocid}, nil
	case "NONE":
		return functions.NoneFailureDestinationDetails{}, nil
	default:
		return nil, fmt.Errorf("invalid failure destination kind %q", kind)
	}
}

func addProvisionedConcurrencyAnnotations(annotations map[string]interface{}, cfg functions.FunctionProvisionedConcurrencyConfig) {
	strategy := "NONE"
	var count *int
	switch typed := cfg.(type) {
	case functions.ConstantProvisionedConcurrencyConfig:
		strategy, count = "CONSTANT", typed.Count
	case functions.NoneProvisionedConcurrencyConfig, nil:
		// NONE is the default representation for absent configuration.
	}
	annotations[annotationPCStrategy] = strategy
	if count != nil {
		annotations[annotationPCCount] = *count
	}
}

func ociFnToV2(ociFn functions.Function) *modelsv2.Fn {
	annotations := make(map[string]interface{})
	invokeEndpoint := fmt.Sprintf(invokeEndpointFmtString, *ociFn.InvokeEndpoint, *ociFn.Id)
	annotations[annotationCompartmentId] = *ociFn.CompartmentId

	image, imageDigest := imageFromSourceDetails(ociFn.SourceDetails)

	annotations[annotationImageDigest] = imageDigest
	annotations[annotationInvokeEndpoint] = invokeEndpoint
	addProvisionedConcurrencyAnnotations(annotations, ociFn.ProvisionedConcurrencyConfig)
	addTagAnnotations(annotations, ociFn.FreeformTags, ociFn.DefinedTags)
	addSourceDetailsAnnotations(annotations, ociFn.SourceDetails)
	addTraceConfigAnnotation(annotations, ociFn.TraceConfig)
	if ociFn.DetachedModeTimeoutInSeconds != nil {
		annotations[annotationDetachedTimeoutSeconds] = *ociFn.DetachedModeTimeoutInSeconds
	}
	addDestinationAnnotations(annotations, ociFn.SuccessDestination, ociFn.FailureDestination)

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
		Image:       image,
		Memory:      uint64(*ociFn.MemoryInMBs),
		Name:        *ociFn.DisplayName,
		Timeout:     timeoutPtr,
		Shape:       string(ociFn.Shape),
		UpdatedAt:   strfmt.DateTime(ociFn.TimeUpdated.Time),
	}
}

func ociFnSummaryToV2(ociFnSummary functions.FunctionSummary) *modelsv2.Fn {
	annotations := make(map[string]interface{})
	invokeEndpoint := fmt.Sprintf(invokeEndpointFmtString, *ociFnSummary.InvokeEndpoint, *ociFnSummary.Id)
	annotations[annotationCompartmentId] = *ociFnSummary.CompartmentId

	image, imageDigest := imageFromSourceDetails(ociFnSummary.SourceDetails)

	annotations[annotationImageDigest] = imageDigest
	annotations[annotationInvokeEndpoint] = invokeEndpoint
	addProvisionedConcurrencyAnnotations(annotations, ociFnSummary.ProvisionedConcurrencyConfig)
	addTagAnnotations(annotations, ociFnSummary.FreeformTags, ociFnSummary.DefinedTags)
	addSourceDetailsAnnotations(annotations, ociFnSummary.SourceDetails)
	addTraceConfigAnnotation(annotations, ociFnSummary.TraceConfig)
	if ociFnSummary.DetachedModeTimeoutInSeconds != nil {
		annotations[annotationDetachedTimeoutSeconds] = *ociFnSummary.DetachedModeTimeoutInSeconds
	}
	addDestinationAnnotations(annotations, ociFnSummary.SuccessDestination, ociFnSummary.FailureDestination)

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
		Image:       image,
		Memory:      uint64(*ociFnSummary.MemoryInMBs),
		Name:        *ociFnSummary.DisplayName,
		Shape:       string(ociFnSummary.Shape),
		Timeout:     timeoutPtr,
		UpdatedAt:   strfmt.DateTime(ociFnSummary.TimeUpdated.Time),
	}
}

func createFunctionSourceDetails(image string, digest *string) (functions.CreateFunctionSourceDetails, error) {
	if image == "" {
		return functions.CreateContainerImageFunctionSourceDetails{}, nil
	}
	return functions.CreateContainerImageFunctionSourceDetails{Image: &image, ImageDigest: digest}, nil
}

func createCodeOnlyFunctionSourceDetails(fn *modelsv2.Fn) (functions.CreateFunctionSourceDetails, error) {
	archiveSource, err := createArchiveSourceDetails(fn)
	if err != nil {
		return nil, err
	}
	runtimeConfig, err := createRuntimeConfig(fn)
	if err != nil {
		return nil, err
	}
	var handler *string
	if strings.TrimSpace(fn.Handler) != "" {
		h := strings.TrimSpace(fn.Handler)
		handler = &h
	}
	return functions.CreateArchiveFunctionSourceDetails{
		ArchiveSourceDetails: archiveSource,
		RuntimeConfig:        runtimeConfig,
		Handler:              handler,
	}, nil
}

func createArchiveSourceDetails(fn *modelsv2.Fn) (functions.CreateArchiveSourceDetails, error) {
	switch strings.ToLower(strings.TrimSpace(fn.SourceType)) {
	case "direct":
		archive, err := readArchiveSource(fn)
		if err != nil {
			return nil, err
		}
		if len(archive) == 0 {
			return nil, fmt.Errorf("direct source requires archive bytes")
		}
		return functions.CreateDirectArchiveSourceDetails{ArchiveFile: archive}, nil
	case "object-storage", "object_storage":
		bucket := strings.TrimSpace(fn.SourceBucketName)
		namespace := strings.TrimSpace(fn.SourceNamespace)
		objectName := strings.TrimSpace(fn.SourceObjectName)
		if bucket == "" || namespace == "" || objectName == "" {
			return nil, fmt.Errorf("object-storage source requires bucket, namespace, and object name")
		}
		details := functions.CreateObjectStorageArchiveSourceDetails{
			BucketName: &bucket,
			Namespace:  &namespace,
			ObjectName: &objectName,
		}
		if version := strings.TrimSpace(fn.SourceObjectVersionID); version != "" {
			details.ObjectVersionId = &version
		}
		return details, nil
	default:
		return nil, fmt.Errorf("unsupported code-only source type %q", fn.SourceType)
	}
}

func createRuntimeConfig(fn *modelsv2.Fn) (functions.CreateRuntimeConfig, error) {
	switch strings.ToUpper(strings.TrimSpace(fn.RuntimeConfigType)) {
	case "FUNCTION_UPDATE":
		runtimeName := strings.TrimSpace(fn.RuntimeName)
		return functions.CreateFunctionUpdateRuntimeConfig{FunctionsRuntimeName: &runtimeName}, nil
	case "MANUAL":
		runtimeName := strings.TrimSpace(fn.RuntimeName)
		runtimeVersionID := strings.TrimSpace(fn.RuntimeVersionID)
		return functions.CreateManualRuntimeConfig{
			FunctionsRuntimeName:      &runtimeName,
			FunctionsRuntimeVersionId: &runtimeVersionID,
		}, nil
	default:
		return nil, fmt.Errorf("unsupported runtime config type %q", fn.RuntimeConfigType)
	}
}

func createCodeOnlyUpdateSourceDetails(fn *modelsv2.Fn) (functions.UpdateFunctionSourceDetails, error) {
	var archiveSource functions.UpdateArchiveSourceDetails
	var err error
	if strings.TrimSpace(fn.SourceType) != "" {
		archiveSource, err = createUpdateArchiveSourceDetails(fn)
		if err != nil {
			return nil, err
		}
	}
	var runtimeConfig functions.UpdateRuntimeConfig
	if strings.TrimSpace(fn.RuntimeConfigType) != "" {
		runtimeConfig, err = createUpdateRuntimeConfig(fn)
		if err != nil {
			return nil, err
		}
	}
	var handler *string
	if strings.TrimSpace(fn.Handler) != "" {
		h := strings.TrimSpace(fn.Handler)
		handler = &h
	}
	return functions.UpdateArchiveFunctionSourceDetails{
		ArchiveSourceDetails: archiveSource,
		Handler:              handler,
		RuntimeConfig:        runtimeConfig,
	}, nil
}

func createUpdateArchiveSourceDetails(fn *modelsv2.Fn) (functions.UpdateArchiveSourceDetails, error) {
	switch strings.ToLower(strings.TrimSpace(fn.SourceType)) {
	case "direct":
		archive, err := readArchiveSource(fn)
		if err != nil {
			return nil, err
		}
		if len(archive) == 0 {
			return nil, fmt.Errorf("direct source requires archive bytes")
		}
		return functions.UpdateDirectArchiveSourceDetails{ArchiveFile: archive}, nil
	case "object-storage", "object_storage":
		details := functions.UpdateObjectStorageArchiveSourceDetails{}
		if value := strings.TrimSpace(fn.SourceBucketName); value != "" {
			details.BucketName = &value
		}
		if value := strings.TrimSpace(fn.SourceNamespace); value != "" {
			details.Namespace = &value
		}
		if value := strings.TrimSpace(fn.SourceObjectName); value != "" {
			details.ObjectName = &value
		}
		if value := strings.TrimSpace(fn.SourceObjectVersionID); value != "" {
			details.ObjectVersionId = &value
		}
		return details, nil
	default:
		return nil, fmt.Errorf("unsupported code-only source type %q", fn.SourceType)
	}
}

func createUpdateRuntimeConfig(fn *modelsv2.Fn) (functions.UpdateRuntimeConfig, error) {
	switch strings.ToUpper(strings.TrimSpace(fn.RuntimeConfigType)) {
	case "FUNCTION_UPDATE":
		runtimeName := strings.TrimSpace(fn.RuntimeName)
		return functions.UpdateFunctionUpdateRuntimeConfig{FunctionsRuntimeName: &runtimeName}, nil
	case "MANUAL":
		runtimeName := strings.TrimSpace(fn.RuntimeName)
		runtimeVersionID := strings.TrimSpace(fn.RuntimeVersionID)
		cfg := functions.UpdateManualRuntimeConfig{FunctionsRuntimeVersionId: &runtimeVersionID}
		if runtimeName != "" {
			cfg.FunctionsRuntimeName = &runtimeName
		}
		return cfg, nil
	default:
		return nil, fmt.Errorf("unsupported runtime config type %q", fn.RuntimeConfigType)
	}
}

func readArchiveSource(fn *modelsv2.Fn) ([]byte, error) {
	if len(fn.SourceArchive) > 0 {
		return []byte(fn.SourceArchive), nil
	}
	sourceFile := strings.TrimSpace(fn.SourceFile)
	if sourceFile == "" {
		return nil, nil
	}
	archive, err := os.ReadFile(sourceFile)
	if err != nil {
		return nil, fmt.Errorf("read direct source archive %q: %w", sourceFile, err)
	}
	return archive, nil
}

func addDestinationAnnotations(annotations map[string]interface{}, success functions.SuccessDestinationDetails, failure functions.FailureDestinationDetails) {
	if annotations == nil {
		return
	}
	if success != nil {
		switch typed := success.(type) {
		case functions.StreamSuccessDestinationDetails:
			annotations[annotationSuccessDestinationKind] = "STREAM"
			if typed.StreamId != nil {
				annotations[annotationSuccessDestinationOCID] = *typed.StreamId
			}
		case functions.QueueSuccessDestinationDetails:
			annotations[annotationSuccessDestinationKind] = "QUEUE"
			if typed.QueueId != nil {
				annotations[annotationSuccessDestinationOCID] = *typed.QueueId
			}
		case functions.NotificationSuccessDestinationDetails:
			annotations[annotationSuccessDestinationKind] = "NOTIFICATIONS"
			if typed.TopicId != nil {
				annotations[annotationSuccessDestinationOCID] = *typed.TopicId
			}
		}
	}
	if failure != nil {
		switch typed := failure.(type) {
		case functions.StreamFailureDestinationDetails:
			annotations[annotationFailureDestinationKind] = "STREAM"
			if typed.StreamId != nil {
				annotations[annotationFailureDestinationOCID] = *typed.StreamId
			}
		case functions.QueueFailureDestinationDetails:
			annotations[annotationFailureDestinationKind] = "QUEUE"
			if typed.QueueId != nil {
				annotations[annotationFailureDestinationOCID] = *typed.QueueId
			}
		case functions.NotificationFailureDestinationDetails:
			annotations[annotationFailureDestinationKind] = "NOTIFICATIONS"
			if typed.TopicId != nil {
				annotations[annotationFailureDestinationOCID] = *typed.TopicId
			}
		}
	}
}

func addSourceDetailsAnnotations(annotations map[string]interface{}, sourceDetails functions.FunctionSourceDetails) {
	if annotations == nil || sourceDetails == nil {
		return
	}
	switch typed := sourceDetails.(type) {
	case functions.PreBuiltFunctionSourceDetails:
		annotations[annotationSourceType] = "PRE_BUILT_FUNCTIONS"
		if typed.PbfListingId != nil {
			annotations[annotationPbfListingID] = *typed.PbfListingId
		}
	}
}

func addTraceConfigAnnotation(annotations map[string]interface{}, traceConfig *functions.FunctionTraceConfig) {
	if annotations == nil || traceConfig == nil {
		return
	}
	trace := map[string]interface{}{}
	if traceConfig.IsEnabled != nil {
		trace["isEnabled"] = *traceConfig.IsEnabled
	}
	if len(trace) > 0 {
		annotations[annotationOCIParityFnTraceConfig] = trace
	}
}

func imageFromSourceDetails(details functions.FunctionSourceDetails) (string, string) {
	if details == nil {
		return "", ""
	}
	if container, ok := details.(functions.ContainerImageFunctionSourceDetails); ok {
		image := ""
		if container.Image != nil {
			image = *container.Image
		}
		digest := ""
		if container.ImageDigest != nil {
			digest = *container.ImageDigest
		}
		return image, digest
	}
	return "", ""
}

func (s *fnsShim) waitForFunctionLifecycleWorkRequest(ctx context.Context, workRequestID *string) error {
	if workRequestID == nil || strings.TrimSpace(*workRequestID) == "" {
		return nil
	}
	wrClient, err := s.newWorkRequestManagementClient()
	if err != nil || wrClient == nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "Tracking work request %s\n", *workRequestID)
	lastStatus := functions.OperationStatusEnum("")
	lastPercent := float32(-1)

	for {
		resp, err := wrClient.GetWorkRequest(ctx, functions.GetWorkRequestRequest{WorkRequestId: workRequestID})
		if err != nil {
			return err
		}
		if resp.Status != lastStatus || (resp.PercentComplete != nil && *resp.PercentComplete != lastPercent) {
			percent := float32(0)
			if resp.PercentComplete != nil {
				percent = *resp.PercentComplete
			}
			fmt.Fprintf(os.Stderr, "Work request %s status: %s (%.0f%%)\n", *workRequestID, resp.Status, percent)
			lastStatus = resp.Status
			lastPercent = percent
		}

		switch resp.Status {
		case functions.OperationStatusAccepted,
			functions.OperationStatusInProgress,
			functions.OperationStatusWaiting,
			functions.OperationStatusCanceling:
			wait := 2 * time.Second
			if resp.RetryAfter != nil && *resp.RetryAfter > 0 {
				wait = time.Duration(*resp.RetryAfter) * time.Second
			}
			time.Sleep(wait)
			continue
		case functions.OperationStatusSucceeded:
			fmt.Fprintf(os.Stderr, "Work request %s completed successfully\n", *workRequestID)
			return nil
		case functions.OperationStatusFailed,
			functions.OperationStatusCanceled,
			functions.OperationStatusNeedsAttention:
			return s.workRequestFailure(ctx, wrClient, workRequestID, resp.Status)
		default:
			return fmt.Errorf("work request %s ended in unexpected status %s", *workRequestID, resp.Status)
		}
	}
}

func (s *fnsShim) newWorkRequestManagementClient() (*functions.WorkRequestManagementClient, error) {
	if s.workRequestClientFactory == nil {
		return nil, nil
	}
	return s.workRequestClientFactory()
}

func (s *fnsShim) workRequestFailure(ctx context.Context, wrClient *functions.WorkRequestManagementClient, workRequestID *string, status functions.OperationStatusEnum) error {
	message := fmt.Sprintf("work request %s ended with status %s", *workRequestID, status)
	if wrClient == nil {
		return fmt.Errorf("%s", message)
	}
	errResp, err := wrClient.ListWorkRequestErrors(ctx, functions.ListWorkRequestErrorsRequest{WorkRequestId: workRequestID})
	if err != nil || len(errResp.Items) == 0 || errResp.Items[0].Message == nil {
		return fmt.Errorf("%s", message)
	}
	return fmt.Errorf("%s: %s", message, *errResp.Items[0].Message)
}
