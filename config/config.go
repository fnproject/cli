package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/fnproject/fn_go"
	"github.com/fnproject/fn_go/provider"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

const (
	rootConfigPathName = ".fn"

	contextsPathName       = "contexts"
	configName             = "config"
	contextConfigFileName  = "config.yaml"
	defaultContextFileName = "default.yaml"
	defaultLocalAPIURL     = "http://localhost:8080"
	DefaultProvider        = "default"

	ReadWritePerms = os.FileMode(0755)

	CurrentContext    = "current-context"
	ContextProvider   = "provider"
	CurrentCliVersion = "cli-version"

	EnvFnRegistry = "registry"
	EnvFnContext  = "context"

	OCI_CLI_AUTH_ENV_VAR            = "OCI_CLI_AUTH"
	OCI_CLI_AUTH_INSTANCE_PRINCIPAL = "instance_principal"
	OCI_CLI_AUTH_INSTANCE_OBO_USER  = "instance_obo_user"
)

var defaultRootConfigContents = &ContextMap{CurrentContext: "default", CurrentCliVersion: Version}

func DefaultContextConfigContents() (contextMap *ContextMap) {
	ociCliAuth := os.Getenv(OCI_CLI_AUTH_ENV_VAR)

	if ociCliAuth == OCI_CLI_AUTH_INSTANCE_OBO_USER {
		contextMap = &ContextMap{
			ContextProvider: fn_go.OracleCSProvider,
			EnvFnRegistry:   "",
		}
	} else if ociCliAuth == OCI_CLI_AUTH_INSTANCE_OBO_USER {
		contextMap = &ContextMap{
			ContextProvider: fn_go.OracleIPProvider,
			EnvFnRegistry:   "",
		}
	} else {
		contextMap = &ContextMap{
			ContextProvider:      fn_go.DefaultProvider,
			provider.CfgFnAPIURL: defaultLocalAPIURL,
			EnvFnRegistry:        "",
		}
		viper.SetDefault(provider.CfgFnAPIURL, defaultLocalAPIURL)
	}

	return contextMap
}

type ContextMap map[string]string

// Init : Initialise/load config direc
func Init() error {
	viper.AutomaticEnv() // read in environment variables that match
	viper.SetEnvPrefix("fn")

	replacer := strings.NewReplacer("-", "_")
	viper.SetEnvKeyReplacer(replacer)

	return ensureConfiguration()
}

// EnsureConfiguration ensures context configuration directory hierarchy is in place, if not
// creates it and the default context configuration files
func ensureConfiguration() error {
	home := GetHomeDir()

	rootConfigPath := filepath.Join(home, rootConfigPathName)
	if _, err := os.Stat(rootConfigPath); os.IsNotExist(err) {
		if err = os.Mkdir(rootConfigPath, ReadWritePerms); err != nil {
			return fmt.Errorf("error creating .fn directory %v", err)
		}
	}

	contextConfigFilePath := filepath.Join(rootConfigPath, contextConfigFileName)
	if _, err := os.Stat(contextConfigFilePath); os.IsNotExist(err) {
		file, err := os.Create(contextConfigFilePath)
		if err != nil {
			return fmt.Errorf("error creating config.yaml file %v", err)
		}

		err = WriteYamlFile(file.Name(), defaultRootConfigContents)
		if err != nil {
			return err
		}
	}
	contextsPath := filepath.Join(rootConfigPath, contextsPathName)
	if _, err := os.Stat(contextsPath); os.IsNotExist(err) {
		if err = os.Mkdir(contextsPath, ReadWritePerms); err != nil {
			return fmt.Errorf("error creating contexts directory %v", err)
		}
	}

	defaultContextPath := filepath.Join(contextsPath, defaultContextFileName)
	if _, err := os.Stat(defaultContextPath); os.IsNotExist(err) {
		_, err = os.Create(defaultContextPath)
		if err != nil {
			return fmt.Errorf("error creating default.yaml context file %v", err)
		}

		err = WriteYamlFile(defaultContextPath, DefaultContextConfigContents())
		if err != nil {
			return err
		}
	}

	return nil
}

// GetContextsPath : Returns the path to the contexts directory.
func GetContextsPath() string {
	contextsPath := filepath.Join(rootConfigPathName, contextsPathName)
	return contextsPath
}

func LoadConfiguration(c *cli.Context) error {
	// Find home directory.
	home := GetHomeDir()
	context := ""

	viper.AddConfigPath(filepath.Join(home, rootConfigPathName))
	viper.SetConfigName(configName)

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	if context = c.String(EnvFnContext); context == "" {
		context = viper.GetString(CurrentContext)
	}

	viper.AddConfigPath(filepath.Join(home, rootConfigPathName, contextsPathName))
	viper.SetConfigName(context)

	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("%v \n", err)
		err := WriteConfigValueToConfigFile(CurrentContext, "default")
		if err != nil {
			return err
		}
		err = WriteConfigValueToConfigFile(CurrentCliVersion, Version)
		if err != nil {
			return err
		}

		fmt.Println("current context has been set to default")
		return nil
	}

	viper.Set(CurrentContext, context)
	return nil
}

func WriteConfigValueToConfigFile(key, value string) error {
	home := GetHomeDir()

	configFilePath := filepath.Join(home, rootConfigPathName, contextConfigFileName)
	f, err := os.OpenFile(configFilePath, os.O_RDWR, ReadWritePerms)
	if err != nil {
		return err
	}
	defer f.Close()

	file, err := DecodeYAMLFile(f.Name())
	if err != nil {
		return err
	}

	configValues := ContextMap{}
	for k, v := range *file {
		if k == key {
			configValues[k] = value
		} else {
			configValues[k] = v
		}
	}
	configValues[key] = value

	err = atomicwrite(f.Name(), &configValues)
	if err != nil {
		return err
	}

	return nil
}

func atomicwrite(file string, c *ContextMap) (err error) {
	// create a temp file
	path, filename := filepath.Split(file)

	f, err := ioutil.TempFile(path, filename) //tempfile name is generated by adding a random string to the end of given filename
	if err != nil {
		return fmt.Errorf("cannot create temp file: %v", err)
	}

	defer f.Close()
	defer os.Remove(f.Name())

	err = WriteYamlFile(f.Name(), c)
	if err != nil {
		return err
	}

	info, err := os.Stat(file)
	if err != nil {
		return err
	} else {
		_ = os.Chmod(f.Name(), info.Mode())
	}

	// replace file with the tempfile
	err = os.Rename(f.Name(), file)
	if err != nil {
		return fmt.Errorf("error replacing file with tempfile")
	}
	return nil
}
