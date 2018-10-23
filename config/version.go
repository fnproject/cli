package config

import (
	"fmt"
	"net/http"
	"path"
	"strings"
	"time"
)

// Version of Fn CLI
var Version = "0.5.23"

func GetVersion(versionType string) string {
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
	if versionType == "current" {
		return GetCurrentVersion(url)
	}

	return GetLatestVersion(url)
}

func GetLatestVersion(url string) string {
	if path.Base(url) != Version {
		return fmt.Sprintf("Client version: %s is not latest: %s", Version, path.Base(url))
	}
	return "Client version is latest version: " + path.Base(url)
}

func GetCurrentVersion(url string) string {
	return path.Base(url)
}
