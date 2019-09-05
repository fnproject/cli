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
