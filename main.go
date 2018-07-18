package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/fnproject/cli/commands"
	"github.com/fnproject/cli/common/color"
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

	// Override app template
	// AppHelpTemplate is the text template for the Default help topic.
	// cli.go uses text/template to render templates. You can render custom help text by setting this variable.
	cli.AppHelpTemplate = `
{{if not .ArgsUsage}}{{boldred .Description}}	{{boldred "-"}}	{{boldred "Version "}}{{boldred .Version}}
	
{{bold "ENVIRONMENT VARIABLES"}}
	FN_API_URL		 {{italic "Fn server address"}}
	FN_REGISTRY		 {{italic "Docker registry to push images to, use username only to push to Docker Hub - [[registry.hub.docker.com/]USERNAME]"}}{{if .VisibleCommands}}
		
{{bold "GENERAL COMMANDS"}}{{end}}{{else}}{{range .VisibleCategories}}{{if .Name}}{{bold .Name}}{{end}}{{end}}
	{{boldcyan .HelpName}}{{if .Usage}}{{" - "}}{{italic .Usage}}
	
{{bold "USAGE"}}
	{{boldcyan .HelpName}}{{if .VisibleFlags}}{{cyan " [global options]"}}{{end}} {{if .ArgsUsage}}{{brightred .ArgsUsage}}{{end}}{{if .Flags}}{{yellow " [command options]"}}{{end}}{{if .Description}}
	
{{bold "DESCRIPTION"}}
	{{.Description}}{{end}}{{end}}{{end}}{{range .VisibleCategories}}{{if .Name}}
	
{{bold .Name}}{{end}}{{range .VisibleCommands}}
	{{join .Names ", "}}				 {{.Usage}}{{end}}{{end}}{{if .VisibleFlags}}
		
{{if not .ArgsUsage}}{{bold "GLOBAL OPTIONS"}}{{else}}{{bold "COMMAND OPTIONS"}}{{end}}
  {{range $index, $option := .VisibleFlags}}{{if $index}}
  {{end}}{{$option}}{{end}}{{end}}
		
{{bold "FURTHER HELP:"}}	{{italic "See "}}{{"'"}}{{brightcyan .HelpName}}{{brightcyan " <command> --help"}}{{"'"}}{{italic " for more information about a command."}}{{if not .ArgsUsage}}
	
{{bold "LEARN MORE:"}}	{{underlinebrightred "https://github.com/fnproject/fn"}}{{else}}{{end}}
	`
	// Override subcommand template
	// SubcommandHelpTemplate is the text template for the subcommand help topic.
	// cli.go uses text/template to render templates. You can render custom help text by setting this variable.
	cli.SubcommandHelpTemplate = `
{{range .VisibleCategories}}{{if .Name}}{{bold .Name}}{{end}}{{end}}
	{{boldcyan .HelpName}}{{if .Usage}}{{" - "}}{{italic .Usage}}
		
{{bold "USAGE"}}
	{{boldcyan .HelpName}}{{if .VisibleFlags}}{{cyan " [global options] "}}{{end}}{{if .ArgsUsage}}{{brightred .ArgsUsage}}{{end}}{{if .Flags}}{{yellow " [command options]"}}{{end}}{{end}}{{if .Description}}
		
{{bold "DESCRIPTION"}}
	{{.Description}}{{end}}{{if .Commands}}
		
{{bold "SUBCOMMANDS"}}{{range .Commands}}
	{{join .Names ", "}}			{{.Usage}}{{end}}{{end}}{{if .VisibleFlags}}
		
{{bold "COMMAND OPTIONS"}}{{range .VisibleFlags}}
	{{.}}{{end}}{{end}}{{if .Commands}}

{{bold "FURTHER HELP:"}}	{{italic "See "}}{{"'"}}{{brightcyan .HelpName}}{{brightcyan " <subcommand> --help"}}{{"'"}}{{italic " for more information about a command."}}{{end}}
`
	//Override command template
	// CommandHelpTemplate is the text template for the command help topic.
	// cli.go uses text/template to render templates. You can render custom help text by setting this variable.
	cli.CommandHelpTemplate = `
{{if .Category}}{{bold .Category}}{{end}}
	{{boldcyan .HelpName}}{{if .Usage}}{{" - "}}{{italic .Usage}}
		
{{bold "USAGE"}}
	{{boldcyan .HelpName}}{{cyan " [global options] "}}{{if .ArgsUsage}}{{brightred .ArgsUsage}}{{end}}{{if .Flags}}{{yellow " [command options]"}}{{end}}{{end}}{{if .Description}}
		
{{bold "DESCRIPTION"}}
	{{.Description}}{{end}}{{if .Subcommands}}
		
{{bold "SUBCOMMANDS"}}{{range .Subcommands}}
	{{join .Names ", "}}			{{.Usage}}{{end}}{{end}}{{if .VisibleFlags}}
		
{{bold "COMMAND OPTIONS"}}
	{{range .Flags}}{{.}}
	{{end}}{{if .Subcommands}}
		
{{bold "FURTHER HELP:"}}	{{italic "See "}}){{"'"}}{{brightcyan .HelpName}}{{brightcyan " <subcommand> --help"}}{{"'"}}{{italic "for more information about a command."}}{{end}}{{end}}
`
	app.CommandNotFound = func(c *cli.Context, cmd string) {
		fmt.Fprintf(os.Stderr, "\n'"+color.Red("%v")+"' is not a Fn Command: "+color.Italic("note the fn CLI command structure has changed, please change your command to use the new structure.\n\n"), cmd)
		fmt.Fprintf(os.Stderr, color.Bold("FURTHER HELP: ")+color.Italic("See ")+"'"+color.BrightCyan("fn <command> --help")+"'"+color.Italic(" for more information.\n"))
	}

	app.Commands = append(app.Commands, commands.GetCommands(commands.Commands)...)
	app.Commands = append(app.Commands, VersionCommand())

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	prepareCmdArgsValidation(app.Commands)

	cli.HelpPrinter = func(w io.Writer, templ string, data interface{}) {
		printHelpCustom(w, templ, data, color.Colors)
	}

	return app
}

//Override function for customised app template
func printHelpCustom(out io.Writer, templ string, data interface{}, customFunc map[string]interface{}) {
	funcMap := color.Colors
	for key, value := range customFunc {
		funcMap[key] = value
	}

	w := tabwriter.NewWriter(out, 1, 8, 2, ' ', 0)
	t := template.Must(template.New("temp").Funcs(funcMap).Parse(templ))
	err := t.Execute(w, data)
	if err != nil {
		fmt.Println("CLI TEMPLATE ERROR:")
		return
	}
	w.Flush()
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
