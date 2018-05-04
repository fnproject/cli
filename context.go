package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/tabwriter"

	"github.com/fnproject/cli/config"
	"github.com/fnproject/cli/utils"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
	yaml "gopkg.in/yaml.v2"
)

var contextsPath = config.GetContextsPath()
var fileExtension = ".yaml"

type ContextMap utils.ContextMap

func contextCmd() cli.Command {
	ctxMap := ContextMap{}
	return cli.Command{
		Name:  "context",
		Usage: "manage context",
		Subcommands: []cli.Command{
			{
				Name:      "create",
				Aliases:   []string{"c"},
				Usage:     "create a new context",
				ArgsUsage: "<context>",
				Action:    createCtx,
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
				Action:    deleteCtx,
			},
			{
				Name:    "list",
				Aliases: []string{"l"},
				Usage:   "list contexts",
				Action:  listCtx,
			},
			{
				Name:      "update",
				Usage:     "update context files",
				ArgsUsage: "<key> <value>",
				Action:    ctxMap.updateCtx,
			},
			{
				Name:      "use",
				Aliases:   []string{"u"},
				Usage:     "use context for future invocations",
				ArgsUsage: "<context>",
				Action:    useCtx,
			},
			{
				Name:   "unset",
				Usage:  "unset current-context",
				Action: unsetCtx,
			},
		},
	}
}

func createCtx(c *cli.Context) error {
	context := c.Args().Get(0)

	err := ValidateContextName(context)
	if err != nil {
		return err
	}

	provider := config.DefaultProvider
	if cProvider := c.String("provider"); cProvider != "" {
		provider = cProvider
	}

	apiURL := ""
	if cApiURL := c.String("api-url"); cApiURL != "" {
		err = ValidateAPIURL(cApiURL)
		if err != nil {
			return err
		}
		apiURL = cApiURL
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
	path := createFilePath(context + fileExtension)
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	contextValues := &utils.ContextMap{
		config.ContextProvider: provider,
		config.EnvFnAPIURL:     apiURL,
		config.EnvFnRegistry:   registry,
	}

	err = utils.WriteYamlFile(file, contextValues)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully created context: %v \n", context)
	return nil
}

func deleteCtx(c *cli.Context) error {
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

	path := createFilePath(context + fileExtension)
	err := os.Remove(path)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully deleted context %v \n", context)
	return nil
}

func useCtx(c *cli.Context) error {
	context := c.Args().Get(0)

	if check, err := checkContextFileExists(context); !check {
		if err != nil {
			return err
		}
		return errors.New("context file not found")
	}

	if context == viper.GetString(config.CurrentContext) {
		return fmt.Errorf("context %v currently in use", context)
	}

	err := config.WriteCurrentContextToConfigFile(context)
	if err != nil {
		return err
	}
	viper.Set(config.CurrentContext, context)

	fmt.Printf("Now using context: %v \n", context)
	return nil
}

func unsetCtx(c *cli.Context) error {
	if currentContext := viper.GetString(config.CurrentContext); currentContext == "" {
		return errors.New("no context currently in use")
	}

	err := config.WriteCurrentContextToConfigFile("")
	if err != nil {
		return err
	}

	fmt.Printf("Successfully unset current context \n")
	return nil
}

func listCtx(c *cli.Context) error {
	currentContext := viper.GetString(config.CurrentContext)
	files, err := getAvailableContexts()
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
	fmt.Fprint(w, "CURRENT", "\t", "NAME", "\t", "PROVIDER", "\t", "API URL", "\t", "REGISTRY", "\n")

	for _, f := range files {
		current := ""
		home := utils.GetHomeDir()
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

		name := strings.Replace(f.Name(), fileExtension, "", 1)
		if currentContext == name {
			current = "*"
		}
		fmt.Fprint(w, current, "\t", name, "\t", v.ContextProvider, "\t", v.EnvFnAPIURL, "\t", v.EnvFnRegistry, "\n")
	}
	return w.Flush()
}

func (ctxMap *ContextMap) updateCtx(c *cli.Context) error {
	key := c.Args().Get(0)
	value := c.Args().Get(1)
	err := ctxMap.Set(key, value)
	if err != nil {
		return err
	}

	fmt.Printf("current context updated %v with %v\n", key, value)
	return err
}

func createFilePath(filename string) string {
	home := utils.GetHomeDir()
	return filepath.Join(home, contextsPath, filename)
}

func checkContextFileExists(filename string) (bool, error) {
	path := createFilePath(filename + fileExtension)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, err
	}
	return true, nil
}

func getAvailableContexts() ([]os.FileInfo, error) {
	home := utils.GetHomeDir()
	files, err := ioutil.ReadDir(filepath.Join(home, contextsPath))
	return files, err
}

func ValidateAPIURL(apiURL string) error {
	if !strings.Contains(apiURL, "://") {
		return errors.New("invalid Fn API URL: does not contain ://")
	}

	_, err := url.Parse(apiURL)
	if err != nil {
		return fmt.Errorf("invalid Fn API URL: %s", err)
	}
	return nil
}

func ValidateContextName(context string) error {
	re := regexp.MustCompile("[^a-zA-Z0-9_-]+")

	for range re.FindAllString(context, -1) {
		return errors.New("please enter a context name with only Alphanumeric, _, or -")
	}
	return nil
}

func (ctxMap *ContextMap) Set(key, value string) error {
	contextFilePath := createFilePath(viper.GetString(config.CurrentContext) + fileExtension)
	f, err := os.OpenFile(contextFilePath, os.O_RDWR, config.ReadWritePerms)
	if err != nil {
		return err
	}
	defer f.Close()

	file, err := utils.DecodeYAMLFile(f)
	if err != nil {
		return err
	}

	if key == config.EnvFnAPIURL {
		err := ValidateAPIURL(value)
		if err != nil {
			return err
		}
	}

	(*file)[key] = value
	return utils.WriteYamlFile(f, file)
}
