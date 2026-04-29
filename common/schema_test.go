package common

import (
	"os"
	"path/filepath"
	"testing"
)

func TestValidateFileAgainstSchemaAcceptsOCIDeployConfig(t *testing.T) {
	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change working directory: %v", err)
	}

	jsonFile := filepath.Join(tmpDir, "temp.json")
	content := `{
		"schema_version": 20180708,
		"name": "hello",
		"version": "0.0.1",
		"runtime": "go",
		"entrypoint": "./func",
		"deploy": {
			"oci": {
				"provisioned_concurrency": {
					"strategy": "CONSTANT",
					"count": 5
				},
				"detached_mode": {
					"timeout": "20m",
					"on_success": {
						"type": "stream",
						"ocid": "ocid1.stream.oc1..example"
					}
				}
			}
		}
	}`
	if err := os.WriteFile(jsonFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp schema file: %v", err)
	}

	if err := ValidateFileAgainstSchema("temp.json", V20180708Schema); err != nil {
		t.Fatalf("ValidateFileAgainstSchema() error = %v", err)
	}
}