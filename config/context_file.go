package config

import (
	"io/ioutil"

	"github.com/go-yaml/yaml"
)

// ContextFile defines the internal structure of a default context
type ContextFile struct {
	ContextProvider string `yaml:"provider" json:"provider"`
	EnvFnAPIURL     string `yaml:"api-url" json:"apiUrl"`
	EnvFnRegistry   string `yaml:"registry" json:"registry"`
}

// NewContextFile creates a new instance of the context YAML file
func NewContextFile(filePath string) (*ContextFile, error) {
	c := &ContextFile{}
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		return c, err
	}
	if err = yaml.Unmarshal(contents, c); err != nil {
		return c, err
	}
	return c, nil
}
