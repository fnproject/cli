package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
	yaml "gopkg.in/yaml.v2"
)

const (
	rootConfigPathName     = ".fn"
	contextsPathName       = "contexts"
	configName             = "config"
	contextConfigFileName  = "config.yaml"
	defaultContextFileName = "default.yaml"

	readWritePerms = os.FileMode(0755)

	CurrentContext  = "current-context"
	ContextProvider = "provider"

	EnvFnRegistry = "registry"
	EnvFnToken    = "token"
	EnvFnAPIURL   = "api_url"
	EnvFnContext  = "context"
)

var defaultRootConfigContents = map[string]string{CurrentContext: "default"}
var defaultContextConfigContents = map[string]string{
	ContextProvider: "default",
	EnvFnAPIURL:     "https://localhost:8080",
	EnvFnRegistry:   "",
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

	contextsPath := filepath.Join(rootConfigPath, contextsPathName)
	if _, err = os.Stat(contextsPath); os.IsNotExist(err) {
		if err = os.Mkdir(contextsPath, readWritePerms); err != nil {
			panic(err)
		}
	}

	defaultContextPath := filepath.Join(contextsPath, defaultContextFileName)
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

func LoadConfiguration(c *cli.Context) error {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	context := ""

	if context = c.String(EnvFnContext); context == "" {
		viper.AddConfigPath(filepath.Join(home, rootConfigPathName))
		viper.SetConfigName(configName)

		readConfig()

		context = viper.GetString(CurrentContext)
		if context == "" {
			fmt.Println("Config file does not contain context")
			os.Exit(1)
		}
	}

	viper.AddConfigPath(filepath.Join(home, rootConfigPathName, contextsPathName))
	viper.SetConfigName(context)
	readConfig()

	return nil
}

func readConfig() {
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
