/*
 * Copyright (c) 2019, 2020 Oracle and/or its affiliates. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package common

import (
	"io/ioutil"
	"os"
	"path"
	"reflect"
	"testing"
)

func TestMergeFuncFileInitYAML(t *testing.T) {

	ff := FuncFileV20180708{
		Schema_version: 0,
		Name:           "old",
		Version:        "old",
		Runtime:        "old",
		Build_image:    "old",
		Run_image:      "old",
		Cmd:            "old",
		Entrypoint:     "old",
		Content_type:   "old",
		Type:           "old",
		Memory:         0,
		Timeout:        nil,
		IDLE_timeout:   nil,
		Config:         nil,
		Annotations:    nil,
		Build:          nil,
		Expects:        Expects{},
		Triggers:       nil,
	}

	tests := []struct {
		name     string
		initYAML string
		wantErr  bool
		wantFF   FuncFileV20180708
	}{
		{
			name:     "invalid init yaml",
			initYAML: "foobaryaml",
			wantErr:  true,
			wantFF:   ff,
		},
		{
			name: "valid init file replaces old func file",
			initYAML: `
schema_version: 20180708
version: 0.0.1
runtime: go
entrypoint: ./func
`,
			wantErr: false,
			wantFF: FuncFileV20180708{
				Schema_version: 0,
				Name:           "old",
				Version:        "old",
				Runtime:        "go",
				Build_image:    "",
				Run_image:      "",
				Cmd:            "",
				Entrypoint:     "./func",
				Content_type:   "",
				Type:           "old",
				Memory:         0,
				Timeout:        nil,
				IDLE_timeout:   nil,
				Config:         nil,
				Annotations:    nil,
				Build:          nil,
				Expects:        Expects{},
				Triggers:       nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			folder, filePath := createInitYAML(tt.initYAML)
			defer os.RemoveAll(folder)
			if err := MergeFuncFileInitYAML(filePath, &ff); (err != nil) != tt.wantErr {
				t.Errorf("MergeFuncFileInitYAML() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(ff, tt.wantFF) {
				t.Errorf("MergeFuncFileInitYAML() did not merge func file correctly, got = %v, want %v", ff, tt.wantFF)
			}
		})
	}
}

func TestMergeFuncFileInitYAMLCopiesDeploySection(t *testing.T) {
	ff := FuncFileV20180708{Name: "hello"}
	initYAML := `
schema_version: 20180708
runtime: go
deploy:
  oci:
    provisioned_concurrency:
      strategy: CONSTANT
      count: 3
    detached_mode:
      timeout: 20m
      on_success:
        type: stream
        ocid: ocid1.stream.oc1..example
`
	folder, filePath := createInitYAML(initYAML)
	defer os.RemoveAll(folder)

	if err := MergeFuncFileInitYAML(filePath, &ff); err != nil {
		t.Fatalf("MergeFuncFileInitYAML() error = %v", err)
	}
	if ff.Deploy == nil || ff.Deploy.OCI == nil || ff.Deploy.OCI.ProvisionedConcurrency == nil {
		t.Fatalf("expected deploy.oci.provisioned_concurrency to be copied from init yaml")
	}
	if ff.Deploy.OCI.ProvisionedConcurrency.Strategy != "CONSTANT" {
		t.Fatalf("expected provisioned concurrency strategy CONSTANT, got %q", ff.Deploy.OCI.ProvisionedConcurrency.Strategy)
	}
	if ff.Deploy.OCI.ProvisionedConcurrency.Count == nil || *ff.Deploy.OCI.ProvisionedConcurrency.Count != 3 {
		t.Fatalf("expected provisioned concurrency count 3, got %#v", ff.Deploy.OCI.ProvisionedConcurrency.Count)
	}
	if ff.Deploy.OCI.DetachedMode == nil || ff.Deploy.OCI.DetachedMode.Timeout != "20m" {
		t.Fatalf("expected detached mode timeout to be copied, got %#v", ff.Deploy.OCI.DetachedMode)
	}
}

func TestFuncFileV20180708OCIManagedFunctionSettingsHelpers(t *testing.T) {
	count := 5
	ff := &FuncFileV20180708{
		Deploy: &FuncDeployConfig{
			OCI: &OCIFunctionDeployConfig{
				ProvisionedConcurrency: &OCIProvisionedConcurrencyConfig{Strategy: "CONSTANT", Count: &count},
				DetachedMode:           &OCIDetachedModeConfig{Timeout: "20m"},
			},
		},
	}

	if !ff.HasOCIManagedFunctionSettings() {
		t.Fatal("expected HasOCIManagedFunctionSettings to return true")
	}
	want := []string{"provisioned_concurrency", "detached_mode"}
	got := ff.OCIManagedFunctionSettingNames()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected OCIManagedFunctionSettingNames %v, got %v", want, got)
	}
}

func createInitYAML(contents string) (string, string) {
	folder, err := ioutil.TempDir(os.TempDir(), "fn-tests")
	if err != nil {
		panic(err)
	}
	filePath := path.Join(folder, "func.init.yaml")
	f, err := os.Create(filePath)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	_, _ = f.WriteString(contents)

	return folder, filePath
}
