package common

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

const (
	AnnotationOCIResourceFreeformTags = "oracle.com/oci/freeformTags"
	AnnotationOCIResourceDefinedTags  = "oracle.com/oci/definedTags"
)

// OCIDefinedTags stores user-friendly defined tag values for func.yaml persistence.
// Values may be strings, numbers, booleans, objects, arrays, or null.
type OCIDefinedTags map[string]map[string]interface{}

func ParseFreeformTagSpecs(specs []string) (map[string]string, error) {
	if len(specs) == 0 {
		return nil, nil
	}
	tags := make(map[string]string)
	for _, spec := range specs {
		parts := strings.SplitN(strings.TrimSpace(spec), "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid tag %q (expected key=value)", spec)
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			return nil, fmt.Errorf("invalid tag %q (key must not be empty)", spec)
		}
		tags[key] = value
	}
	return tags, nil
}

func ParseDefinedTagSpecs(specs []string) (OCIDefinedTags, error) {
	if len(specs) == 0 {
		return nil, nil
	}
	tags := make(OCIDefinedTags)
	for _, spec := range specs {
		parts := strings.SplitN(strings.TrimSpace(spec), "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid defined tag %q (expected namespace.key=value)", spec)
		}
		qualifier := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		nsKey := strings.SplitN(qualifier, ".", 2)
		if len(nsKey) != 2 || strings.TrimSpace(nsKey[0]) == "" || strings.TrimSpace(nsKey[1]) == "" {
			return nil, fmt.Errorf("invalid defined tag %q (expected namespace.key=value)", spec)
		}
		namespace := strings.TrimSpace(nsKey[0])
		key := strings.TrimSpace(nsKey[1])
		if tags[namespace] == nil {
			tags[namespace] = make(map[string]interface{})
		}
		tags[namespace][key] = parseDefinedTagValue(value)
	}
	return tags, nil
}

func ConvertDefinedTagsToInterface(tags OCIDefinedTags) map[string]map[string]interface{} {
	if len(tags) == 0 {
		return nil
	}
	converted := make(map[string]map[string]interface{}, len(tags))
	for namespace, values := range tags {
		converted[namespace] = make(map[string]interface{}, len(values))
		for key, value := range values {
			converted[namespace][key] = value
		}
	}
	return converted
}

func ConvertDefinedTagsFromInterface(tags map[string]map[string]interface{}) OCIDefinedTags {
	if len(tags) == 0 {
		return nil
	}
	converted := make(OCIDefinedTags, len(tags))
	for namespace, values := range tags {
		converted[namespace] = make(map[string]interface{}, len(values))
		for key, value := range values {
			converted[namespace][key] = value
		}
	}
	return converted
}

func parseDefinedTagValue(value string) interface{} {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if !shouldParseDefinedTagValueAsJSON(trimmed) {
		return value
	}
	var decoded interface{}
	if err := json.Unmarshal([]byte(trimmed), &decoded); err == nil {
		return decoded
	}
	return value
}

func shouldParseDefinedTagValueAsJSON(value string) bool {
	if value == "" {
		return false
	}
	switch value[0] {
	case '{', '[', '"':
		return true
	default:
		return false
	}
}

func SetOCIResourceTagsOnFuncFile(ff *FuncFileV20180708, freeform map[string]string, defined OCIDefinedTags) {
	if ff == nil {
		return
	}
	if len(freeform) == 0 && len(defined) == 0 {
		return
	}
	if ff.Deploy == nil {
		ff.Deploy = &FuncDeployConfig{}
	}
	if ff.Deploy.OCI == nil {
		ff.Deploy.OCI = &OCIFunctionDeployConfig{}
	}
	if len(freeform) > 0 {
		ff.Deploy.OCI.FreeformTags = freeform
	}
	if len(defined) > 0 {
		ff.Deploy.OCI.DefinedTags = defined
	}
}

func OCIManagedResourceTagSettingNames(ff *FuncFileV20180708) []string {
	if ff == nil || ff.Deploy == nil || ff.Deploy.OCI == nil {
		return nil
	}
	settings := []string{}
	if len(ff.Deploy.OCI.FreeformTags) > 0 {
		settings = append(settings, "freeform_tags")
	}
	if len(ff.Deploy.OCI.DefinedTags) > 0 {
		settings = append(settings, "defined_tags")
	}
	sort.Strings(settings)
	return settings
}

func freeformTagsFromAnnotationValue(value interface{}) (map[string]string, error) {
	if value == nil {
		return nil, nil
	}
	result := map[string]string{}
	switch typed := value.(type) {
	case map[string]string:
		for key, v := range typed {
			result[key] = v
		}
		return result, nil
	case map[string]interface{}:
		for key, v := range typed {
			result[key] = fmt.Sprint(v)
		}
		return result, nil
	default:
		return nil, fmt.Errorf("invalid freeform tags annotation")
	}
}

