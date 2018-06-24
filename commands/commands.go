package commands

import (
	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/objects/app"
	"github.com/fnproject/cli/objects/call"
	"github.com/fnproject/cli/objects/context"
	"github.com/fnproject/cli/objects/log"
	"github.com/fnproject/cli/objects/route"
	"github.com/fnproject/cli/objects/server"
	"github.com/fnproject/cli/run"
	"github.com/urfave/cli"
	"github.com/fnproject/cli/objects/fn"
)

type cmd map[string]cli.Command

// Commands map of all top-level commands
var Commands = cmd{
	"build":        BuildCommand(),
	"build-server": BuildServerCommand(),
	"bump":         common.BumpCommand(),
	"call":         CallCommand(),
	"configure":    ConfigureCommand(),
	"create":       CreateCommand(),
	"delete":       DeleteCommand(),
	"deploy":       DeployCommand(),
	"get":          GetCommand(),
	"init":         InitCommand(),
	"inspect":      InspectCommand(),
	"list":         ListCommand(),
	"push":         PushCommand(),
	"run":          run.RunCommand(),
	"start":        StartCommand(),
	"stop":         StopCommand(),
	"test":         TestCommand(),
	"unset":        UnsetCommand(),
	"update":       UpdateCommand(),
	"use":          UseCommand(),
}

var CreateCmds = cmd{
	"apps":      app.Create(),
	"routes":    route.Create(),
	"functions": fn.Create(),
	"context":   context.Create(),
}

var ConfigCmds = cmd{
	"apps":      app.SetConfig(),
	"functions": fn.SetConfig(),
	"routes":    route.SetConfig(),
}

var ConfigListCmds = cmd{
	"apps":      app.ListConfig(),
	"functions": fn.ListConfig(),
	"routes":    route.ListConfig(),
}

var ConfigGetCmds = cmd{
	"apps":      app.GetConfig(),
	"functions": fn.GetConfig(),
	"routes":    route.GetConfig(),
}

var ConfigSetCmds = cmd{
	"apps":      app.SetConfig(),
	"functions": fn.SetConfig(),
	"routes":    route.SetConfig(),
}

var ConfigUnsetCmds = cmd{
	"apps":      app.UnsetConfig(),
	"functions": fn.UnsetConfig(),
	"routes":    route.UnsetConfig(),
}

var DeleteCmds = cmd{
	"apps":      app.Delete(),
	"routes":    route.Delete(),
	"functions": fn.Delete(),
	"context":   context.Delete(),
}

var GetCmds = cmd{
	"config": ConfigCommand("get"),
	"logs":   log.Get(),
	"calls":  call.Get(),
}

var InspectCmds = cmd{
	"apps":      app.Inspect(),
	"functions": fn.Inspect(),
	"routes":    route.Inspect(),
}

var ListCmds = cmd{
	"config":    ConfigCommand("list"),
	"apps":      app.List(),
	"functions": fn.List(),
	"routes":    route.List(),
	"calls":     call.List(),
	"context":   context.List(),
}

var UnsetCmds = cmd{
	"config":  ConfigCommand("unset"),
	"context": context.Unset(),
}

var UpdateCmds = cmd{
	"apps":      app.Update(),
	"routes":    route.Update(),
	"functions": fn.Update(),
	"context":   context.Update(),
	"server":    server.Update(),
}

var UseCmds = cmd{
	"context": context.Use(),
}

// GetCommands returns a list of cli.commands
func GetCommands(commands map[string]cli.Command) []cli.Command {
	cmds := []cli.Command{}
	for _, cmd := range commands {
		cmds = append(cmds, cmd)
	}
	return cmds
}
