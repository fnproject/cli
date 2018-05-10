package main

import (
	"fmt"

	"github.com/fnproject/cli/client"
	fnclient "github.com/fnproject/fn_go/client/version"
	"github.com/urfave/cli"
)

// Version of Fn CLI
var Version = "0.4.88"

func version() cli.Command {
	return cli.Command{
		Name:   "version",
		Usage:  "displays cli and server versions",
		Action: versionCMD,
	}
}

func versionCMD(c *cli.Context) error {
	t, reg := client.GetTransportAndRegistry()
	// dirty hack, swagger paths live under /v1
	// version is also there, but it shouldn't
	// dropping base path to get appropriate URL for request eventually
	t.BasePath = ""
	fmt.Println("Client version: ", Version)
	versionClient := fnclient.New(t, reg)
	v, err := versionClient.GetVersion(nil)
	if err != nil {
		fmt.Println("Server version: ", "?")
		return nil
	}
	fmt.Println("Server version: ", v.Payload.Version)
	return nil
}
