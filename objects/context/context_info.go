package context

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/fnproject/cli/config"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

// ContextInfo holds the information found in the context YAML file
type ContextInfo struct {
	Current  bool   `json:"current"`
	Name     string `json:"name"`
	Provider string `json:"provider"`
	APIURL   string `json:"apiUrl"`
	Registry string `json:"registry"`
}

// NewContextInfo creates an instance of the contextInfo
// by parsing the provided context YAML file
func NewContextInfo(f os.FileInfo) (*ContextInfo, error) {
	currentContext := viper.GetString(config.CurrentContext)

	fileName := f.Name()
	yamlFile, err := getFileBytes(fileName)
	if err != nil {
		return nil, err
	}

	v := config.ContextFile{}
	err = yaml.Unmarshal(yamlFile, &v)
	if err != nil {
		return nil, err
	}

	name := strings.Replace(f.Name(), fileExtension, "", 1)

	isCurrent := false
	if currentContext == name {
		isCurrent = true
	}

	return &ContextInfo{
		Current:  isCurrent,
		Name:     name,
		Provider: v.ContextProvider,
		APIURL:   v.EnvFnAPIURL,
		Registry: v.EnvFnRegistry,
	}, nil
}

func getFileBytes(name string) ([]byte, error) {
	home := config.GetHomeDir()
	path := filepath.Join(home, contextsPath, name)
	yamlFile, err := ioutil.ReadFile(path)
	return yamlFile, err
}
