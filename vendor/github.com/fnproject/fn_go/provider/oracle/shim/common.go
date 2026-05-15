package shim

import (
	"context"
	"fmt"
)

const annotationCompartmentId = "oracle.com/oci/compartmentId"

const (
	annotationFreeformTags = "oracle.com/oci/freeformTags"
	annotationDefinedTags  = "oracle.com/oci/definedTags"
)

// OCI update config is wholesale replacement of the map. Here we do the FnV2 server-side merge on the client instead.
// Based on https://github.com/fnproject/fn/blob/d55e01ab7d565e9796748f2f40662e94394aff07/api/models/fn.go#L274-L285
func mergeConfig(oldConfig map[string]string, changeConfig map[string]string) map[string]string {
	if changeConfig != nil {
		if oldConfig == nil {
			oldConfig = make(map[string]string)
		}
		for k, v := range changeConfig {
			if v == "" {
				delete(oldConfig, k)
			} else {
				oldConfig[k] = v
			}
		}
	}
	return oldConfig
}

// Helper func to convert nil context to context.Background
func ctxOrBackground(ctx context.Context) context.Context {
	if ctx == nil {
		return context.Background()
	}
	return ctx
}

func parseFreeformTagsAnnotation(annotations map[string]interface{}) (map[string]string, error) {
	if annotations == nil || len(annotations) == 0 {
		return nil, nil
	}
	raw, ok := annotations[annotationFreeformTags]
	if !ok {
		return nil, nil
	}
	converted := map[string]string{}
	switch typed := raw.(type) {
	case map[string]string:
		for key, value := range typed {
			converted[key] = value
		}
	case map[string]interface{}:
		for key, value := range typed {
			converted[key] = fmt.Sprint(value)
		}
	default:
		return nil, fmt.Errorf("invalid freeform tags annotation")
	}
	return converted, nil
}

func parseDefinedTagsAnnotation(annotations map[string]interface{}) (map[string]map[string]interface{}, error) {
	if annotations == nil || len(annotations) == 0 {
		return nil, nil
	}
	raw, ok := annotations[annotationDefinedTags]
	if !ok {
		return nil, nil
	}
	converted := map[string]map[string]interface{}{}
	switch typed := raw.(type) {
	case map[string]map[string]interface{}:
		return typed, nil
	case map[string]interface{}:
		for namespace, values := range typed {
			m, ok := values.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid defined tags annotation")
			}
			converted[namespace] = m
		}
		return converted, nil
	default:
		return nil, fmt.Errorf("invalid defined tags annotation")
	}
}

func addTagAnnotations(annotations map[string]interface{}, freeform map[string]string, defined map[string]map[string]interface{}) {
	if annotations == nil {
		return
	}
	if len(freeform) > 0 {
		value := map[string]interface{}{}
		for key, tagValue := range freeform {
			value[key] = tagValue
		}
		annotations[annotationFreeformTags] = value
	}
	if len(defined) > 0 {
		annotations[annotationDefinedTags] = defined
	}
}
