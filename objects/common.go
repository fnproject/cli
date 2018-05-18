package objects

import (
	"github.com/fnproject/cli/app"
	fnclient "github.com/fnproject/fn_go/client"
	"github.com/urfave/cli"
)

type FnClient struct {
	Client *fnclient.Fn
}

func (client *FnClient) getSubCommands(signature string) []cli.Command {
	var subCommands []cli.Command
	subCommands = append(subCommands, app.getCommand(signature, client))
	subCommands = append(subCommands, fnClient.routes(signature))
	subCommands = append(subCommands, contextCommand(signature))

	return subCommands
}
