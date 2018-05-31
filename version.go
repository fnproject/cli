package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/fnproject/cli/client"
	fnclient "github.com/fnproject/fn_go/client/version"
	"github.com/urfave/cli"
)

// Version of Fn CLI
var Version = "0.4.102"

func version() cli.Command {
	return cli.Command{
		Name:   "version",
		Usage:  "displays cli and server versions",
		Action: versionCMD,
	}
}

func versionCMD(c *cli.Context) error {
	t, reg, err := client.GetTransportAndRegistry()
	if err != nil {
		return err
	}
	// dirty hack, swagger paths live under /v1
	// version is also there, but it shouldn't
	// dropping base path to get appropriate URL for request eventually
	t.BasePath = ""

	ver := getLatestVersion()
	if ver == "" {
		ver = "Client version: " + Version
	}
	fmt.Println(ver)
	versionClient := fnclient.New(t, reg)
	v, err := versionClient.GetVersion(nil)
	if err != nil {
		fmt.Println("Server version: ", "?")
		return nil
	}
	fmt.Println("Server version: ", v.Payload.Version)
	return nil
}

func printLatestVersion() {
	v := getLatestVersion()
	if v != "" {
		fmt.Fprintln(os.Stderr, v)
	}
}

func getLatestVersion() string {
	base := "https://github.com/fnproject/cli/releases"
	url := ""
	c := http.Client{}
	c.Timeout = time.Second * 3
	c.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		url = req.URL.String()
		return nil
	}
	r, err := c.Get(fmt.Sprintf("%s/latest", base))
	if err != nil {
		return ""
	}
	defer r.Body.Close()
	if !strings.Contains(url, base) {
		return ""
	}
	if path.Base(url) != Version {
		return fmt.Sprintf("Client version: %s is not latest: %s", Version, path.Base(url))
	}
	return "Client version is latest version: " + path.Base(url)
}
