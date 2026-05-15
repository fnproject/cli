package shim

import (
	"fmt"
	"github.com/fnproject/fn_go/clientv2/fns"
	"github.com/fnproject/fn_go/modelsv2"
	"github.com/fnproject/fn_go/provider/oracle/shim/client"
	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"
	"github.com/oracle/oci-go-sdk/v65/functions"
	"strconv"
	"strings"
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
	sourceDetails, err := parseSourceDetailsAnnotation(params.Body.Annotations)
	if err != nil {
		return nil, err
	}
	var imagePtr *string
	if params.Body.Image != "" {
		imagePtr = &params.Body.Image
	}

	details := functions.CreateFunctionDetails{
		DisplayName:                  &params.Body.Name,
		ApplicationId:                &params.Body.AppID,
		Image:                        imagePtr,
		MemoryInMBs:                  &memory,
		ImageDigest:                  digest,
		SourceDetails:                sourceDetails,
		ProvisionedConcurrencyConfig: pcConfig,
		Config:                       params.Body.Config,
		FreeformTags:                 freeformTags,
		DefinedTags:                  definedTags,
		TimeoutInSeconds:             parseTimeout(params.Body.Timeout),
	}
	if detachedTimeoutSeconds, err := parseDetachedTimeoutAnnotation(params.Body.Annotations); err != nil {
		return nil, err
	} else if detachedTimeoutSeconds != nil {
		details.DetachedModeTimeoutInSeconds = detachedTimeoutSeconds
	}
	if successDestination, failureDestination, err := parseDestinationAnnotations(params.Body.Annotations); err != nil {
		return nil, err
	} else {
		details.SuccessDestination = successDestination
		details.FailureDestination = failureDestination
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
	freeformTags, err := parseFreeformTagsAnnotation(params.Body.Annotations)
	if err != nil {
		return nil, err
	}
	definedTags, err := parseDefinedTagsAnnotation(params.Body.Annotations)
	if err != nil {
		return nil, err
	}

	details := functions.UpdateFunctionDetails{
		Image:            imagePtr,
		ImageDigest:      digest,
		MemoryInMBs:      memoryPtr,
		Config:           params.Body.Config,
		FreeformTags:     freeformTags,
		DefinedTags:      definedTags,
		TimeoutInSeconds: parseTimeout(params.Body.Timeout),
	}
	if detachedTimeoutSeconds, err := parseDetachedTimeoutAnnotation(params.Body.Annotations); err != nil {
		return nil, err
	} else if detachedTimeoutSeconds != nil {
		details.DetachedModeTimeoutInSeconds = detachedTimeoutSeconds
	}
	if successDestination, failureDestination, err := parseDestinationAnnotations(params.Body.Annotations); err != nil {
		return nil, err
	} else {
		details.SuccessDestination = successDestination
		details.FailureDestination = failureDestination
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
		v := int(typed)
		return &v, nil
	case int64:
		v := int(typed)
		return &v, nil
	case float64:
		v := int(typed)
		return &v, nil
	case string:
		v, err := strconv.Atoi(typed)
		if err != nil {
			return nil, fmt.Errorf("invalid detached timeout annotation")
		}
		return &v, nil
	default:
		return nil, fmt.Errorf("invalid detached timeout annotation")
	}
}

func parseDestinationAnnotations(annotations map[string]interface{}) (functions.SuccessDestinationDetails, functions.FailureDestinationDetails, error) {
	var success functions.SuccessDestinationDetails
	var failure functions.FailureDestinationDetails
	if annotations == nil || len(annotations) == 0 {
		return nil, nil, nil
	}
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
		s, err := parseSuccessDestination(strings.ToUpper(strings.TrimSpace(kind)), ocid)
		if err != nil {
			return nil, nil, err
		}
		success = s
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
		f, err := parseFailureDestination(strings.ToUpper(strings.TrimSpace(kind)), ocid)
		if err != nil {
			return nil, nil, err
		}
		failure = f
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

func parseSourceDetailsAnnotation(annotations map[string]interface{}) (functions.FunctionSourceDetails, error) {
	if annotations == nil || len(annotations) == 0 {
		return nil, nil
	}
	rawType, ok := annotations[annotationSourceType]
	if !ok {
		return nil, nil
	}
	sourceType, ok := rawType.(string)
	if !ok {
		return nil, fmt.Errorf("invalid function source type annotation")
	}
	switch strings.ToUpper(strings.TrimSpace(sourceType)) {
	case "PRE_BUILT_FUNCTIONS":
		rawListingID, ok := annotations[annotationPbfListingID]
		if !ok {
			return nil, fmt.Errorf("invalid pbf listing annotation")
		}
		listingID, ok := rawListingID.(string)
		if !ok || strings.TrimSpace(listingID) == "" {
			return nil, fmt.Errorf("invalid pbf listing annotation")
		}
		return functions.PreBuiltFunctionSourceDetails{PbfListingId: &listingID}, nil
	default:
		return nil, fmt.Errorf("unsupported function source type %q", sourceType)
	}
}

func addProvisionedConcurrencyAnnotations(annotations map[string]interface{}, cfg functions.FunctionProvisionedConcurrencyConfig) {
	strategy := "NONE"
	var count *int

	switch typed := cfg.(type) {
	case functions.ConstantProvisionedConcurrencyConfig:
		strategy = "CONSTANT"
		count = typed.Count
	case functions.NoneProvisionedConcurrencyConfig:
		strategy = "NONE"
	case nil:
		strategy = "NONE"
	default:
		strategy = "NONE"
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

	// For pbf functions image and its digest will be always empty
	imageDigest := ""
	if ociFn.ImageDigest != nil {
		imageDigest = *ociFn.ImageDigest
	}

	image := ""
	if ociFn.Image != nil {
		image = *ociFn.Image
	}

	annotations[annotationImageDigest] = imageDigest
	annotations[annotationInvokeEndpoint] = invokeEndpoint
	addProvisionedConcurrencyAnnotations(annotations, ociFn.ProvisionedConcurrencyConfig)
	addTagAnnotations(annotations, ociFn.FreeformTags, ociFn.DefinedTags)
	addSourceDetailsAnnotations(annotations, ociFn.SourceDetails)
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

	// For pbf functions image and its digest will be always empty
	imageDigest := ""
	if ociFnSummary.ImageDigest != nil {
		imageDigest = *ociFnSummary.ImageDigest
	}

	image := ""
	if ociFnSummary.Image != nil {
		image = *ociFnSummary.Image
	}

	annotations[annotationImageDigest] = imageDigest
	annotations[annotationInvokeEndpoint] = invokeEndpoint
	addProvisionedConcurrencyAnnotations(annotations, ociFnSummary.ProvisionedConcurrencyConfig)
	addTagAnnotations(annotations, ociFnSummary.FreeformTags, ociFnSummary.DefinedTags)
	addSourceDetailsAnnotations(annotations, ociFnSummary.SourceDetails)
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
