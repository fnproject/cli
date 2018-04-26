package main

import (
	"errors"
	"regexp"
	"strings"

	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"

	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/fnproject/cli/config"
	"github.com/urfave/cli"
)

var contextsPath = filepath.Join(config.RootConfigPathName, config.ContextsPathName)
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
				ArgsUsage: "<context>",
				Action:    create,
				Flags: []cli.Flag{
					cli.StringFlag{
						Name:  "provider",
						Usage: "context provider",
					},
					cli.StringFlag{
						Name:  "api-url",
						Usage: "context api url",
					},
					cli.StringFlag{
						Name:  "registry",
						Usage: "context registry",
					},
				},
			},
			{
				Name:      "delete",
				Aliases:   []string{"d"},
				Usage:     "delete a context",
				ArgsUsage: "<context>",
				Action:    delete,
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

func create(c *cli.Context) error {
	context := c.Args().Get(0)

	re := regexp.MustCompile("[^a-zA-Z0-9_-]+")

	for range re.FindAllString(context, -1) {
		fmt.Fprintf(os.Stderr, "please enter a context name with ASCII characters only \n")
		os.Exit(1)
	}

	provider := config.DefaultProvider
	if cProvider := c.String("provider"); cProvider != "" {
		provider = cProvider
	}

	apiUrl := config.DefaultLocalApiUrl
	if cApiUrl := c.String("api-url"); cApiUrl != "" {
		apiUrl = cApiUrl
	}

	registry := ""
	if cRegistry := c.String("registry"); cRegistry != "" {
		registry = cRegistry
	}

	if check, err := checkContextFileExists(context); check {
		if err != nil {
			return fmt.Errorf("%v", err)
		}
		return errors.New("context already exists")

	}
	path, err := createFilePath(context)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	_, err = os.Create(path)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	contextValues := map[string]string{
		config.ContextProvider: provider,
		config.EnvFnAPIURL:     apiUrl,
		config.EnvFnRegistry:   registry,
	}

	err = config.WriteYamlFile(path, contextValues)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	fmt.Printf("Successfully created context: %v \n", context)
	return nil
}

func delete(c *cli.Context) error {
	context := c.Args().Get(0)

	if check, err := checkContextFileExists(context); !check {
		if err != nil {
			return fmt.Errorf("%v", err)
		}
		return errors.New("context file not found")
	}

	if context == viper.GetString(config.CurrentContext) {
		return fmt.Errorf("can not delete the current context: %v", context)
	}

	if context == "default" {
		return errors.New("can not delete default context")
	}

	path, err := createFilePath(context)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	os.Remove(path)
	fmt.Printf("Context %v deleted \n", context)
	return nil
}

func set(c *cli.Context) error {
	context := c.Args().Get(0)

	home, err := config.GetHomeDir()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	configFilePath := filepath.Join(home, config.RootConfigPathName, config.ContextConfigFileName)

	if check, err := checkContextFileExists(context); !check {
		if err != nil {
			return fmt.Errorf("%v", err)
		}
		return errors.New("context file not found")
	}

	if context == viper.GetString(config.CurrentContext) {
		return fmt.Errorf("context %v already set", context)
	}

	viper.Set(config.CurrentContext, context)

	configCurrentContext := map[string]string{config.CurrentContext: context}
	err = config.WriteYamlFile(configFilePath, configCurrentContext)
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	fmt.Printf("Successfully set context: %v \n", context)
	return nil
}

func list(c *cli.Context) error {
	currentContext := viper.GetString(config.CurrentContext)
	files, err := getAvailableContexts()
	if err != nil {
		return fmt.Errorf("%v", err)
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprint(w, "CURRENT", "\t", "NAME", "\t", "PROVIDER", "\t", "API URL", "\t", "REGISTRY", "\n")

	for _, f := range files {
		current := ""

		home, err := config.GetHomeDir()
		if err != nil {
			return fmt.Errorf("%v", err)
		}

		path := filepath.Join(home, contextsPath, f.Name())
		yamlFile, err := ioutil.ReadFile(path)
		if err != nil {
			return fmt.Errorf("%v", err)
		}

		v := config.ContextFile{}
		err = yaml.Unmarshal(yamlFile, &v)

		name := strings.Replace(f.Name(), ".yaml", "", 1)
		if currentContext == name {
			current = "*"
		}
		fmt.Fprint(w, current, "\t", name, "\t", v.ContextProvider, "\t", v.EnvFnAPIURL, "\t", v.EnvFnRegistry, "\n")
	}
	w.Flush()
	return nil
}

func createFilePath(filename string) (string, error) {
	contextFileName := filename + ".yaml"
	home, err := config.GetHomeDir()
	if err != nil {
		return "", fmt.Errorf("%v", err)
	}
	path := filepath.Join(home, contextsPath, contextFileName)
	return path, nil
}

func checkContextFileExists(filename string) (bool, error) {
	path, err := createFilePath(filename)
	if err != nil {
		return false, fmt.Errorf("%v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, nil
	}
	return true, nil
}

func getAvailableContexts() ([]os.FileInfo, error) {
	home, err := config.GetHomeDir()
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}

	files, err := ioutil.ReadDir(filepath.Join(home, contextsPath))
	if err != nil {
		return nil, err
	}

	return files, nil
}
