package utils

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/go-yaml/yaml"
	homedir "github.com/mitchellh/go-homedir"
)

const (
	readWritePerms = os.FileMode(0755)
)

type ContextMap map[string]string

func DecodeYAMLFile(file *os.File) (*ContextMap, error) {
	var yf ContextMap
	err := yaml.NewDecoder(file).Decode(&yf)
	return &yf, err
}

func WriteYamlFile(file *os.File, values *ContextMap) error {
	b, err := yaml.Marshal(values)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(file.Name(), b, readWritePerms)
}

func GetHomeDir() string {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatalf("could not get home directory: %s\n", err)
	}

	return home
}
