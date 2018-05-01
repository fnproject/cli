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

var contextsPath = config.GetContextsPath()

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
			{
				Name:   "unset",
				Usage:  "unset current-context",
				Action: unset,
			},
		},
	}
}

func create(c *cli.Context) error {
	context := c.Args().Get(0)

	err := ValidateContextName(context)
	if err != nil {
		return err
	}

	provider := config.DefaultProvider
	if cProvider := c.String("provider"); cProvider != "" {
		provider = cProvider
	}

	apiUrl := ""
	if cApiUrl := c.String("api-url"); cApiUrl != "" {
		apiUrl = cApiUrl
	}

	registry := ""
	if cRegistry := c.String("registry"); cRegistry != "" {
		registry = cRegistry
	}

	if check, err := checkContextFileExists(context); check {
		if err != nil {
			return err
		}
		return errors.New("context already exists")

	}
	path, err := createFilePath(context)
	if err != nil {
		return err
	}

	contextValues := map[string]string{
		config.ContextProvider: provider,
		config.EnvFnAPIURL:     apiUrl,
		config.EnvFnRegistry:   registry,
	}

	err = config.WriteYamlFile(path, contextValues)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully created context: %v \n", context)
	return nil
}

func delete(c *cli.Context) error {
	context := c.Args().Get(0)

	if check, err := checkContextFileExists(context); !check {
		if err != nil {
			return err
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
		return err
	}

	err = os.Remove(path)
	if err != nil {
		return err
	}

	fmt.Printf("Context %v deleted \n", context)
	return nil
}

func set(c *cli.Context) error {
	context := c.Args().Get(0)

	if check, err := checkContextFileExists(context); !check {
		if err != nil {
			return err
		}
		return errors.New("context file not found")
	}

	if context == viper.GetString(config.CurrentContext) {
		return fmt.Errorf("context %v already set", context)
	}

	err := config.WriteCurrentContextToConfigFile(context)
	if err != nil {
		return err
	}
	viper.Set(config.CurrentContext, context)

	fmt.Printf("Successfully set context: %v \n", context)
	return nil
}

func unset(c *cli.Context) error {
	if currentContext := viper.GetString(config.CurrentContext); currentContext == "" {
		return errors.New("no context set")
	}

	err := config.WriteCurrentContextToConfigFile("")
	if err != nil {
		return err
	}

	fmt.Printf("Successfully unset current context \n")
	return nil
}

func list(c *cli.Context) error {
	currentContext := viper.GetString(config.CurrentContext)
	files, err := getAvailableContexts()
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprint(w, "CURRENT", "\t", "NAME", "\t", "PROVIDER", "\t", "API URL", "\t", "REGISTRY", "\n")

	for _, f := range files {
		current := ""
		home := config.GetHomeDir()
		path := filepath.Join(home, contextsPath, f.Name())
		yamlFile, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}

		v := config.ContextFile{}
		err = yaml.Unmarshal(yamlFile, &v)
		if err != nil {
			return err
		}

		name := strings.Replace(f.Name(), ".yaml", "", 1)
		if currentContext == name {
			current = "*"
		}
		fmt.Fprint(w, current, "\t", name, "\t", v.ContextProvider, "\t", v.EnvFnAPIURL, "\t", v.EnvFnRegistry, "\n")
	}
	return w.Flush()
}

func createFilePath(filename string) (string, error) {
	contextFileName := filename + ".yaml"
	home := config.GetHomeDir()
	path := filepath.Join(home, contextsPath, contextFileName)
	return path, nil
}

func checkContextFileExists(filename string) (bool, error) {
	path, err := createFilePath(filename)
	if err != nil {
		return false, err
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, err
	}
	return true, nil
}

func getAvailableContexts() ([]os.FileInfo, error) {
	home := config.GetHomeDir()
	files, err := ioutil.ReadDir(filepath.Join(home, contextsPath))
	if err != nil {
		return nil, err
	}

	return files, nil
}

func ValidateContextName(context string) error {
	re := regexp.MustCompile("[^a-zA-Z0-9_-]+")

	for range re.FindAllString(context, -1) {
		return errors.New("please enter a context name with only Alphanumeric, _, or -")
	}
	return nil
}
