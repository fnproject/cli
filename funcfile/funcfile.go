package funcfile

import (
	"fmt"
	"os"
	"strings"
)

type InputMap struct {
	Body interface{}
}
type OutputMap struct {
	Body interface{}
}

type Fftest struct {
	Name   string     `yaml:"name,omitempty" json:"name,omitempty"`
	Input  *InputMap  `yaml:"input,omitempty" json:"input,omitempty"`
	Output *OutputMap `yaml:"outoutput,omitempty" json:"output,omitempty"`
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
	Tests      []Fftest `yaml:"tests,omitempty" json:"tests,omitempty"`
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

const (
	envFnRegistry = "FN_REGISTRY"
)

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
