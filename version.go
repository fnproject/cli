package main

import (
	"github.com/fnproject/cli/client"
	fnclient "github.com/fnproject/fn_go/client/version"
	"github.com/urfave/cli"
	"fmt"
)

// Version of Fn CLI
var Version = "0.4.11"

func version() cli.Command {
	return cli.Command{
		Name:   "version",
		Usage:  "displays fn and functions daemon versions",
		Action: versionCMD,
	}
}

func versionCMD(c *cli.Context) error {
	t, reg := client.GetTransportAndRegistry()
	// dirty hack, swagger paths live under /v1
	// version is also there, but it shouldn't
	// dropping base path to get appropriate URL for request eventually
	t.BasePath = ""
	version_client := fnclient.New(t, reg)
	v, err := version_client.GetVersion(nil)
	if err != nil {
		return err
	}
	fmt.Println("Client version: ", Version)
	fmt.Println("Server version: ", v.Payload.Version)
	return nil
}
