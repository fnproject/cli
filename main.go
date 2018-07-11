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
	cli.AppHelpTemplate = `{{"\n"}}{{"\t"}}{{if not .ArgsUsage}}{{boldred .Description}}{{"\t"}}{{boldred "-"}}{{"\t"}}{{boldred "Version "}}{{boldred .Version}}{{"\n"}}{{"\n"}}{{"\t"}}{{bold "ENVIRONMENT VARIABLES"}}{{"\n"}}{{"\t"}}{{"\t"}}FN_API_URL{{"\t"}}{{italic "Fn server address"}}{{"\n"}}{{"\t"}}{{"\t"}}FN_REGISTRY{{"\t"}}{{italic "Docker registry to push images to, use username only to push to Docker Hub - [[registry.hub.docker.com/]USERNAME]"}}{{if .VisibleCommands}}{{"\n"}}{{"\n"}}{{"\t"}}{{bold "GENERAL COMMANDS"}}{{end}}{{else}}{{range .VisibleCategories}}{{if .Name}}{{bold .Name}}{{end}}{{end}}{{"\n"}}{{"\t"}}{{"\t"}}{{boldcyan .HelpName}}{{if .Usage}}{{" - "}}{{italic .Usage}}{{"\n"}}{{"\n"}}{{"\t"}}{{bold "USAGE"}}{{"\n"}}{{"\t"}}{{"\t"}}{{boldcyan .HelpName}}{{if .VisibleFlags}}{{cyan " [global options]"}}{{end}} {{if .ArgsUsage}}{{brightred .ArgsUsage}}{{end}}{{if .Flags}}{{yellow " [command options]"}}{{end}}{{if .Description}}{{"\n"}}{{"\n"}}{{"\t"}}{{bold "DESCRIPTION"}}{{"\n"}}{{"\t"}}{{"\t"}}{{.Description}}{{end}}{{end}}{{end}}{{range .VisibleCategories}}{{if .Name}}{{"\n"}}{{"\n"}}{{"\t"}}{{bold .Name}}{{end}}{{range .VisibleCommands}}{{"\n"}}{{"\t"}}{{"\t"}}{{join .Names ", "}}{{"\t"}}{{"\t"}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{if .VisibleFlags}}{{"\n"}}{{"\n"}}{{"\t"}}{{if not .ArgsUsage}}{{bold "GLOBAL OPTIONS"}}{{else}}{{bold "COMMAND OPTIONS"}}{{end}}{{"\n"}}{{"\t"}}{{"\t"}}{{range $index, $option := .VisibleFlags}}{{if $index}}{{"\n"}}{{"\t"}}{{"\t"}}{{end}}{{$option}}{{end}}{{end}}{{"\n"}}{{"\n"}}{{"\t"}}{{bold "FURTHER HELP:"}}{{"\t"}}{{italic "See "}}{{"'"}}{{brightcyan "fn <command> --help"}}{{"'"}}{{italic " for more information about a command."}}{{if not .ArgsUsage}}{{"\n"}}{{"\n"}}{{"\t"}}{{bold "LEARN MORE:"}}{{"\t"}}{{"\t"}}{{underlinebrightred "https://github.com/fnproject/fn"}}{{else}}{{end}}
	`
	// Override command template
	// SubcommandHelpTemplate is the text template for the subcommand help topic.
	// cli.go uses text/template to render templates. You can
	// render custom help text by setting this variable.
	cli.SubcommandHelpTemplate = `{{"\n"}}{{range .VisibleCategories}}{{if .Name}}{{"\t"}}{{bold .Name}}{{end}}{{end}}{{"\n"}}{{"\t"}}{{"\t"}}{{boldcyan .HelpName}}{{if .Usage}}{{"-"}}{{italic .Usage}}{{"\n"}}{{"\n"}}{{"\t"}}{{bold "USAGE"}}{{"\n"}}{{"\t"}}{{"\t"}}{{boldcyan .HelpName}}{{if .VisibleFlags}}{{cyan " [global options] "}}{{end}}{{if .ArgsUsage}}{{brightred .ArgsUsage}}{{end}}{{if .Flags}}{{yellow " [command options]"}}{{end}}{{end}}{{if .Description}}{{"\n"}}{{"\n"}}{{"\t"}}{{bold "DESCRIPTION"}}{{"\n"}}{{"\t"}}{{"\t"}}{{.Description}}{{end}}{{if .Commands}}{{"\n"}}{{"\n"}}{{"\t"}}{{bold "SUBCOMMANDS"}}{{range .Commands}}{{"\n"}}{{"\t"}}{{"\t"}}{{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{if .VisibleFlags}}{{"\n"}}{{"\n"}}{{"\t"}}{{bold "COMMAND OPTIONS"}}{{range .VisibleFlags}}{{"\n"}}{{"\t"}}{{"\t"}}{{.}}{{end}}{{end}}{{if .Commands}}{{"\n"}}{{"\n"}}{{"\t"}}{{bold "FURTHER HELP:"}}{{"\t"}}{{italic "See "}}{{"'"}}{{brightcyan "fn <command> --help"}}{{"'"}}{{italic " for more information about a command."}}{{end}}
`
	//Override command template
	// CommandHelpTemplate is the text template for the command help topic.
	// cli.go uses text/template to render templates. You can
	// render custom help text by setting this variable.
	cli.CommandHelpTemplate = `{{"\n"}}{{if .Category}}{{"\t"}}{{bold .Category}}{{end}}{{"\n"}}{{"\t"}}{{"\t"}}{{boldcyan .HelpName}}{{if .Usage}}{{" - "}}{{italic .Usage}}{{"\n"}}{{"\n"}}{{"\t"}}{{bold "USAGE"}}{{"\n"}}{{"\t"}}{{"\t"}}{{boldcyan .HelpName}}{{cyan " [global options] "}}{{if .ArgsUsage}}{{brightred .ArgsUsage}}{{end}}{{if .Flags}}{{yellow " [command options]"}}{{end}}{{end}}{{if .Description}}{{"\n"}}{{"\n"}}{{"\t"}}{{bold "DESCRIPTION"}}{{"\n"}}{{"\t"}}{{"\t"}}{{.Description}}{{end}}{{if .Subcommands}}{{"\n"}}{{"\n"}}{{"\t"}}{{bold "SUBCOMMANDS"}}{{range .Subcommands}}{{"\n"}}{{"\t"}}{{"\t"}}{{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{if .VisibleFlags}}{{"\n"}}{{"\n"}}{{"\t"}}{{bold "COMMAND OPTIONS"}}{{"\n"}}{{"\t"}}{{"\t"}}{{range .Flags}}{{.}}{{"\n"}}{{"\t"}}{{"\t"}}{{end}}{{if .Subcommands}}{{"\n"}}{{"\n"}}{{"\t"}}{{bold "FURTHER HELP:"}}{{"\t"}}{{italic "See "}}){{"'"}}{{brightcyan "fn <command> --help"}}{{"'"}}{{italic "for more information about a command."}}{{end}}{{end}}
`

	//fmt.Println("AH .......... !!!")

	app.CommandNotFound = func(c *cli.Context, cmd string) {
		fmt.Fprintf(os.Stderr, "Command not found: \"%v\" -- see `fn --help` for more information.\n", cmd)
		fmt.Fprintf(os.Stderr, "Note: the fn CLI command structure has changed, change your command to use the new structure.\n")
	}

	app.Commands = append(app.Commands, commands.GetCommands(commands.Commands)...)
	app.Commands = append(app.Commands, VersionCommand())

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))

	prepareCmdArgsValidation(app.Commands)

	cli.HelpPrinter = func(w io.Writer, templ string, data interface{}) {
		doStuff(w, templ, data, colour.Colours)
	}

	return app
}

func doStuff(out io.Writer, templ string, data interface{}, customFunc map[string]interface{}) {
	funcMap := colour.Colours
	if customFunc != nil {
		for key, value := range customFunc {
			funcMap[key] = value
		}
	}

	w := tabwriter.NewWriter(out, 1, 8, 2, ' ', 0)
	t := template.Must(template.New("temp").Funcs(funcMap).Parse(templ))
	err := t.Execute(w, data)
	if err != nil {
		// If the writer is closed, t.Execute will fail, and there's nothing
		// we can do to recover.
		// if os.Getenv("CLI_TEMPLATE_ERROR_DEBUG") != "" {
		// 	fmt.Fprintf(ErrWriter, "CLI TEMPLATE ERROR: %#v\n", err)
		// }
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
	//fmt.Println(colour.X)

	if err != nil {
		// TODO: this doesn't seem to get called even when an error returns from a command, but maybe urfave is doing a non zero exit anyways? nope: https://github.com/urfave/cli/issues/610
		fmt.Fprintf(os.Stderr, "ERROR: %v\n", err)
		fmt.Fprintf(os.Stderr, "Client version: %s\n", Version)
		os.Exit(1)
	}

}
