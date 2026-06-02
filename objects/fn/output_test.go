package fn

import (
	"encoding/json"
	"testing"

	models "github.com/fnproject/fn_go/modelsv2"
	"github.com/jmoiron/jsonq"
)

func TestGetDetachedModeView(t *testing.T) {
	fn := &models.Fn{Annotations: map[string]interface{}{
		annotationDetachedTimeoutSeconds: 1200,
		annotationSuccessDestinationKind: "STREAM",
		annotationSuccessDestinationOCID: "ocid1.stream.oc1..abc",
		annotationFailureDestinationKind: "NOTIFICATIONS",
		annotationFailureDestinationOCID: "ocid1.onstopic.oc1..abc",
	}}
	view := getDetachedModeView(fn)
	if view == nil {
		t.Fatal("expected detached mode view")
	}
	if view.Timeout != "20m" {
		t.Fatalf("expected timeout 20m, got %q", view.Timeout)
	}
	if view.OnSuccess == nil || view.OnSuccess.Type != "stream" {
		t.Fatalf("expected stream onSuccess, got %#v", view.OnSuccess)
	}
	if view.OnFailure == nil || view.OnFailure.Type != "notifications" {
		t.Fatalf("expected notifications onFailure, got %#v", view.OnFailure)
	}
}

func TestBuildInspectFnMapIncludesDetachedMode(t *testing.T) {
	fn := &models.Fn{Annotations: map[string]interface{}{
		annotationDetachedTimeoutSeconds: 1200,
		annotationSuccessDestinationKind: "STREAM",
		annotationSuccessDestinationOCID: "ocid1.stream.oc1..abc",
	}}
	inspect, err := buildInspectFnMap(fn)
	if err != nil {
		t.Fatalf("buildInspectFnMap() error = %v", err)
	}
	detached, ok := inspect["detachedMode"]
	if !ok {
		t.Fatal("expected detachedMode field in inspect map")
	}
	data, err := json.Marshal(detached)
	if err != nil {
		t.Fatalf("failed to marshal detachedMode: %v", err)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to unmarshal detachedMode: %v", err)
	}
	if got["timeout"] != "20m" {
		t.Fatalf("expected timeout 20m, got %#v", got["timeout"])
	}
}

func TestBuildInspectFnMapSupportsNestedDetachedModeQuery(t *testing.T) {
	fn := &models.Fn{Annotations: map[string]interface{}{
		annotationDetachedTimeoutSeconds: 1200,
		annotationSuccessDestinationKind: "STREAM",
		annotationSuccessDestinationOCID: "ocid1.stream.oc1..abc",
	}}
	inspect, err := buildInspectFnMap(fn)
	if err != nil {
		t.Fatalf("buildInspectFnMap() error = %v", err)
	}
	jq := jsonq.NewQuery(inspect)
	value, err := jq.Interface("detachedMode", "timeout")
	if err != nil {
		t.Fatalf("expected nested query to succeed, got error %v", err)
	}
	if value != "20m" {
		t.Fatalf("expected timeout 20m, got %#v", value)
	}
}

func TestBuildInspectFnMapIncludesTraceConfig(t *testing.T) {
	fn := &models.Fn{Annotations: map[string]interface{}{
		annotationOCIParityFnTraceConfig: map[string]interface{}{
			"isEnabled": true,
		},
	}}
	inspect, err := buildInspectFnMap(fn)
	if err != nil {
		t.Fatalf("buildInspectFnMap() error = %v", err)
	}
	trace, ok := inspect["traceConfig"]
	if !ok {
		t.Fatal("expected traceConfig field in inspect map")
	}
	data, err := json.Marshal(trace)
	if err != nil {
		t.Fatalf("failed to marshal traceConfig: %v", err)
	}
	var got map[string]interface{}
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("failed to unmarshal traceConfig: %v", err)
	}
	if got["isEnabled"] != true {
		t.Fatalf("expected isEnabled=true in traceConfig, got %#v", got["isEnabled"])
	}
}
