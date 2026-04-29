package common

import (
	"bytes"
	"strings"
	"testing"
)

func TestWarnIfOCIManagedFunctionSettingsUnsupported(t *testing.T) {
	count := 2
	ff := &FuncFileV20180708{
		Name: "hello",
		Deploy: &FuncDeployConfig{
			OCI: &OCIFunctionDeployConfig{
				ProvisionedConcurrency: &OCIProvisionedConcurrencyConfig{
					Strategy: "CONSTANT",
					Count:    &count,
				},
			},
		},
	}

	var stderr bytes.Buffer
	warned := WarnIfOCIManagedFunctionSettingsUnsupported(&stderr, nil, ff.Name, ff)
	if !warned {
		t.Fatal("expected warning helper to report that a warning was emitted")
	}
	output := stderr.String()
	if !strings.Contains(output, "OCI-specific deploy settings") {
		t.Fatalf("expected warning output to mention OCI-specific deploy settings, got %q", output)
	}
	if !strings.Contains(output, "provisioned_concurrency") {
		t.Fatalf("expected warning output to include setting name, got %q", output)
	}
}