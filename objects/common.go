package objects

import (
	"github.com/fnproject/cli/common"
	"github.com/fnproject/cli/objects/app"
	"github.com/fnproject/cli/objects/route"
	"github.com/urfave/cli"
)

func GetSubCommands(signature string, client *common.FnClient) []cli.Command {
	var subCommands []cli.Command
	subCommands = append(subCommands, app.GetCommand(signature, client))
	subCommands = append(subCommands, route.GetCommand(signature, client))
	subCommands = append(subCommands, contextCommand(signature))

	return subCommands
}
