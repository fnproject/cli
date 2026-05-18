package common

import "testing"

func TestApplyOCIResourceTagFlagsToAnnotations(t *testing.T) {
	annotations, err := ApplyOCIResourceTagFlagsToAnnotations(nil, []string{"Department=Finance"}, []string{"Operations.CostCenter=42"}, nil, nil, false, false)
	if err != nil {
		t.Fatalf("ApplyOCIResourceTagFlagsToAnnotations() error = %v", err)
	}
	freeform, err := freeformTagsFromAnnotationValue(annotations[AnnotationOCIResourceFreeformTags])
	if err != nil {
		t.Fatalf("freeformTagsFromAnnotationValue() error = %v", err)
	}
	if freeform["Department"] != "Finance" {
		t.Fatalf("expected Department=Finance, got %#v", freeform)
	}
	defined, err := definedTagsFromAnnotationValue(annotations[AnnotationOCIResourceDefinedTags])
	if err != nil {
		t.Fatalf("definedTagsFromAnnotationValue() error = %v", err)
	}
	if defined["Operations"]["CostCenter"] != "42" {
		t.Fatalf("expected Operations.CostCenter=42, got %#v", defined)
	}

	annotations, err = ApplyOCIResourceTagFlagsToAnnotations(annotations, nil, nil, []string{"Department"}, []string{"Operations.CostCenter"}, false, false)
	if err != nil {
		t.Fatalf("ApplyOCIResourceTagFlagsToAnnotations(remove) error = %v", err)
	}
	freeform, _ = freeformTagsFromAnnotationValue(annotations[AnnotationOCIResourceFreeformTags])
	if len(freeform) != 0 {
		t.Fatalf("expected freeform tags to be empty, got %#v", freeform)
	}
	defined, _ = definedTagsFromAnnotationValue(annotations[AnnotationOCIResourceDefinedTags])
	if len(defined) != 0 {
		t.Fatalf("expected defined tags to be empty, got %#v", defined)
	}
}

func TestApplyOCIResourceTagFlagsToAnnotationsKeepsEmptyMapsForRemoval(t *testing.T) {
	annotations, err := ApplyOCIResourceTagFlagsToAnnotations(nil, []string{"Department=Finance"}, []string{"Operations.CostCenter=42"}, nil, nil, false, false)
	if err != nil {
		t.Fatalf("ApplyOCIResourceTagFlagsToAnnotations() error = %v", err)
	}
	annotations, err = ApplyOCIResourceTagFlagsToAnnotations(annotations, nil, nil, []string{"Department"}, []string{"Operations.CostCenter"}, false, false)
	if err != nil {
		t.Fatalf("ApplyOCIResourceTagFlagsToAnnotations(remove) error = %v", err)
	}
	if _, ok := annotations[AnnotationOCIResourceFreeformTags]; !ok {
		t.Fatal("expected freeform tags annotation to remain present as empty map for removal semantics")
	}
	if _, ok := annotations[AnnotationOCIResourceDefinedTags]; !ok {
		t.Fatal("expected defined tags annotation to remain present as empty map for removal semantics")
	}
}

func TestSetOCIResourceTagsOnFuncFile(t *testing.T) {
	ff := &FuncFileV20180708{}
	freeform := map[string]string{"Department": "Finance"}
	defined := OCIDefinedTags{"Operations": {"CostCenter": "42"}}

	SetOCIResourceTagsOnFuncFile(ff, freeform, defined)

	if ff.Deploy == nil || ff.Deploy.OCI == nil {
		t.Fatal("expected deploy.oci to be initialized")
	}
	if ff.Deploy.OCI.FreeformTags["Department"] != "Finance" {
		t.Fatalf("expected Department freeform tag, got %#v", ff.Deploy.OCI.FreeformTags)
	}
	if ff.Deploy.OCI.DefinedTags["Operations"]["CostCenter"] != "42" {
		t.Fatalf("expected Operations.CostCenter defined tag, got %#v", ff.Deploy.OCI.DefinedTags)
	}
}

func TestParseDefinedTagValuePreservesPlainScalarsAsStrings(t *testing.T) {
	if got := parseDefinedTagValue("10"); got != "10" {
		t.Fatalf("expected plain numeric scalar to remain string, got %#v", got)
	}
	if got := parseDefinedTagValue("true"); got != "true" {
		t.Fatalf("expected plain boolean scalar to remain string, got %#v", got)
	}
}

func TestParseDefinedTagValueStillAllowsExplicitJSON(t *testing.T) {
	if got := parseDefinedTagValue(`"10"`); got != "10" {
		t.Fatalf("expected quoted JSON string to decode to string, got %#v", got)
	}
	obj, ok := parseDefinedTagValue(`{"level":2}`).(map[string]interface{})
	if !ok || obj["level"] != float64(2) {
		t.Fatalf("expected JSON object to remain supported, got %#v", obj)
	}
}

func TestApplyOCIResourceTagFlagsToAnnotationsClearDefinedTagsUsesEmptyMap(t *testing.T) {
	annotations, err := ApplyOCIResourceTagFlagsToAnnotations(nil, nil, []string{"dry_run_tag.example-tag=10"}, nil, nil, false, false)
	if err != nil {
		t.Fatalf("ApplyOCIResourceTagFlagsToAnnotations(seed) error = %v", err)
	}
	annotations, err = ApplyOCIResourceTagFlagsToAnnotations(annotations, nil, nil, nil, nil, false, true)
	if err != nil {
		t.Fatalf("ApplyOCIResourceTagFlagsToAnnotations(clear-defined) error = %v", err)
	}
	raw, ok := annotations[AnnotationOCIResourceDefinedTags]
	if !ok {
		t.Fatal("expected defined tags annotation to be present")
	}
	definedRaw, ok := raw.(map[string]map[string]interface{})
	if !ok {
		t.Fatalf("expected defined tags annotation to be an empty map payload, got %#v", raw)
	}
	if len(definedRaw) != 0 {
		t.Fatalf("expected empty defined tags payload, got %#v", definedRaw)
	}
}
