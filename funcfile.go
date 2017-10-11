package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

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

type Funcfile struct {
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
	Type        string              `yaml:"type,omitempty" json:"type,omitempty"`
	Memory      uint64              `yaml:"memory,omitempty" json:"memory,omitempty"`
	Format      string              `yaml:"format,omitempty" json:"format,omitempty"`
	Timeout     *int32              `yaml:"timeout,omitempty" json:"timeout,omitempty"`
	Path        string              `yaml:"path,omitempty" json:"path,omitempty"`
	Config      map[string]string   `yaml:"config,omitempty" json:"config,omitempty"`
	Headers     map[string][]string `yaml:"headers,omitempty" json:"headers,omitempty"`
	IDLETimeout *int32              `yaml:"idle_timeout,omitempty" json:"idle_timeout,omitempty"`

	// Run/test
	Expects Expects `yaml:"expects,omitempty" json:"expects,omitempty"`
}

func (ff *Funcfile) ImageName() string {
	fname := ff.Name
	if !strings.Contains(fname, "/") {
		// then we'll prefix FN_REGISTRY
		reg := os.Getenv(envFnRegistry)
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

func (ff *Funcfile) RuntimeTag() (runtime, tag string) {
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
func findAndParseFuncfile(path string) (fpath string, ff *Funcfile, err error) {
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

func loadFuncfile() (string, *Funcfile, error) {
	return findAndParseFuncfile(".")
}

func parseFuncfile(path string) (*Funcfile, error) {
	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		return decodeFuncfileJSON(path)
	case ".yaml", ".yml":
		return decodeFuncfileYAML(path)
	}
	return nil, errUnexpectedFileFormat
}

func storeFuncfile(path string, ff *Funcfile) error {
	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		return encodeFuncfileJSON(path, ff)
	case ".yaml", ".yml":
		return encodeFuncfileYAML(path, ff)
	}
	return errUnexpectedFileFormat
}

func decodeFuncfileJSON(path string) (*Funcfile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("could not open %s for parsing. Error: %v", path, err)
	}
	ff := &Funcfile{}
	// ff.Route = &fnmodels.Route{}
	err = json.NewDecoder(f).Decode(ff)
	// ff := fff.MakeFuncFile()
	return ff, err
}

func decodeFuncfileYAML(path string) (*Funcfile, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not open %s for parsing. Error: %v", path, err)
	}
	ff := &Funcfile{}
	err = yaml.Unmarshal(b, ff)
	// ff := fff.MakeFuncFile()
	return ff, err
}

func encodeFuncfileJSON(path string, ff *Funcfile) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("could not open %s for encoding. Error: %v", path, err)
	}
	return json.NewEncoder(f).Encode(ff)
}

func encodeFuncfileYAML(path string, ff *Funcfile) error {
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
