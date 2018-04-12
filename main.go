package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

var aliases map[string]cli.Command

func aliasesFn() []cli.Command {
	cmds := []cli.Command{}
	for alias, cmd := range aliases {
		cmd.Name = alias
		cmd.Hidden = true
		cmds = append(cmds, cmd)
	}
	return cmds
}

func newFn() *cli.App {
	app := cli.NewApp()
	app.Name = "fn"
	app.Version = Version
	app.Authors = []cli.Author{{Name: "Fn Project"}}
	app.Description = "Fn command line tool"
	app.Before = func(c *cli.Context) error {
		loadConfiguration(c)
		commandArgOverrides(c)
		return nil
	}
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose,v", // v is taken for version by default with urfave/cli
			Usage: "Use --verbose to enable verbose mode for debugging",
		},
		cli.StringFlag{
			Name:  "context",
			Usage: "Use --context to select context configuration file",
		},
		cli.StringFlag{
			Name:  "registry",
			Usage: "Use --registry to select registry",
		},
	}
	cli.VersionFlag = cli.BoolFlag{
		Name:  "version",
		Usage: "print only the version",
	}
	cli.AppHelpTemplate = `{{.Name}} {{.Version}}{{if .Description}}

{{.Description}}{{end}}

ENVIRONMENT VARIABLES:
   FN_API_URL - Fn server address
   FN_REGISTRY - Docker registry to push images to, use username only to push to Docker Hub - [[registry.hub.docker.com/]USERNAME]{{if .VisibleCommands}}

COMMANDS:{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{end}}{{range .VisibleCommands}}
     {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{end}}{{if .VisibleFlags}}

GLOBAL OPTIONS:
   {{range $index, $option := .VisibleFlags}}{{if $index}}
   {{end}}{{$option}}{{end}}{{end}}

LEARN MORE:
   https://github.com/fnproject/fn
`

	app.CommandNotFound = func(c *cli.Context, cmd string) {
		fmt.Fprintf(os.Stderr, "Command not found: \"%v\" -- see `fn --help` for more information.\n", cmd)
	}
	app.Commands = []cli.Command{
		startCmd(),
		updateCmd(),
		initFn(),
		apps(),
		routes(),
		images(),
		lambda(),
		version(),
		calls(),
		deploy(),
		logs(),
		testfn(),
		buildServer(),
		contextCmd(),
	}
	app.Commands = append(app.Commands, aliasesFn()...)

	prepareCmdArgsValidation(app.Commands)

	return app
}

func parseArgs(c *cli.Context) ([]string, []string) {
	args := strings.Split(c.Command.ArgsUsage, " ")
	var reqArgs []string
	var optArgs []string
	for _, arg := range args {
		if strings.HasPrefix(arg, "[") {
			optArgs = append(optArgs, arg)
		} else if strings.Trim(arg, " ") != "" {
			reqArgs = append(reqArgs, arg)
		}
	}
	return reqArgs, optArgs
}

func prepareCmdArgsValidation(cmds []cli.Command) {
	// TODO: refactor fn to use urfave/cli.v2
	// v1 doesn't let us validate args before the cmd.Action

	for i, cmd := range cmds {
		prepareCmdArgsValidation(cmd.Subcommands)
		if cmd.Action == nil {
			continue
		}
		action := cmd.Action
		cmd.Action = func(c *cli.Context) error {
			reqArgs, _ := parseArgs(c)
			if c.NArg() < len(reqArgs) {
				var help bytes.Buffer
				cli.HelpPrinter(&help, cli.CommandHelpTemplate, c.Command)
				return fmt.Errorf("Missing required arguments: %s", strings.Join(reqArgs[c.NArg():], " "))
			}
			return cli.HandleAction(action, c)
		}
		cmds[i] = cmd
	}
}

func init() {
	viper.AutomaticEnv() // read in environment variables that match

	viper.SetEnvPrefix("fn")
	viper.SetDefault(envFnAPIURL, "http://localhost:8080")

	// create aliases after api_url set
	aliases = map[string]cli.Command{
		"build":  build(),
		"bump":   bump(),
		"deploy": deploy(),
		"push":   push(),
		"run":    run(),
		"call":   call(),
		"calls":  calls(),
		"logs":   logs(),
	}

	EnsureConfiguration()
}

func loadConfiguration(c *cli.Context) error {
	// Find home directory.
	home, err := homedir.Dir()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	context := ""

	if context = c.String(envFnContext); context == "" {
		viper.AddConfigPath(filepath.Join(home, rootConfigPathName))
		viper.SetConfigName(configName)

		readConfig()

		context = viper.GetString(currentContext)
		if context == "" {
			fmt.Println("Config file does not contain context")
			os.Exit(1)
		}
	}

	fmt.Println("Context: ", context)

	viper.AddConfigPath(filepath.Join(home, rootConfigPathName, contextsPathName))
	viper.SetConfigName(context)
	readConfig()

	fmt.Println("envFnApiUrl", viper.GetString(envFnAPIURL))
	return nil
}

func commandArgOverrides(c *cli.Context) {
	if registry := c.String(envFnRegistry); registry != "" {
		viper.Set(envFnRegistry, registry)
	}
}

func readConfig() {
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func main() {
	app := newFn()
	err := app.Run(os.Args)
	if err != nil {
		// TODO: this doesn't seem to get called even when an error returns from a command, but maybe urfave is doing a non zero exit anyways? nope: https://github.com/urfave/cli/issues/610
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}
