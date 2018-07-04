package main

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/fnproject/cli/commands"
	"github.com/fnproject/cli/common/colour"
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
	{{"\t"}}{{if not .ArgsUsage}}` + colour.BoldRed("{{.Description}}") + `{{"\t"}}` + colour.BoldRed("-") + `{{"\t"}}` + colour.BoldRed("Version {{.Version}}") + `
	
		{{"\t"}}` + colour.Bold("ENVIRONMENT VARIABLES:") + `
			{{"\t"}}{{"\t"}}FN_API_URL{{"\t"}}` + colour.Italic("Fn server address") + `
			{{"\t"}}{{"\t"}}FN_REGISTRY{{"\t"}}` + colour.Italic("Docker registry to push images to, use username only to push to Docker Hub - [[registry.hub.docker.com/]USERNAME]") + `{{if .VisibleCommands}}
	
		{{"\t"}}` + colour.Bold("GENERAL COMMANDS:") + `{{end}}{{else}}{{range .VisibleCategories}}{{if .Name}}` + colour.Bold("{{.Name}}:") + `{{end}}{{end}}
			{{"\t"}}` + colour.BoldCyan("{{ .HelpName}}") + `{{if .Usage}} - ` + colour.Italic("{{.Usage}}") + ` 
	
		{{"\t"}}` + colour.Bold("USAGE:") + `
			{{"\t"}}` + colour.BoldCyan("{{ .HelpName}}") + ` {{if .VisibleFlags}} ` + colour.Cyan("[global options]") + `{{end}} {{if .ArgsUsage}}` + colour.BrightRed("{{.ArgsUsage}}") + `{{end}} {{if .Flags}}` + colour.Yellow("[command options]") + `{{end}}{{if .Description}}
	  
		{{"\t"}}` + colour.Bold("DESCRIPTION:") + `
			{{"\t"}}{{.Description}}{{end}} {{end}}{{end}}{{range .VisibleCategories}}{{if .Name}}
	
		{{"\t"}}` + colour.Bold("{{.Name}}:") + `{{end}}{{range .VisibleCommands}}
			{{"\t"}}{{"\t"}}{{join .Names ", "}}{{"\t"}}{{"\t"}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{if .VisibleFlags}}
	        
		{{"\t"}}{{if not .ArgsUsage}}` + colour.Bold("GLOBAL OPTIONS:") + `{{else}}` + colour.Bold("COMMAND OPTIONS:") + `{{end}}
			{{"\t"}}{{"\t"}}{{range $index, $option := .VisibleFlags}}{{if $index}}
			{{"\t"}}{{"\t"}}{{end}}{{$option}}{{end}}{{end}}
	
		{{"\t"}}` + colour.Bold("FURTHER HELP:") + `{{"\t"}}` + colour.Italic("See ") + `'` + colour.BrightCyan("fn <command> --help") + `' ` + colour.Italic("for more information about a command.") + `{{if not .ArgsUsage}}
	
		{{"\t"}}` + colour.Bold("LEARN MORE:") + `{{"\t"}}{{"\t"}}` + colour.UnderlineBrightRed("https://github.com/fnproject/fn") + `{{else}}{{end}}
	`
	// Override command template
	// SubcommandHelpTemplate is the text template for the subcommand help topic.
	// cli.go uses text/template to render templates. You can
	// render custom help text by setting this variable.
	cli.SubcommandHelpTemplate = `{{range .VisibleCategories}}{{if .Name}}
    ` + colour.Bold("{{.Name}}:") + `{{end}}{{end}}
        ` + colour.BoldCyan("{{ .HelpName}}") + `{{if .Usage}} - ` + colour.Italic("{{.Usage}}") + `

    ` + colour.Bold("USAGE:") + `
        ` + colour.BoldCyan("{{ .HelpName}}") + ` {{if .VisibleFlags}} ` + colour.Cyan("[global options]") + `{{end}} {{if .ArgsUsage}}` + colour.BrightRed("{{.ArgsUsage}}") + `{{end}} {{if .Flags}}` + colour.Yellow("[command options]") + `{{end}} {{end}}{{if .Description}}
    
    ` + colour.Bold("DESCRIPTION:") + `
        {{.Description}}{{end}}{{if .Commands}}

    ` + colour.Bold("SUBCOMMANDS:") + ` {{range .Commands}}
        {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{if .VisibleFlags}}

    ` + colour.Bold("COMMAND OPTIONS:") + ` {{range .VisibleFlags}}
        {{.}}{{end}}{{end}}{{if .Commands}}

    ` + colour.Bold("FURTHER HELP:") + ` ` + colour.Italic("See ") + `'` + colour.BrightCyan("fn <command> --help") + `' ` + colour.Italic("for more information about a command.") + `{{end}}
`
	//Override command template
	// CommandHelpTemplate is the text template for the command help topic.
	// cli.go uses text/template to render templates. You can
	// render custom help text by setting this variable.
	cli.CommandHelpTemplate = `{{if .Category}}
    ` + colour.Bold("{{.Category}}:") + `{{end}}
    ` + colour.BoldCyan("{{.HelpName}}") + `{{if .Usage}} - ` + colour.Italic("{{.Usage}}") + `
    
    ` + colour.Bold("USAGE:") + `
    ` + colour.BoldCyan("{{ .HelpName}}") + ` ` + colour.Cyan("[global options]") + ` {{if .ArgsUsage}}` + colour.BrightRed("{{.ArgsUsage}}") + `{{end}} {{if .Flags}}` + colour.Yellow("[command options]") + `{{end}}{{end}}{{if .Description}}
    
    ` + colour.Bold("DESCRIPTION:") + `
        {{.Description}}{{end}}{{if .Subcommands}}

    ` + colour.Bold("SUBCOMMANDS:") + ` {{range .Subcommands}}
        {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{if .VisibleFlags}}
   
    ` + colour.Bold("COMMAND OPTIONS:") + `
        {{range .Flags}}{{.}}
        {{end}}{{if .Subcommands}}
    
    ` + colour.Bold("FURTHER HELP:") + ` ` + colour.Italic("See ") + `'` + colour.BrightCyan("fn <command> --help") + `' ` + colour.Italic("for more information about a command.") + `{{end}}{{end}}
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
	//fmt.Println(colour.X)

	if err != nil {
		// TODO: this doesn't seem to get called even when an error returns from a command, but maybe urfave is doing a non zero exit anyways? nope: https://github.com/urfave/cli/issues/610
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		fmt.Fprintf(os.Stderr, "Client version: %s\n", Version)
		os.Exit(1)
	}

}
