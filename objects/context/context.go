package context

import (
	"encoding/json"
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
	"github.com/fnproject/fn_go/provider"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

var contextsPath = config.GetContextsPath()
var fileExtension = ".yaml"

type ContextMap config.ContextMap

func create(c *cli.Context) error {
	context := c.Args().Get(0)

	err := ValidateContextName(context)
	if err != nil {
		return err
	}

	providerId := config.DefaultProvider
	if cProvider := c.String("provider"); cProvider != "" {
		providerId = cProvider
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
		return errors.New("Context already exists")
	}
	path := createFilePath(context + fileExtension)
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	contextValues := &config.ContextMap{
		config.ContextProvider: providerId,
		provider.CfgFnAPIURL:   apiURL,
		config.EnvFnRegistry:   registry,
	}

	err = config.WriteYamlFile(file.Name(), contextValues)
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
		return errors.New("Context file not found")
	}

	if context == viper.GetString(config.CurrentContext) {
		return fmt.Errorf("Cannot delete the current context: %v", context)
	}

	if context == "default" {
		return errors.New("Cannot delete default context")
	}

	path := createFilePath(context + fileExtension)
	err := os.Remove(path)
	if err != nil {
		return err
	}

	fmt.Printf("Successfully deleted context %v \n", context)
	return nil
}

func inspect(c *cli.Context) error {
	context := c.Args().Get(0)
	if context == "" {
		if currentContext := viper.GetString(config.CurrentContext); currentContext != "" {
			context = currentContext
		} else {
			return errors.New("no context is set, please provider a context to inspect.")
		}
	}
	return printContext(context)
}

func printContext(context string) error {
	if check, err := checkContextFileExists(context); !check {
		if err != nil {
			return err
		}
		return errors.New("Context file not found")
	}

	contextPath := filepath.Join(config.GetHomeDir(), ".fn", "contexts", (context + fileExtension))
	b, err := ioutil.ReadFile(contextPath)
	if err != nil {
		return err
	}

	fmt.Printf("Current context: %s\n\n", context)
	fmt.Println(string(b))
	return nil
}

func use(c *cli.Context) error {
	context := c.Args().Get(0)

	if check, err := checkContextFileExists(context); !check {
		if err != nil {
			return err
		}
		return errors.New("Context file not found")
	}

	if context == viper.GetString(config.CurrentContext) {
		return fmt.Errorf("Context %v currently in use", context)
	}

	err := config.WriteConfigValueToConfigFile(config.CurrentContext, context)
	if err != nil {
		return err
	}
	viper.Set(config.CurrentContext, context)

	fmt.Printf("Now using context: %v \n", context)
	return nil
}

func unset(c *cli.Context) error {
	if currentContext := viper.GetString(config.CurrentContext); currentContext == "" {
		return errors.New("No context currently in use")
	}

	err := config.WriteConfigValueToConfigFile(config.CurrentContext, "")
	if err != nil {
		return err
	}

	fmt.Printf("Successfully unset current context \n")
	return nil
}

func printContexts(c *cli.Context, contexts []*ContextInfo) error {
	outputFormat := strings.ToLower(c.String("output"))
	if outputFormat == "json" {
		b, err := json.MarshalIndent(contexts, "", "    ")
		if err != nil {
			return err
		}
		fmt.Fprint(os.Stdout, string(b))
	} else {
		w := tabwriter.NewWriter(os.Stdout, 0, 8, 1, '\t', 0)
		fmt.Fprint(w, "CURRENT", "\t", "NAME", "\t", "PROVIDER", "\t", "API URL", "\t", "REGISTRY", "\n")

		for _, ctx := range contexts {
			current := ""
			if ctx.Current {
				current = "*"
			}
			fmt.Fprint(w, current, "\t", ctx.Name, "\t", ctx.Provider, "\t", ctx.APIURL, "\t", ctx.Registry, "\n")
		}
		if err := w.Flush(); err != nil {
			return err
		}
	}
	return nil
}

func list(c *cli.Context) error {
	contexts, err := getAvailableContexts()
	if err != nil {
		return err
	}
	return printContexts(c, contexts)
}

func (ctxMap *ContextMap) update(c *cli.Context) error {
	key := c.Args().Get(0)
	value := c.Args().Get(1)
	err := ctxMap.Set(key, value)
	if err != nil {
		return err
	}

	fmt.Printf("Current context updated %v with %v\n", key, value)
	return err
}

func createFilePath(filename string) string {
	home := config.GetHomeDir()
	return filepath.Join(home, contextsPath, filename)
}

func checkContextFileExists(filename string) (bool, error) {
	path := createFilePath(filename + fileExtension)

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false, err
	}
	return true, nil
}

func getAvailableContexts() ([]*ContextInfo, error) {
	home := config.GetHomeDir()
	files, err := ioutil.ReadDir(filepath.Join(home, contextsPath))
	if err != nil {
		return nil, err
	}

	var contexts []*ContextInfo
	for _, f := range files {
		c, err := NewContextInfo(f)
		if err != nil {
			return nil, err
		}
		contexts = append(contexts, c)
	}
	return contexts, err
}

func ValidateAPIURL(apiURL string) error {
	if !strings.Contains(apiURL, "://") {
		return errors.New("Invalid Fn API URL: does not contain ://")
	}

	_, err := url.Parse(apiURL)
	if err != nil {
		return fmt.Errorf("Invalid Fn API URL: %s", err)
	}
	return nil
}

func ValidateContextName(context string) error {
	re := regexp.MustCompile("[^a-zA-Z0-9_-]+")

	for range re.FindAllString(context, -1) {
		return errors.New("Please enter a context name with only Alphanumeric, _, or -")
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

	file, err := config.DecodeYAMLFile(f.Name())
	if err != nil {
		return err
	}

	if key == provider.CfgFnAPIURL {
		err := ValidateAPIURL(value)
		if err != nil {
			return err
		}
	}

	(*file)[key] = value
	return config.WriteYamlFile(f.Name(), file)
}
