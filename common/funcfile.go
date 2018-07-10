package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/fnproject/cli/config"
	"github.com/spf13/viper"

	yaml "gopkg.in/yaml.v2"
)

var (
	validFuncfileNames = [...]string{
		"func.yaml",
		"func.yml",
		"func.json",
	}
)

// InputMap to be used within FFTest.
type InputMap struct {
	Body interface{}
}

// OutputMap to be used within FFTest.
type OutputMap struct {
	Body interface{}
}

// FFTest represents a test for a funcfile.
type FFTest struct {
	Name   string     `yaml:"name,omitempty" json:"name,omitempty"`
	Input  *InputMap  `yaml:"input,omitempty" json:"input,omitempty"`
	Output *OutputMap `yaml:"outoutput,omitempty" json:"output,omitempty"`
	Err    *string    `yaml:"err,omitempty" json:"err,omitempty"`
	// Env    map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
}

type inputVar struct {
	Name     string `yaml:"name" json:"name"`
	Required bool   `yaml:"required" json:"required"`
}

// Expects represents expected env vars in funcfile.
type Expects struct {
	Config []inputVar `yaml:"config" json:"config"`
}

// FuncFile defines the internal structure of a func.yaml/json/yml
type FuncFile struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`

	// Build params
	Version     string   `yaml:"version,omitempty" json:"version,omitempty"`
	Runtime     string   `yaml:"runtime,omitempty" json:"runtime,omitempty"`
	Entrypoint  string   `yaml:"entrypoint,omitempty" json:"entrypoint,omitempty"`
	Cmd         string   `yaml:"cmd,omitempty" json:"cmd,omitempty"`
	Build       []string `yaml:"build,omitempty" json:"build,omitempty"`
	Tests       []FFTest `yaml:"tests,omitempty" json:"tests,omitempty"`
	BuildImage  string   `yaml:"build_image,omitempty" json:"build_image,omitempty"` // Image to use as base for building
	RunImage    string   `yaml:"run_image,omitempty" json:"run_image,omitempty"`     // Image to use for running
	ContentType string   `yaml:"content_type,omitempty" json:"content_type,omitempty"`

	// Route params
	// TODO embed models.Route
	Type        string                 `yaml:"type,omitempty" json:"type,omitempty"`
	Memory      uint64                 `yaml:"memory,omitempty" json:"memory,omitempty"`
	Cpus        string                 `yaml:"cpus,omitempty" json:"cpus,omitempty"`
	Format      string                 `yaml:"format,omitempty" json:"format,omitempty"`
	Timeout     *int32                 `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Path        string                 `yaml:"path,omitempty" json:"path,omitempty"`
	Config      map[string]string      `yaml:"config,omitempty" json:"config,omitempty"`
	Headers     map[string][]string    `yaml:"headers,omitempty" json:"headers,omitempty"`
	IDLETimeout *int32                 `yaml:"idle_timeout,omitempty" json:"idle_timeout,omitempty"`
	Annotations map[string]interface{} `yaml:"annotations,omitempty" json:"annotations,omitempty"`

	// Run/test
	Expects Expects `yaml:"expects,omitempty" json:"expects,omitempty"`
}

// ImageName returns the name of a funcfile image
func (ff *FuncFile) ImageName() string {
	fname := ff.Name
	if !strings.Contains(fname, "/") {

		reg := viper.GetString(config.EnvFnRegistry)
		if reg != "" {
			if reg[len(reg)-1] != '/' {
				reg += "/"
			}
			fname = fmt.Sprintf("%s%s", reg, fname)
		}
	}
	if ff.Version != "" {
		fname = fmt.Sprintf("%s:%s", fname, ff.Version)
	}
	return fname
}

// RuntimeTag returns the runtime and tag.
func (ff *FuncFile) RuntimeTag() (runtime, tag string) {
	if ff.Runtime == "" {
		return "", ""
	}

	rt := ff.Runtime
	tagpos := strings.Index(rt, ":")
	if tagpos == -1 {
		return rt, ""
	}

	return rt[:tagpos], rt[tagpos+1:]
}

// findFuncfile for a func.yaml/json/yml file in path
func findFuncfile(path string) (string, error) {
	for _, fn := range validFuncfileNames {
		fullfn := filepath.Join(path, fn)
		if Exists(fullfn) {
			return fullfn, nil
		}
	}
	return "", NewNotFoundError("could not find function file")
}

// FindAndParseFuncfile for a func.yaml/json/yml file.
func FindAndParseFuncfile(path string) (fpath string, ff *FuncFile, err error) {
	fpath, err = findFuncfile(path)
	if err != nil {
		return "", nil, err
	}
	ff, err = ParseFuncfile(fpath)
	if err != nil {
		return "", nil, err
	}
	return fpath, ff, err
}

// LoadFuncfile returns a parsed funcfile.
func LoadFuncfile(path string) (string, *FuncFile, error) {
	return FindAndParseFuncfile(path)
}

// ParseFuncfile check file type to decode and parse.
func ParseFuncfile(path string) (*FuncFile, error) {
	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		return decodeFuncfileJSON(path)
	case ".yaml", ".yml":
		return decodeFuncfileYAML(path)
	}
	return nil, errUnexpectedFileFormat
}

func storeFuncfile(path string, ff *FuncFile) error {
	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		return encodeFuncfileJSON(path, ff)
	case ".yaml", ".yml":
		return EncodeFuncfileYAML(path, ff)
	}
	return errUnexpectedFileFormat
}

func decodeFuncfileJSON(path string) (*FuncFile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open %s for parsing. Error: %v", path, err)
	}
	ff := &FuncFile{}
	// ff.Route = &fnmodels.Route{}
	err = json.NewDecoder(f).Decode(ff)
	// ff := fff.MakeFuncFile()
	return ff, err
}

func decodeFuncfileYAML(path string) (*FuncFile, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not open %s for parsing. Error: %v", path, err)
	}
	ff := &FuncFile{}
	err = yaml.Unmarshal(b, ff)
	// ff := fff.MakeFuncFile()
	return ff, err
}

func encodeFuncfileJSON(path string, ff *FuncFile) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open %s for encoding. Error: %v", path, err)
	}
	return json.NewEncoder(f).Encode(ff)
}

// EncodeFuncfileYAML encodes function file.
func EncodeFuncfileYAML(path string, ff *FuncFile) error {
	b, err := yaml.Marshal(ff)
	if err != nil {
		return fmt.Errorf("could not encode function file. Error: %v", err)
	}
	return ioutil.WriteFile(path, b, os.FileMode(0644))
}

// IsFuncFile check vaid funcfile.
func IsFuncFile(path string, info os.FileInfo) bool {
	if info.IsDir() {
		return false
	}

	basefn := filepath.Base(path)
	for _, fn := range validFuncfileNames {
		if basefn == fn {
			return true
		}
	}
	return false
}
