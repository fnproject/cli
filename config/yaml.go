package config

import (
	"io/ioutil"

	"github.com/go-yaml/yaml"
)

func DecodeYAMLFile(filename string) (*ContextMap, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	yf := &ContextMap{}
	err = yaml.Unmarshal(b, yf)
	return yf, err
}

func WriteYamlFile(filename string, values *ContextMap) error {
	b, err := yaml.Marshal(values)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filename, b, ReadWritePerms)
}
