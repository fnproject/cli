package objects

import (
	"github.com/fnproject/cli/objects/app"
	"github.com/fnproject/cli/objects/route"
	fnclient "github.com/fnproject/fn_go/client"
	"github.com/urfave/cli"
)

type FnClient struct {
	Client *fnclient.Fn
}

func getSubCommands(signature string, client *FnClient) []cli.Command {
	var subCommands []cli.Command
	subCommands = append(subCommands, app.getCommand(signature, client))
	subCommands = append(subCommands, route.routes(signature, client))
	subCommands = append(subCommands, contextCommand(signature))

	return subCommands
}
