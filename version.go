package main

import (
	"fmt"
	"os"

	client "github.com/fnproject/cli/client"
	versionclient "github.com/fnproject/fn_go/client/version"
	"github.com/urfave/cli"
)

// Version of Fn CLI
var Version = "0.4.11"

func version() cli.Command {
	r := versionCmd{client: versionclient.New(client.GetTransportAndRegistry())}
	return cli.Command{
		Name:   "version",
		Usage:  "displays fn and functions daemon versions",
		Action: r.version,
	}
}

type versionCmd struct {
	client *versionclient.Client
}

func (r *versionCmd) version(c *cli.Context) error {
	apiURL := os.Getenv("API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}
	fmt.Println("Client version:", Version)
	v, err := r.client.GetVersion(nil)
	if err != nil {
		return err
	}
	fmt.Println("Server version", v.Payload.Version)
	return nil
}
