package fn

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/fnproject/cli/common"
	models "github.com/fnproject/fn_go/modelsv2"
	defaultprovider "github.com/fnproject/fn_go/provider/defaultprovider"
)

func TestApplyProvisionedConcurrencyNoopForNilOrNonOracleProvider(t *testing.T) {
	count := 40
	cfg := &common.OCIProvisionedConcurrencyConfig{Strategy: common.ProvisionedConcurrencyStrategyConstant, Count: &count}

	if err := ApplyProvisionedConcurrency(nil, "ocid1.fn.oc1..example", cfg); err != nil {
		t.Fatalf("expected nil provider to be a no-op, got error %v", err)
	}

	nonOracle := &defaultprovider.Provider{FnApiUrl: &url.URL{Scheme: "http", Host: "localhost:8080"}}
	if err := ApplyProvisionedConcurrency(nonOracle, "ocid1.fn.oc1..example", cfg); err != nil {
		t.Fatalf("expected non-oracle provider to be a no-op, got error %v", err)
	}
}

func TestFormatProvisionedConcurrencyDisplay(t *testing.T) {
	count := 5
	fn := &models.Fn{Annotations: map[string]interface{}{
		annotationProvisionedConcurrencyStrategy: "CONSTANT",
		annotationProvisionedConcurrencyCount:    count,
	}}
	if got := formatProvisionedConcurrencyDisplay(fn); got != "constant:5" {
		t.Fatalf("expected constant:5, got %q", got)
	}

	fn = &models.Fn{Annotations: map[string]interface{}{
		annotationProvisionedConcurrencyStrategy: "NONE",
	}}
	if got := formatProvisionedConcurrencyDisplay(fn); got != "none" {
		t.Fatalf("expected none, got %q", got)
	}
}

func TestBuildInspectFnMapIncludesProvisionedConcurrency(t *testing.T) {
	count := 2
	fn := &models.Fn{
		Name: "hello",
		Annotations: map[string]interface{}{
			annotationProvisionedConcurrencyStrategy: "CONSTANT",
			annotationProvisionedConcurrencyCount:    count,
		},
	}
	inspect, err := buildInspectFnMap(fn)
	if err != nil {
		t.Fatalf("buildInspectFnMap() error = %v", err)
	}
	pc, ok := inspect["provisionedConcurrency"]
	if !ok {
		t.Fatal("expected provisionedConcurrency field in inspect map")
	}
	data, err := json.Marshal(pc)
	if err != nil {
		t.Fatalf("failed to marshal provisionedConcurrency: %v", err)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to unmarshal provisionedConcurrency: %v", err)
	}
	if got["strategy"] != "CONSTANT" {
		t.Fatalf("expected strategy CONSTANT, got %#v", got["strategy"])
	}
	if got["count"] != float64(2) {
		t.Fatalf("expected count 2, got %#v", got["count"])
	}
}

func TestSetProvisionedConcurrencyAnnotations(t *testing.T) {
	count := 40
	fn := &models.Fn{}
	cfg := &common.OCIProvisionedConcurrencyConfig{Strategy: common.ProvisionedConcurrencyStrategyConstant, Count: &count}
	if err := SetProvisionedConcurrencyAnnotations(fn, cfg); err != nil {
		t.Fatalf("SetProvisionedConcurrencyAnnotations() error = %v", err)
	}
	if fn.Annotations[annotationProvisionedConcurrencyStrategy] != common.ProvisionedConcurrencyStrategyConstant {
		t.Fatalf("expected strategy annotation %q, got %#v", common.ProvisionedConcurrencyStrategyConstant, fn.Annotations[annotationProvisionedConcurrencyStrategy])
	}
	if fn.Annotations[annotationProvisionedConcurrencyCount] != count {
		t.Fatalf("expected count annotation %d, got %#v", count, fn.Annotations[annotationProvisionedConcurrencyCount])
	}
}

func TestSetPBFSourceAnnotations(t *testing.T) {
	fn := &models.Fn{}
	if err := setPBFSourceAnnotations(fn, "ocid1.pbflisting.oc1..example"); err != nil {
		t.Fatalf("setPBFSourceAnnotations() error = %v", err)
	}
	if got := fn.Annotations[annotationSourceType]; got != "PRE_BUILT_FUNCTIONS" {
		t.Fatalf("expected PRE_BUILT_FUNCTIONS source type, got %#v", got)
	}
	if got := fn.Annotations[annotationPbfListingID]; got != "ocid1.pbflisting.oc1..example" {
		t.Fatalf("expected pbf listing ocid annotation, got %#v", got)
	}
}

func TestWithFuncFileV20180708AppliesResourceTags(t *testing.T) {
	ff := &common.FuncFileV20180708{
		Deploy: &common.FuncDeployConfig{
			OCI: &common.OCIFunctionDeployConfig{
				FreeformTags: map[string]string{"Department": "Finance"},
				DefinedTags:  common.OCIDefinedTags{"Operations": {"CostCenter": "42"}},
			},
		},
	}
	fn := &models.Fn{}
	if err := WithFuncFileV20180708(ff, fn); err != nil {
		t.Fatalf("WithFuncFileV20180708() error = %v", err)
	}
	freeformRaw, ok := fn.Annotations[common.AnnotationOCIResourceFreeformTags].(map[string]interface{})
	if !ok || freeformRaw["Department"] != "Finance" {
		t.Fatalf("expected freeform tag annotation, got %#v", fn.Annotations[common.AnnotationOCIResourceFreeformTags])
	}
	definedRaw, ok := fn.Annotations[common.AnnotationOCIResourceDefinedTags].(map[string]map[string]interface{})
	if !ok || definedRaw["Operations"]["CostCenter"] != "42" {
		t.Fatalf("expected defined tag annotation, got %#v", fn.Annotations[common.AnnotationOCIResourceDefinedTags])
	}
}

func TestCreateCommandAllowsOptionalImageForPBF(t *testing.T) {
	cmd := Create()
	if cmd.ArgsUsage != "<app-name> <function-name> [image]" {
		t.Fatalf("expected optional image args usage for PBF support, got %q", cmd.ArgsUsage)
	}
}

func TestResolvePBFMemory(t *testing.T) {
	min := int64(512)
	resolved, err := resolvePBFMemory(0, &min)
	if err != nil {
		t.Fatalf("resolvePBFMemory(auto) error = %v", err)
	}
	if resolved != 512 {
		t.Fatalf("expected auto-selected memory 512, got %d", resolved)
	}
	resolved, err = resolvePBFMemory(1024, &min)
	if err != nil {
		t.Fatalf("resolvePBFMemory(override) error = %v", err)
	}
	if resolved != 1024 {
		t.Fatalf("expected user-selected memory 1024, got %d", resolved)
	}
	if _, err := resolvePBFMemory(256, &min); err == nil {
		t.Fatal("expected an error when memory is below the PBF minimum")
	}
}
