package app

import (
	"net/url"
	"testing"

	"github.com/fnproject/fn_go/modelsv2"
	defaultprovider "github.com/fnproject/fn_go/provider/defaultprovider"
	fnprovideroracle "github.com/fnproject/fn_go/provider/oracle"
)

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
