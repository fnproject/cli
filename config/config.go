package config

import (
	"fmt"
	"io/ioutil"
	"log"
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
func EnsureConfiguration() error {
	home := GetHomeDir()

	rootConfigPath := filepath.Join(home, RootConfigPathName)
	if _, err := os.Stat(rootConfigPath); os.IsNotExist(err) {
		if err = os.Mkdir(rootConfigPath, readWritePerms); err != nil {
			return fmt.Errorf("error creating .fn directory %v", err)

		}
	}

	contextConfigFilePath := filepath.Join(rootConfigPath, ContextConfigFileName)
	if _, err := os.Stat(contextConfigFilePath); os.IsNotExist(err) {
		_, err = os.Create(contextConfigFilePath)
		if err != nil {
			return fmt.Errorf("error creating config.yaml file %v", err)
		}

		err = WriteYamlFile(contextConfigFilePath, DefaultRootConfigContents)
		if err != nil {
			return err
		}
	}

	contextsPath := filepath.Join(rootConfigPath, ContextsPathName)
	if _, err := os.Stat(contextsPath); os.IsNotExist(err) {
		if err = os.Mkdir(contextsPath, readWritePerms); err != nil {
			return fmt.Errorf("error creating contexts directory %v", err)
		}
	}

	defaultContextPath := filepath.Join(contextsPath, DefaultContextFileName)
	if _, err := os.Stat(defaultContextPath); os.IsNotExist(err) {
		_, err = os.Create(defaultContextPath)
		if err != nil {
			return fmt.Errorf("error creating default.yaml context file %v", err)
		}

		err = WriteYamlFile(defaultContextPath, DefaultContextConfigContents)
		if err != nil {
			return err
		}
	}

	return nil
}

func LoadConfiguration(c *cli.Context) error {
	// Find home directory.
	home := GetHomeDir()
	context := ""
	if context = c.String(EnvFnContext); context == "" {
		viper.AddConfigPath(filepath.Join(home, RootConfigPathName))
		viper.SetConfigName(ConfigName)

		if err := viper.ReadInConfig(); err != nil {
			return err
		}
		context = viper.GetString(CurrentContext)
	}

	viper.AddConfigPath(filepath.Join(home, RootConfigPathName, ContextsPathName))
	viper.SetConfigName(context)

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("%v \n", err)
		err := WriteCurrentContextToConfigFile("default")
		if err != nil {
			return err
		}
		fmt.Println("current context has been set to default")
		return nil
	}

	viper.Set(CurrentContext, context)
	return nil
}

// WriteYamlFile writes to the yaml file
func WriteYamlFile(filename string, value map[string]string) error {
	marshaled, err := yaml.Marshal(value)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filename, marshaled, readWritePerms)
	if err != nil {
		return err
	}

	return nil
}

func WriteCurrentContextToConfigFile(value string) error {
	home := GetHomeDir()

	configFilePath := filepath.Join(home, RootConfigPathName, ContextConfigFileName)
	file, err := decodeYAMLFile(configFilePath)
	if err != nil {
		return err
	}

	configValues := map[string]string{}

	for k, v := range file {
		if k == CurrentContext {
			configValues[k] = value
		} else {
			configValues[k] = v
		}
	}

	err = WriteYamlFile(configFilePath, configValues)
	if err != nil {
		return err
	}

	return nil
}

func decodeYAMLFile(path string) (map[string]string, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not open %s for parsing. Error: %v", path, err)
	}

	yf := map[string]string{}
	err = yaml.Unmarshal(b, yf)
	if err != nil {
		return nil, err
	}
	return yf, err
}

func GetHomeDir() string {
	home, err := homedir.Dir()
	if err != nil {
		log.Fatalln("could not get home directory:", err)
	}

	return home
}
