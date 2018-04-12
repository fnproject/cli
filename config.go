package main

import "os"

const (
	rootConfigPathName     = ".fn"
	contextsPathName       = "contexts"
	configName             = "config"
	contextConfigFileName  = "config.yaml"
	defaultContextFileName = "default.yaml"

	readWritePerms = os.FileMode(0755)

	currentContext  = "current-context"
	contextProvider = "provider"

	envFnRegistry = "registry"
	envFnToken    = "token"
	envFnAPIURL   = "api_url"
	envFnContext  = "context"
)

var defaultRootConfigContents = map[string]string{currentContext: "default"}
var defaultContextConfigContents = map[string]string{
	contextProvider: "default",
	envFnAPIURL:     "https://localhost:8080",
	envFnRegistry:   "",
}
