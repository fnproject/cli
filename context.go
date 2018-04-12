package main

import (
	//"errors"

	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/urfave/cli"
	yaml "gopkg.in/yaml.v2"
)

var contextsPath string
var defaultContextPath string

func contextCmd() cli.Command {
	return cli.Command{
		Name:  "context",
		Usage: "manage context",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Aliases:   []string{"c"},
				Usage:     "create a new context",
				ArgsUsage: "<context> <provider>",
				Action:    create,
				Flags: []cli.Flag{
					cli.StringSliceFlag{
						Name:  "config",
						Usage: "context configuration",
					},
				},
			},
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "list contexts",
				Action:  list,
			},
			{
				Name:      "set",
				Aliases:   []string{"s"},
				Usage:     "set context for future invocations",
				ArgsUsage: "<context>",
				Action:    set,
			},
		},
	}
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

func create(c *cli.Context) {
	context := c.Args().Get(0)
	provider := c.Args().Get(1)

	fmt.Println("Context: ", context)
	fmt.Println("Provider: ", provider)
}

func set(c *cli.Context) {
	context := c.Args().Get(0)

	fmt.Println("Context: ", context)
}

func list(c *cli.Context) {

}
