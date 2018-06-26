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
    {{if .ArgsUsage}}{{else}} ` + "\x1b[31;1m{{if .ArgsUsage}}{{else}}{{.Description}} - Version {{.Version}}\x1b[0m" + `

    ` + "\x1b[1mENVIRONMENT VARIABLES:\x1b[0m" + `
        FN_API_URL   ` + "\x1b[3mFn server address\x1b[0m" + `
        FN_REGISTRY  ` + "\x1b[3mDocker registry to push images to, use username only to push to Docker Hub - [[registry.hub.docker.com/]USERNAME]\x1b[0m" + `{{end}}{{if .VisibleCommands}}

    ` + "\x1b[1m{{if .ArgsUsage}}{{else}}GENERAL COMMANDS:\x1b[0m" + `{{end}}{{end}}{{range .VisibleCategories}}{{if .Name}}

    ` + "\x1b[1m{{.Name}}:\x1b[0m" + `{{end}}{{range .VisibleCommands}}
        {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{if .VisibleFlags}}

    ` + "\x1b[1mGLOBAL OPTIONS:\x1b[0m" + `
       {{range $index, $option := .VisibleFlags}}{{if $index}}
       {{end}}{{$option}}{{end}}{{end}}

    ` + "\x1b[1mFURTHER HELP:\x1b[0m" + ` ` + "\x1b[3mSee \x1b[0m" + `'` + "\x1b[96;21mfn <command> --help\x1b[0m" + `' ` + "\x1b[3mfor more information about a command.\x1b[0m" + `

    ` + "\x1b[1mLEARN MORE:\x1b[0m" + ` ` + "\x1b[91;4mhttps://github.com/fnproject/fn\x1b[0m" + ``

	// Override command template
	// SubcommandHelpTemplate is the text template for the subcommand help topic.
	// cli.go uses text/template to render templates. You can
	// render custom help text by setting this variable.
	cli.SubcommandHelpTemplate = `{{range .VisibleCategories}}{{if .Name}}
    ` + "\x1b[1m{{.Name}}:\x1b[0m" + `{{end}}{{end}}
        ` + "\x1b[36;1m{{ .HelpName}}\x1b[0m" + `{{if .Usage}} - ` + "\x1b[3m{{.Usage}}\x1b[0m" + `

    ` + "\x1b[1mUSAGE:\x1b[0m" + `
        ` + "\x1b[36;1m{{ .HelpName}}\x1b[0m" + ` {{if .VisibleFlags}} ` + "\x1b[36;21m[global options]\x1b[0m" + `{{end}} {{if .ArgsUsage}}` + "\x1b[91;21m{{.ArgsUsage}}\x1b[0m" + `{{end}} {{if .Flags}}` + "\x1b[33;21m[command options]\x1b[0m" + `{{end}} {{end}}{{if .Description}}
    
    ` + "\x1b[1mDESCRIPTION:\x1b[0m" + `
        {{.Description}}{{end}}{{if .Commands}}

    ` + "\x1b[1mSUBCOMMANDS:\x1b[0m" + ` {{range .Commands}}
        {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{if .VisibleFlags}}

    ` + "\x1b[1mCOMMAND OPTIONS:\x1b[0m" + ` {{range .VisibleFlags}}
        {{.}}{{end}}{{end}}{{if .Commands}}

    ` + "\x1b[1mFURTHER HELP:\x1b[0m" + ` ` + "\x1b[3mSee \x1b[0m" + `'` + "\x1b[96;21mfn <command> --help\x1b[0m" + `' ` + "\x1b[3mfor more information about a command.\x1b[0m" + `{{end}}
    `
	//Override command template
	// CommandHelpTemplate is the text template for the command help topic.
	// cli.go uses text/template to render templates. You can
	// render custom help text by setting this variable.
	cli.CommandHelpTemplate = `{{if .Category}}
    ` + "\x1b[1m{{.Category}}:\x1b[0m" + `{{end}}
    ` + "\x1b[36;1m{{.HelpName}}\x1b[0m" + `{{if .Usage}} - ` + "\x1b[3m{{.Usage}}\x1b[0m" + `
    
    ` + "\x1b[1mUSAGE:\x1b[0m" + `
    ` + "\x1b[36;1m{{.HelpName}}\x1b[0m" + ` ` + "\x1b[36;21m[global options]\x1b[0m" + ` {{if .ArgsUsage}}` + "\x1b[91;21m{{.ArgsUsage}}\x1b[0m" + `{{end}} {{if .Flags}}` + "\x1b[33;21m[command options]\x1b[0m" + `{{end}}{{end}}{{if .Description}}
    
    ` + "\x1b[1mDESCRIPTION:\x1b[0m" + `
        {{.Description}}{{end}}{{if .Subcommands}}

    ` + "\x1b[1mSUBCOMMANDS:\x1b[0m" + ` {{range .Subcommands}}
        {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{if .VisibleFlags}}
   
    ` + "\x1b[1mCOMMAND OPTIONS:\x1b[0m" + `
        {{range .Flags}}{{.}}
        {{end}}{{if .Subcommands}}
    
    ` + "\x1b[1mFURTHER HELP:\x1b[0m" + ` ` + "\x1b[3mSee \x1b[0m" + `'` + "\x1b[96;21mfn <command> --help\x1b[0m" + `' ` + "\x1b[3mfor more information about a command.\x1b[0m" + `{{end}}{{end}}
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
