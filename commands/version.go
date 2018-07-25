package commands

import (
	"fmt"

	"github.com/fnproject/cli/client"
	"github.com/fnproject/cli/config"
	"github.com/urfave/cli"
)

// VersionCommand
func VersionCommand() cli.Command {
	return cli.Command{
		Name:        "version",
		Usage:       "Display Fn CLI and Fn Server versions",
		Description: "This command shows the version of the Fn CLI being used and the version of the Fn Server referenced by the current context, if available.",
		Action:      versionCMD,
	}
}

func versionCMD(c *cli.Context) error {
	provider, err := client.CurrentProvider()
	if err != nil {
		return err
	}

	ver := config.GetVersion("latest")
	if ver == "" {
		ver = "Client version: " + config.Version
	}
	fmt.Println(ver)

	versionClient := provider.VersionClient()
	v, err := versionClient.GetVersion(nil)
	if err != nil {
		fmt.Println("Server version: ", "?")
		return nil
	}
	fmt.Println("Server version: ", v.Payload.Version)
	return nil
}
