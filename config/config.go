/*
 * Copyright (c) 2019, 2020 Oracle and/or its affiliates. All rights reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/fnproject/fn_go"
	"github.com/fnproject/fn_go/provider"
	"github.com/fnproject/fn_go/provider/oracle"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

const (
	rootConfigPathName = ".fn"

	contextsPathName                       = "contexts"
	configName                             = "config"
	contextConfigFileName                  = "config.yaml"
	defaultContextFileName                 = "default.yaml"
	defaultOracleRegionContextFileNameTmpl = "%s.yaml"
	defaultLocalAPIURL                     = "http://localhost:8080"
	DefaultProvider                        = "default"

	ReadWritePerms = os.FileMode(0755)

	CurrentContext      = "current-context"
	ContextProvider     = "provider"
	CurrentCliVersion   = "cli-version"
	ContainerEngineType = "container-enginetype"

	EnvFnRegistry = "registry"
	EnvFnContext  = "context"

	OCI_CLI_AUTH_ENV_VAR                  = "OCI_CLI_AUTH"
	OCI_CLI_CLOUDSHELL_ENV_VAR            = "OCI_CLI_CLOUD_SHELL"
	OCI_CLOUDSHELL_OS_NAME                = "Oracle Linux Server"
	OCI_CLOUDSHELL_OL8_VERSION_IDENTIFIER = "el8"
	DEFAULT_LINUX_OS_RELEASE_FILE_PATH    = "/etc/os-release"
	OCI_CLI_AUTH_INSTANCE_PRINCIPAL       = "instance_principal"
	OCI_CLI_AUTH_INSTANCE_OBO_USER        = "instance_obo_user"
	DefaultContainerEngineType            = "docker"
)

var EnvIsOL8CloudShell bool = false

var defaultRootConfigContents = &ContextMap{CurrentContext: "default", CurrentCliVersion: Version, ContainerEngineType: DefaultContainerEngineType}

type ViperConfigSource struct {
}

func (*ViperConfigSource) GetString(key string) string {
	return viper.GetString(key)
}

func (*ViperConfigSource) GetBool(key string) bool {
	return viper.GetBool(key)
}
func (*ViperConfigSource) IsSet(key string) bool {
	return viper.IsSet(key)
}

func DefaultContextConfigContents() (contextMap *ContextMap) {
	//Read OCI_CLI_AUTH environment variable to determine what oracle provider to use
	ociCliAuth := os.Getenv(OCI_CLI_AUTH_ENV_VAR)
	//For the oracle cloudshell and instance principal providers,
	//the default api url is determined by the provider by using the configured region in the environment

	switch ociCliAuth {
	case OCI_CLI_AUTH_INSTANCE_OBO_USER:
		contextMap = &ContextMap{
			ContextProvider: fn_go.OracleCSProvider,
			EnvFnRegistry:   "",
		}
	case OCI_CLI_AUTH_INSTANCE_PRINCIPAL:
		contextMap = &ContextMap{
			ContextProvider: fn_go.OracleIPProvider,
			EnvFnRegistry:   "",
		}
	default:
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

	rootConfigContents := defaultRootConfigContents

	rootConfigPath := filepath.Join(home, rootConfigPathName)
	if _, err := os.Stat(rootConfigPath); os.IsNotExist(err) {
		if err = os.Mkdir(rootConfigPath, ReadWritePerms); err != nil {
			return fmt.Errorf("error creating .fn directory %v", err)
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

	if os.Getenv(OCI_CLI_AUTH_ENV_VAR) == OCI_CLI_AUTH_INSTANCE_OBO_USER {
		region, tenancy, err := oracle.GetOCIRegionTenancy()

		if err != nil {
			return err
		}

		defaultOracleRegionContextFileName := fmt.Sprintf(defaultOracleRegionContextFileNameTmpl, region)
		defaultOracleRegionContextPath := filepath.Join(contextsPath, defaultOracleRegionContextFileName)
		rootConfigContents = &ContextMap{CurrentContext: region, CurrentCliVersion: Version, ContainerEngineType: DefaultContainerEngineType}
		//check for supported container engine

		if _, err := os.Stat(defaultOracleRegionContextPath); os.IsNotExist(err) {
			_, err = os.Create(defaultOracleRegionContextPath)

			if err != nil {
				return fmt.Errorf("error creating %s context file %v", defaultOracleRegionContextFileName, err)
			}

			domain, err := oracle.GetRealmDomain()
			if err != nil {
				return fmt.Errorf("error creating %s context file %v", defaultOracleRegionContextFileName, err)
			}

			apiUrl := fmt.Sprintf(oracle.FunctionsAPIURLTmpl, region, domain)
			contextMap := &ContextMap{
				ContextProvider:         fn_go.OracleCSProvider,
				provider.CfgFnAPIURL:    apiUrl,
				EnvFnRegistry:           "",
				oracle.CfgCompartmentID: tenancy,
			}

			err = WriteYamlFile(defaultOracleRegionContextPath, contextMap)

			if err != nil {
				return err
			}
		}
	}

	contextConfigFilePath := filepath.Join(rootConfigPath, contextConfigFileName)
	if _, err := os.Stat(contextConfigFilePath); os.IsNotExist(err) {
		file, err := os.Create(contextConfigFilePath)
		if err != nil {
			return fmt.Errorf("error creating config.yaml file %v", err)
		}

		err = WriteYamlFile(file.Name(), rootConfigContents)
		if err != nil {
			return err
		}
	}

	// This check is specific for running fn-cli on Oracle Cloud Shell based on OL8
	if os.Getenv(OCI_CLI_CLOUDSHELL_ENV_VAR) == "True" {
		osrelease, err := Parse(DEFAULT_LINUX_OS_RELEASE_FILE_PATH)
		if err != nil {
			return fmt.Errorf("failed to parse /etc/os-relese file: %v", err)
		}

		if osrelease.Name == OCI_CLOUDSHELL_OS_NAME {
			if strings.Contains(osrelease.PlatformID, OCI_CLOUDSHELL_OL8_VERSION_IDENTIFIER) {
				EnvIsOL8CloudShell = true
			}
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
	var containerEngineType string
	if containerEngineType = viper.GetString(ContainerEngineType); containerEngineType == "" {
		containerEngineType = DefaultContainerEngineType
	}
	viper.SetDefault(ContainerEngineType, containerEngineType)

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
		err = WriteConfigValueToConfigFile(ContainerEngineType, DefaultContainerEngineType)
		if err != nil {
			return err
		}

		fmt.Println("current context has been set to default")
		return nil
	}

	viper.Set(CurrentContext, context)
	return nil
}

func ValidateContainerEngineType(containerEngineType string) error {
	switch containerEngineType {
	case "docker", "podman":
		return nil
	default:
		return errors.New("Invalid Container Engine")
	}
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
