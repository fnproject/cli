package main

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

type inputMap struct {
	Body interface{}
}
type outputMap struct {
	Body interface{}
}

type fftest struct {
	Name   string     `yaml:"name,omitempty" json:"name,omitempty"`
	Input  *inputMap  `yaml:"input,omitempty" json:"input,omitempty"`
	Output *outputMap `yaml:"outoutput,omitempty" json:"output,omitempty"`
	Err    *string    `yaml:"err,omitempty" json:"err,omitempty"`
	// Env    map[string]string `yaml:"env,omitempty" json:"env,omitempty"`
}

type InputVar struct {
	Name     string `yaml:"name" json:"name"`
	Required bool   `yaml:"required" json:"required"`
}
type Expects struct {
	Config []InputVar `yaml:"config" json:"config"`
}

type funcfile struct {
	Name string `yaml:"name,omitempty" json:"name,omitempty"`

	// Build params
	Version    string   `yaml:"version,omitempty" json:"version,omitempty"`
	Runtime    string   `yaml:"runtime,omitempty" json:"runtime,omitempty"`
	Entrypoint string   `yaml:"entrypoint,omitempty" json:"entrypoint,omitempty"`
	Cmd        string   `yaml:"cmd,omitempty" json:"cmd,omitempty"`
	Build      []string `yaml:"build,omitempty" json:"build,omitempty"`
	Tests      []fftest `yaml:"tests,omitempty" json:"tests,omitempty"`
	BuildImage string   `yaml:"build_image,omitempty" json:"build_image,omitempty"` // Image to use as base for building
	RunImage   string   `yaml:"run_image,omitempty" json:"run_image,omitempty"`     // Image to use for running

	// Route params
	// TODO embed models.Route
	Type             string                 `yaml:"type,omitempty" json:"type,omitempty"`
	Memory           uint64                 `yaml:"memory,omitempty" json:"memory,omitempty"`
	Cpus             string                 `yaml:"cpus,omitempty" json:"cpus,omitempty"`
	Format           string                 `yaml:"format,omitempty" json:"format,omitempty"`
	Timeout          *int32                 `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Path             string                 `yaml:"path,omitempty" json:"path,omitempty"`
	Config           map[string]string      `yaml:"config,omitempty" json:"config,omitempty"`
	Headers          map[string][]string    `yaml:"headers,omitempty" json:"headers,omitempty"`
	IDLETimeout      *int32                 `yaml:"idle_timeout,omitempty" json:"idle_timeout,omitempty"`
	RouteAnnotations map[string]interface{} `yaml:"route_annotations,omitempty" json:"type:omitempty"`

	// Run/test
	Expects Expects `yaml:"expects,omitempty" json:"expects,omitempty"`
}

func (ff *funcfile) ImageName() string {
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

func (ff *funcfile) RuntimeTag() (runtime, tag string) {
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
		if exists(fullfn) {
			return fullfn, nil
		}
	}
	return "", newNotFoundError("could not find function file")
}
func findAndParseFuncfile(path string) (fpath string, ff *funcfile, err error) {
	fpath, err = findFuncfile(path)
	if err != nil {
		return "", nil, err
	}
	ff, err = parseFuncfile(fpath)
	if err != nil {
		return "", nil, err
	}
	return fpath, ff, err
}

func loadFuncfile() (string, *funcfile, error) {
	return findAndParseFuncfile(".")
}

func parseFuncfile(path string) (*funcfile, error) {
	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		return decodeFuncfileJSON(path)
	case ".yaml", ".yml":
		return decodeFuncfileYAML(path)
	}
	return nil, errUnexpectedFileFormat
}

func storeFuncfile(path string, ff *funcfile) error {
	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		return encodeFuncfileJSON(path, ff)
	case ".yaml", ".yml":
		return encodeFuncfileYAML(path, ff)
	}
	return errUnexpectedFileFormat
}

func decodeFuncfileJSON(path string) (*funcfile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open %s for parsing. Error: %v", path, err)
	}
	ff := &funcfile{}
	// ff.Route = &fnmodels.Route{}
	err = json.NewDecoder(f).Decode(ff)
	// ff := fff.MakeFuncFile()
	return ff, err
}

func decodeFuncfileYAML(path string) (*funcfile, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not open %s for parsing. Error: %v", path, err)
	}
	ff := &funcfile{}
	err = yaml.Unmarshal(b, ff)
	// ff := fff.MakeFuncFile()
	return ff, err
}

func encodeFuncfileJSON(path string, ff *funcfile) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open %s for encoding. Error: %v", path, err)
	}
	return json.NewEncoder(f).Encode(ff)
}

func encodeFuncfileYAML(path string, ff *funcfile) error {
	b, err := yaml.Marshal(ff)
	if err != nil {
		return fmt.Errorf("could not encode function file. Error: %v", err)
	}
	return ioutil.WriteFile(path, b, os.FileMode(0644))
}

func isFuncfile(path string, info os.FileInfo) bool {
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
