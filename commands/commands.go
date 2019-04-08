package commands

import (
	"sort"

	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/objects/app"
	"github.com/fnproject/cli/objects/context"
	"github.com/fnproject/cli/objects/fn"
	"github.com/fnproject/cli/objects/server"
	"github.com/fnproject/cli/objects/trigger"
	"github.com/urfave/cli"
)

//Cmd is a mapping from a commands name to its corresponding structure
type Cmd map[string]cli.Command

// Commands map of all top-level commands
var Commands = Cmd{
	"build":        BuildCommand(),
	"build-server": BuildServerCommand(),
	"bump":         common.BumpCommand(),
	"invoke":       InvokeCommand(),
	"configure":    ConfigureCommand(),
	"create":       CreateCommand(),
	"delete":       DeleteCommand(),
	"deploy":       DeployCommand(),
	"get":          GetCommand(),
	"init":         InitCommand(),
	"inspect":      InspectCommand(),
	"list":         ListCommand(),
	"migrate":      MigrateCommand(),
	"push":         PushCommand(),
	"start":        StartCommand(),
	"stop":         StopCommand(),
	"unset":        UnsetCommand(),
	"update":       UpdateCommand(),
	"use":          UseCommand(),
}

var CreateCmds = Cmd{
	"apps":      app.Create(),
	"functions": fn.Create(),
	"triggers":  trigger.Create(),
	"context":   context.Create(),
}

var ConfigCmds = Cmd{
	"apps":      app.SetConfig(),
	"functions": fn.SetConfig(),
}

var ConfigListCmds = Cmd{
	"apps":      app.ListConfig(),
	"functions": fn.ListConfig(),
}

var ConfigGetCmds = Cmd{
	"apps":      app.GetConfig(),
	"functions": fn.GetConfig(),
}

var ConfigSetCmds = Cmd{
	"apps":      app.SetConfig(),
	"functions": fn.SetConfig(),
}

var ConfigUnsetCmds = Cmd{
	"apps":      app.UnsetConfig(),
	"functions": fn.UnsetConfig(),
}

var DeleteCmds = Cmd{
	"apps":      app.Delete(),
	"functions": fn.Delete(),
	"context":   context.Delete(),
	"triggers":  trigger.Delete(),
	"config":    ConfigCommand("delete"),
}

var GetCmds = Cmd{
	"config": ConfigCommand("get"),
}

var InspectCmds = Cmd{
	"apps":      app.Inspect(),
	"context":   context.Inspect(),
	"functions": fn.Inspect(),
	"triggers":  trigger.Inspect(),
}

var ListCmds = Cmd{
	"config":    ConfigCommand("list"),
	"apps":      app.List(),
	"functions": fn.List(),
	"triggers":  trigger.List(),
	"contexts":  context.List(),
}

var UnsetCmds = Cmd{
	"config":  ConfigCommand("unset"),
	"context": context.Unset(),
}

var UpdateCmds = Cmd{
	"apps":      app.Update(),
	"functions": fn.Update(),
	"context":   context.Update(),
	"server":    server.Update(),
	"trigger":   trigger.Update(),
}

var UseCmds = Cmd{
	"context": context.Use(),
}

// GetCommands returns a list of cli.commands
func GetCommands(commands map[string]cli.Command) []cli.Command {
	cmds := []cli.Command{}
	for _, cmd := range commands {
		cmds = append(cmds, cmd)
	}

	sort.Sort(cli.CommandsByName(cmds))
	return cmds
}
