package main

import (
	"io/ioutil"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	yaml "gopkg.in/yaml.v2"
)

const (
	rootConfigPathName     = ".fn"
	contextsPathName       = "contexts"
	configName             = "config"
	contextConfigFileName  = "config.yaml"
	defaultContextFileName = "default.yaml"

	readWritePerms = os.FileMode(0755)

	currentContext  = "current-context"
	contextProvider = "provider"

	envFnRegistry = "registry"
	envFnToken    = "token"
	envFnAPIURL   = "api_url"
	envFnContext  = "context"
)

var defaultRootConfigContents = map[string]string{currentContext: "default"}
var defaultContextConfigContents = map[string]string{
	contextProvider: "default",
	envFnAPIURL:     "https://localhost:8080",
	envFnRegistry:   "",
}

// EnsureConfiguration ensures context configuration directory hierarchy is in place, if not
// creates it and the default context configuration files
func EnsureConfiguration() {
	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}

	rootConfigPath := filepath.Join(home, rootConfigPathName)
	if _, err := os.Stat(rootConfigPath); os.IsNotExist(err) {
		if err = os.Mkdir(rootConfigPath, readWritePerms); err != nil {
			panic(err)
		}
	}

	contextConfigFilePath := filepath.Join(rootConfigPath, contextConfigFileName)
	if _, err = os.Stat(contextConfigFilePath); os.IsNotExist(err) {
		_, err = os.Create(contextConfigFilePath)
		if err != nil {
			panic(err)
		}

		err = writeYamlFile(contextConfigFilePath, defaultRootConfigContents)
		if err != nil {
			panic(err)
		}
	}

	contextsPath = filepath.Join(rootConfigPath, contextsPathName)
	if _, err = os.Stat(contextsPath); os.IsNotExist(err) {
		if err = os.Mkdir(contextsPath, readWritePerms); err != nil {
			panic(err)
		}
	}

	defaultContextPath = filepath.Join(contextsPath, defaultContextFileName)
	if _, err = os.Stat(defaultContextPath); os.IsNotExist(err) {
		_, err = os.Create(defaultContextPath)
		if err != nil {
			panic(err)
		}

		err = writeYamlFile(defaultContextPath, defaultContextConfigContents)
		if err != nil {
			panic(err)
		}
	}
}

func writeYamlFile(filename string, value map[string]string) error {
	marshaled, err := yaml.Marshal(value)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(filename, marshaled, readWritePerms)
}