func definedTagsFromAnnotationValue(value interface{}) (OCIDefinedTags, error) {
	if value == nil {
		return nil, nil
	}
	result := OCIDefinedTags{}
	switch typed := value.(type) {
	case map[string]map[string]string:
		for namespace, values := range typed {
			result[namespace] = map[string]interface{}{}
			for key, v := range values {
				result[namespace][key] = v
			}
		}
		return result, nil
	case map[string]map[string]interface{}:
		return ConvertDefinedTagsFromInterface(typed), nil
	case map[string]interface{}:
		for namespace, rawValues := range typed {
			converted := map[string]interface{}{}
			switch values := rawValues.(type) {
			case map[string]interface{}:
				for key, value := range values {
					converted[key] = value
				}
			case map[string]string:
				for key, value := range values {
					converted[key] = value
				}
			default:
				return nil, fmt.Errorf("invalid defined tags annotation")
			}
			result[namespace] = converted
		}
		return result, nil
	default:
		return nil, fmt.Errorf("invalid defined tags annotation")
	}
}

func setFreeformTagAnnotation(annotations map[string]interface{}, tags map[string]string) {
	value := map[string]interface{}{}
	for key, tagValue := range tags {
		value[key] = tagValue
	}
	annotations[AnnotationOCIResourceFreeformTags] = value
}

func setDefinedTagAnnotation(annotations map[string]interface{}, tags OCIDefinedTags) {
	if len(tags) == 0 {
		annotations[AnnotationOCIResourceDefinedTags] = map[string]map[string]interface{}{}
		return
	}
	annotations[AnnotationOCIResourceDefinedTags] = ConvertDefinedTagsToInterface(tags)
}

func parseDefinedTagKeySpec(spec string) (string, string, error) {
	parts := strings.SplitN(strings.TrimSpace(spec), ".", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
		return "", "", fmt.Errorf("invalid defined tag key %q (expected namespace.key)", spec)
	}
	return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), nil
}

func ApplyOCIResourceTagFlagsToAnnotations(
	annotations map[string]interface{},
	freeformSpecs []string,
	definedSpecs []string,
	removeFreeform []string,
	removeDefined []string,
	clearFreeform bool,
	clearDefined bool,
) (map[string]interface{}, error) {
	if annotations == nil {
		annotations = make(map[string]interface{})
	}
	freeform, err := freeformTagsFromAnnotationValue(annotations[AnnotationOCIResourceFreeformTags])
	if err != nil {
		return nil, err
	}
	defined, err := definedTagsFromAnnotationValue(annotations[AnnotationOCIResourceDefinedTags])
	if err != nil {
		return nil, err
	}
	if freeform == nil {
		freeform = map[string]string{}
	}
	if defined == nil {
		defined = OCIDefinedTags{}
	}

	if clearFreeform {
		freeform = map[string]string{}
	}
	if clearDefined {
		defined = OCIDefinedTags{}
	}

	parsedFreeform, err := ParseFreeformTagSpecs(freeformSpecs)
	if err != nil {
		return nil, err
	}
	for key, value := range parsedFreeform {
		freeform[key] = value
	}

	parsedDefined, err := ParseDefinedTagSpecs(definedSpecs)
	if err != nil {
		return nil, err
	}
	for namespace, values := range parsedDefined {
		if defined[namespace] == nil {
			defined[namespace] = map[string]interface{}{}
		}
		for key, value := range values {
			defined[namespace][key] = value
		}
	}

	for _, key := range removeFreeform {
		delete(freeform, strings.TrimSpace(key))
	}
	for _, spec := range removeDefined {
		namespace, key, err := parseDefinedTagKeySpec(spec)
		if err != nil {
			return nil, err
		}
		if defined[namespace] != nil {
			delete(defined[namespace], key)
			if len(defined[namespace]) == 0 {
				delete(defined, namespace)
			}
		}
	}

	if len(freeform) > 0 || clearFreeform || len(removeFreeform) > 0 {
		setFreeformTagAnnotation(annotations, freeform)
	} else {
		delete(annotations, AnnotationOCIResourceFreeformTags)
	}
	if len(defined) > 0 || clearDefined || len(removeDefined) > 0 {
		setDefinedTagAnnotation(annotations, defined)
	} else {
		delete(annotations, AnnotationOCIResourceDefinedTags)
	}
	return annotations, nil
}

func ApplyOCIResourceTagsToAnnotations(annotations map[string]interface{}, freeform map[string]string, defined OCIDefinedTags) map[string]interface{} {
	if annotations == nil {
		annotations = make(map[string]interface{})
	}
	if len(freeform) > 0 {
		setFreeformTagAnnotation(annotations, freeform)
	}
	if len(defined) > 0 {
		setDefinedTagAnnotation(annotations, defined)
	}
	return annotations
}
