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