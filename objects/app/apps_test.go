package app

import (
	"reflect"
	"net/url"
	"testing"

	"github.com/fnproject/fn_go/modelsv2"
	defaultprovider "github.com/fnproject/fn_go/provider/defaultprovider"
	fnprovideroracle "github.com/fnproject/fn_go/provider/oracle"
)

func TestBuildAppInspectDataIncludesOCIParityFields(t *testing.T) {
	app := &modelsv2.App{
		Name: "parity-app",
		Annotations: map[string]interface{}{
			annotationOCIParityTraceConfig: map[string]interface{}{
				"isEnabled": true,
			},
			annotationOCIParityNetworkSecurityGroupIds: []interface{}{"ocid1.nsg.oc1..aaaa"},
			annotationOCIParityImagePolicyConfig: map[string]interface{}{
				"isPolicyEnabled": true,
			},
			annotationOCIParitySecurityAttributes: map[string]interface{}{
				"oracle-zpr": map[string]interface{}{
					"MaxEgressCount": map[string]interface{}{
						"value": "42",
						"mode":  "enforce",
					},
				},
			},
		},
	}

	inspectData, err := buildAppInspectData(app)
	if err != nil {
		t.Fatalf("buildAppInspectData() error = %v", err)
	}

	if _, ok := inspectData["traceConfig"]; !ok {
		t.Fatalf("expected traceConfig in inspect output, got %#v", inspectData)
	}
	if _, ok := inspectData["networkSecurityGroupIds"]; !ok {
		t.Fatalf("expected networkSecurityGroupIds in inspect output, got %#v", inspectData)
	}
	if _, ok := inspectData["imagePolicyConfig"]; !ok {
		t.Fatalf("expected imagePolicyConfig in inspect output, got %#v", inspectData)
	}
	if _, ok := inspectData["securityAttributes"]; !ok {
		t.Fatalf("expected securityAttributes in inspect output, got %#v", inspectData)
	}

	gotNSGs, ok := inspectData["networkSecurityGroupIds"].([]interface{})
	if !ok || len(gotNSGs) != 1 || gotNSGs[0] != "ocid1.nsg.oc1..aaaa" {
		t.Fatalf("unexpected networkSecurityGroupIds value: %#v", inspectData["networkSecurityGroupIds"])
	}

	gotTrace, ok := inspectData["traceConfig"].(map[string]interface{})
	if !ok || !reflect.DeepEqual(gotTrace, map[string]interface{}{"isEnabled": true}) {
		t.Fatalf("unexpected traceConfig value: %#v", inspectData["traceConfig"])
	}
}

func TestSetSubnetIDAnnotations(t *testing.T) {
	app := &modelsv2.App{}
	err := setSubnetIDAnnotations(app, []string{" ocid1.subnet.oc1..aaaa ", "ocid1.subnet.oc1..bbbb"})
	if err != nil {
		t.Fatalf("setSubnetIDAnnotations() error = %v", err)
	}
	got, ok := app.Annotations[annotationSubnet].([]interface{})
	if !ok {
		t.Fatalf("expected %s annotation to be []interface{}, got %#v", annotationSubnet, app.Annotations[annotationSubnet])
	}
	if len(got) != 2 || got[0] != "ocid1.subnet.oc1..aaaa" || got[1] != "ocid1.subnet.oc1..bbbb" {
		t.Fatalf("unexpected subnet annotation value: %#v", got)
	}
}

func TestSetSubnetIDAnnotationsConflictsWithExistingAnnotation(t *testing.T) {
	app := &modelsv2.App{Annotations: map[string]interface{}{annotationSubnet: []interface{}{"ocid1.subnet.oc1..existing"}}}
	err := setSubnetIDAnnotations(app, []string{"ocid1.subnet.oc1..new"})
	if err == nil {
		t.Fatal("expected conflict error when subnet annotation already exists")
	}
}

func TestValidateSubnetIDUpdateSupported(t *testing.T) {
	if err := validateSubnetIDUpdateSupported(nil, []string{"ocid1.subnet.oc1..aaaa"}); err != nil {
		t.Fatalf("expected nil provider to allow update validation, got %v", err)
	}
	nonOracle := &defaultprovider.Provider{FnApiUrl: &url.URL{Scheme: "http", Host: "localhost:8080"}}
	if err := validateSubnetIDUpdateSupported(nonOracle, []string{"ocid1.subnet.oc1..aaaa"}); err != nil {
		t.Fatalf("expected non-oracle provider to allow subnet-id update validation, got %v", err)
	}
	oracleProvider := &fnprovideroracle.OracleProvider{}
	if err := validateSubnetIDUpdateSupported(oracleProvider, []string{"ocid1.subnet.oc1..aaaa"}); err == nil {
		t.Fatal("expected oracle provider to reject subnet-id update validation")
	}
}

func TestValidateSubnetIDCreateRequired(t *testing.T) {
	oracleProvider := &fnprovideroracle.OracleProvider{}
	if err := validateSubnetIDCreateRequired(oracleProvider, &modelsv2.App{}); err == nil {
		t.Fatal("expected oracle provider to require subnet annotations on create")
	}
	app := &modelsv2.App{Annotations: map[string]interface{}{annotationSubnet: []interface{}{"ocid1.subnet.oc1..aaaa"}}}
	if err := validateSubnetIDCreateRequired(oracleProvider, app); err != nil {
		t.Fatalf("expected oracle provider create validation to pass when subnet annotation exists, got %v", err)
	}
	nonOracle := &defaultprovider.Provider{FnApiUrl: &url.URL{Scheme: "http", Host: "localhost:8080"}}
	if err := validateSubnetIDCreateRequired(nonOracle, &modelsv2.App{}); err != nil {
		t.Fatalf("expected non-oracle provider create validation to pass without subnets, got %v", err)
	}
}
