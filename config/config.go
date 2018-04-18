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
	RootConfigPathName     = ".fn"
	ContextsPathName       = "contexts"
	ConfigName             = "config"
	ContextConfigFileName  = "config.yaml"
	DefaultContextFileName = "default.yaml"
	DefaultLocalApiUrl     = "http://localhost:8080/v1"
	DefaultProvider        = "default"

	readWritePerms = os.FileMode(0755)

	CurrentContext  = "current-context"
	ContextProvider = "provider"

	EnvFnRegistry = "registry"
	EnvFnToken    = "token"
	EnvFnAPIURL   = "api_url"
	EnvFnContext  = "context"

	OracleKeyID         = "key_id"
	OraclePrivateKey    = "private_key"
	OracleCompartmentID = "compartment_id"
	OracleDisableCerts  = "disable_certs"
)

var DefaultRootConfigContents = map[string]string{CurrentContext: ""}
var DefaultContextConfigContents = map[string]string{
	ContextProvider: DefaultProvider,
	EnvFnAPIURL:     DefaultLocalApiUrl,
	EnvFnRegistry:   "",
}

type ContextFile struct {
	ContextProvider string `yaml:"provider"`
	EnvFnAPIURL     string `yaml:"api_url"`
	EnvFnRegistry   string `yaml:"registry"`
}

// EnsureConfiguration ensures context configuration directory hierarchy is in place, if not
// creates it and the default context configuration files
func EnsureConfiguration() {
	home, err := GetHomeDir()
	if err != nil {
		fmt.Printf("%v", err)
	}

	rootConfigPath := filepath.Join(home, RootConfigPathName)
	if _, err := os.Stat(rootConfigPath); os.IsNotExist(err) {
		if err = os.Mkdir(rootConfigPath, readWritePerms); err != nil {
			fmt.Printf("error creating .fn directory %v", err)
		}
	}

	contextConfigFilePath := filepath.Join(rootConfigPath, ContextConfigFileName)
	if _, err := os.Stat(contextConfigFilePath); os.IsNotExist(err) {
		_, err = os.Create(contextConfigFilePath)
		if err != nil {
			fmt.Printf("error creating config.yaml file %v", err)
		}

		err = WriteYamlFile(contextConfigFilePath, DefaultRootConfigContents)
		if err != nil {
			fmt.Printf("%v", err)
		}
	}

	contextsPath := filepath.Join(rootConfigPath, ContextsPathName)
	if _, err := os.Stat(contextsPath); os.IsNotExist(err) {
		if err = os.Mkdir(contextsPath, readWritePerms); err != nil {
			fmt.Printf("error creating contexts directory %v", err)
		}
	}

	defaultContextPath := filepath.Join(contextsPath, DefaultContextFileName)
	if _, err := os.Stat(defaultContextPath); os.IsNotExist(err) {
		_, err = os.Create(defaultContextPath)
		if err != nil {
			fmt.Printf("error creating default.yaml context file %v", err)
		}

		err = WriteYamlFile(defaultContextPath, DefaultContextConfigContents)
		if err != nil {
			fmt.Printf("%v", err)
		}
	}
}

func LoadConfiguration(c *cli.Context) {
	// Find home directory.
	home, err := GetHomeDir()
	if err != nil {
		fmt.Printf("%v", err)
	}

	context := ""
	if context = c.String(EnvFnContext); context == "" {
		viper.AddConfigPath(filepath.Join(home, RootConfigPathName))
		viper.SetConfigName(ConfigName)

		if err := readConfig(); err != nil {
			fmt.Printf("%v: ", err)
		}

		context = viper.GetString(CurrentContext)
	}

	viper.AddConfigPath(filepath.Join(home, RootConfigPathName, ContextsPathName))
	viper.SetConfigName(context)

	if err := readConfig(); err != nil {
		fmt.Printf("%v \n", err)
		configFilePath := filepath.Join(home, RootConfigPathName, ContextConfigFileName)
		configCurrentContext := map[string]string{CurrentContext: "default"}
		err = WriteYamlFile(configFilePath, configCurrentContext)
		if err != nil {
			fmt.Printf("%v \n", err)
		}

		fmt.Println("current context has been set to default")
		os.Exit(1)
	}

	viper.Set(CurrentContext, context)
}

// WriteYamlFile writes to the yaml file
func WriteYamlFile(filename string, value map[string]string) error {
	marshaled, err := yaml.Marshal(value)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	err = ioutil.WriteFile(filename, marshaled, readWritePerms)
	if err != nil {
		return fmt.Errorf("error writing to file %v", err)
	}

	return nil
}

func readConfig() error {
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

func GetHomeDir() (string, error) {
	home, err := homedir.Dir()
	if err != nil {
		return "", fmt.Errorf("error getting home directory %v", err)
	}

	return home, nil
}
