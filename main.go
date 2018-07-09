package main

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fnproject/cli/commands"
	"github.com/fnproject/cli/config"
	"github.com/spf13/viper"
	"github.com/urfave/cli"
)

func newFn() *cli.App {
	app := cli.NewApp()
	app.Name = "fn"
	app.Version = Version
	app.Authors = []cli.Author{{Name: "Fn Project"}}
	app.Description = "Fn Command Line Tool"
	app.EnableBashCompletion = true
	app.Before = func(c *cli.Context) error {
		err := config.LoadConfiguration(c)
		if err != nil {
			return err
		}
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
		Usage: "Display version",
	}

	// AppHelpTemplate is the text template for the Default help topic.
	// cli.go uses text/template to render templates. You can
	// render custom help text by setting this variable.
	cli.AppHelpTemplate = `
	{{if .ArgsUsage}}{{else}}{{.Description}} - Version {{.Version}} 

	ENVIRONMENT VARIABLES:
	   FN_API_URL   Fn server address
	   FN_REGISTRY  Docker registry to push images to, use username only to push to Docker Hub - [[registry.hub.docker.com/]USERNAME]{{end}}{{if .VisibleCommands}}

	{{if .ArgsUsage}}{{else}}GENERAL COMMANDS:{{end}}{{end}}{{range .VisibleCategories}}{{if .Name}}

	{{.Name}}:{{end}}{{range .VisibleCommands}}
		{{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{if .VisibleFlags}}

	GLOBAL OPTIONS:
	   {{range $index, $option := .VisibleFlags}}{{if $index}}
	   {{end}}{{$option}}{{end}}{{end}}

	FURTHER HELP:
	   See 'fn <command> --help' for more information about a command.

	LEARN MORE:
	   https://github.com/fnproject/fn
	`
	//Override command template
	// SubcommandHelpTemplate is the text template for the subcommand help topic.
	// cli.go uses text/template to render templates. You can
	// render custom help text by setting this variable.
	cli.SubcommandHelpTemplate = `{{range .VisibleCategories}}{{if .Name}}
	{{.Name}}:{{end}}{{end}}
		{{ .HelpName}}{{if .Usage}} - {{.Usage}}

	USAGE:
		{{ .HelpName}} {{if .VisibleFlags}}[global options]{{end}} {{if .ArgsUsage}}{{.ArgsUsage}}{{end}} {{if .Flags}}[command options]{{end}} {{end}}{{if .Description}}
	
	DESCRIPTION:
		{{.Description}}{{end}}{{if .Commands}}

	SUBCOMMANDS: {{range .Commands}}
		{{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{if .VisibleFlags}}

	COMMAND OPTIONS: {{range .VisibleFlags}}
		{{.}}{{end}}{{end}}{{if .Commands}}

	FURTHER HELP:
		See '{{ .HelpName}} <subcommand> --help' for more information about a subcommand.{{end}}
	`

	//Override command template
	// CommandHelpTemplate is the text template for the command help topic.
	// cli.go uses text/template to render templates. You can
	// render custom help text by setting this variable.
	cli.CommandHelpTemplate = `{{if .Category}}
	{{.Category}}:{{end}}
		{{.HelpName}}{{if .Usage}} - {{.Usage}}
	
	USAGE:
		{{.HelpName}} [global options] {{if .ArgsUsage}}{{.ArgsUsage}}{{end}} {{if .Flags}}[command options]{{end}}{{end}}{{if .Description}}
	
	DESCRIPTION:
		{{.Description}}{{end}}{{if .Subcommands}}

	SUBCOMMANDS: {{range .Subcommands}}
		{{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{if .VisibleFlags}}
   
	COMMAND OPTIONS:
		{{range .Flags}}{{.}}
		{{end}}{{if .Subcommands}}
	
	FURTHER HELP:
		See '{{ .HelpName}} <subcommand> --help' for more information about a subcommand.{{end}}{{end}}
	`

	app.CommandNotFound = func(c *cli.Context, cmd string) {
		fmt.Fprintf(os.Stderr, "Command not found: \"%v\" -- see `fn --help` for more information.\n", cmd)
		fmt.Fprintf(os.Stderr, "Note: the fn CLI command structure has changed, change your command to use the new structure.\n")
	}

	app.Commands = append(app.Commands, commands.GetCommands(commands.Commands)...)
	app.Commands = append(app.Commands, VersionCommand())

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

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
	err := config.Init()
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		os.Exit(1)
	}
}

func commandArgOverrides(c *cli.Context) {
	if registry := c.String(config.EnvFnRegistry); registry != "" {
		viper.Set(config.EnvFnRegistry, registry)
	}
}

func main() {
	app := newFn()

	err := app.Run(os.Args)
	if err != nil {
		// TODO: this doesn't seem to get called even when an error returns from a command, but maybe urfave is doing a non zero exit anyways? nope: https://github.com/urfave/cli/issues/610
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		fmt.Fprintf(os.Stderr, "Client version: %s\n", Version)
		os.Exit(1)
	}
}
