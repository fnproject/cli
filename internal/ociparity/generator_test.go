package ociparity

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateAppFiles(t *testing.T) {
	tmp := t.TempDir()
	specPath := filepath.Join(tmp, "spec.yaml")
	spec := `openapi: 3.0.0
components:
  schemas:
    CreateApplicationDetails:
      type: object
      properties:
        traceConfig:
          type: object
          description: app trace config
        networkSecurityGroupIds:
          type: array
          items:
            type: string
          description: NSG ids
    UpdateApplicationDetails:
      type: object
      properties:
        traceConfig:
          type: object
        networkSecurityGroupIds:
          type: array
          items:
            type: string
`
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}
	files, err := GenerateAppFiles(specPath)
	if err != nil {
		t.Fatalf("GenerateAppFiles() error = %v", err)
	}
	flags := files[filepath.ToSlash("objects/app/generated_oci_parity_flags.go")]
	if !strings.Contains(flags, "trace-config") || !strings.Contains(flags, "network-security-group-ids") {
		t.Fatalf("generated flags missing expected fields:\n%s", flags)
	}
	apply := files[filepath.ToSlash("objects/app/generated_oci_parity_apply.go")]
	if !strings.Contains(apply, "ApplyGeneratedOCIParityAppFlags") {
		t.Fatalf("generated apply helper missing:\n%s", apply)
	}
	list := files[filepath.ToSlash("objects/app/generated_oci_parity_list.go")]
	if !strings.Contains(list, "ApplyGeneratedOCIParityAppListParams") || !strings.Contains(list, "params.SortOrder") {
		t.Fatalf("generated list helper missing expected content:\n%s", list)
	}
	shim := files[filepath.ToSlash("vendor/github.com/fnproject/fn_go/provider/oracle/shim/generated_app_oci_parity.go")]
	if !strings.Contains(shim, "applyGeneratedOCIParityCreateApplicationDetails") || !strings.Contains(shim, "applyGeneratedOCIParityListApplicationsRequest") {
		t.Fatalf("generated shim helper missing:\n%s", shim)
	}
}

func TestGenerateAppFilesSwagger2Definitions(t *testing.T) {
	tmp := t.TempDir()
	specPath := filepath.Join(tmp, "spec.yaml")
	spec := `swagger: '2.0'
definitions:
  CreateApplicationDetails:
    type: object
    properties:
      traceConfig:
        type: object
      networkSecurityGroupIds:
        type: array
        items:
          type: string
  UpdateApplicationDetails:
    type: object
    properties:
      traceConfig:
        type: object
      networkSecurityGroupIds:
        type: array
        items:
          type: string
`
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}
	files, err := GenerateAppFiles(specPath)
	if err != nil {
		t.Fatalf("GenerateAppFiles() error = %v", err)
	}
	flags := files[filepath.ToSlash("objects/app/generated_oci_parity_flags.go")]
	if !strings.Contains(flags, "trace-config") || !strings.Contains(flags, "network-security-group-ids") {
		t.Fatalf("generated flags missing expected app fields:\n%s", flags)
	}
	if !strings.Contains(flags, "lifecycle-state") || !strings.Contains(flags, "sort-order") || !strings.Contains(flags, "display-name") {
		t.Fatalf("generated list flags missing expected app fields:\n%s", flags)
	}
}

func TestGenerateFnFilesSwagger2Definitions(t *testing.T) {
	tmp := t.TempDir()
	specPath := filepath.Join(tmp, "spec.yaml")
	spec := `swagger: '2.0'
definitions:
  CreateFunctionDetails:
    type: object
    properties:
      traceConfig:
        type: object
      timeoutInSeconds:
        type: integer
      detachedModeTimeoutInSeconds:
        type: integer
      successDestination:
        type: object
      failureDestination:
        type: object
  UpdateFunctionDetails:
    type: object
    properties:
      traceConfig:
        type: object
      timeoutInSeconds:
        type: integer
      detachedModeTimeoutInSeconds:
        type: integer
      successDestination:
        type: object
      failureDestination:
        type: object
`
	if err := os.WriteFile(specPath, []byte(spec), 0o644); err != nil {
		t.Fatal(err)
	}
	files, err := GenerateFnFiles(specPath)
	if err != nil {
		t.Fatalf("GenerateFnFiles() error = %v", err)
	}
	flags := files[filepath.ToSlash("objects/fn/generated_oci_parity_flags.go")]
	for _, expected := range []string{
		"trace-config",
		"timeout-in-seconds",
		"detached-mode-timeout-in-seconds",
		"success-destination",
		"failure-destination",
		"display-name",
		"sort-order",
	} {
		if !strings.Contains(flags, expected) {
			t.Fatalf("generated function flags missing %q:\n%s", expected, flags)
		}
	}
	list := files[filepath.ToSlash("objects/fn/generated_oci_parity_list.go")]
	if !strings.Contains(list, "ApplyGeneratedOCIParityFnListParams") || !strings.Contains(list, "params.SortOrder") {
		t.Fatalf("generated function list helper missing expected content:\n%s", list)
	}
	shim := files[filepath.ToSlash("vendor/github.com/fnproject/fn_go/provider/oracle/shim/generated_fn_oci_parity.go")]
	for _, expected := range []string{
		"applyGeneratedOCIParityCreateFunctionDetails",
		"applyGeneratedOCIParityListFunctionsRequest",
		"parseGeneratedSuccessDestination",
		"parseGeneratedFailureDestination",
	} {
		if !strings.Contains(shim, expected) {
			t.Fatalf("generated function shim helper missing %q:\n%s", expected, shim)
		}
	}
}
