package main

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"time"

	"github.com/fnproject/cli/client"
	"github.com/urfave/cli"
)

// Version of Fn CLI
var Version = "0.4.127"

// VersionCommand
func VersionCommand() cli.Command {
	return cli.Command{
		Name:        "version",
		Usage:       "Display CLI and server versions",
		Description: "This is commands shows the latest client and server version.",
		Action:      versionCMD,
	}
}

func versionCMD(c *cli.Context) error {
	provider, err := client.CurrentProvider()
	if err != nil {
		return err
	}

	ver := getLatestVersion()
	if ver == "" {
		ver = "Client version: " + Version
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

// PrintLatestVersion to terminal
func PrintLatestVersion() {
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
