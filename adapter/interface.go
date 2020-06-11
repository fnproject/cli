// Package adapter allows this CLI to switch between clients provided by fn_go and oci-go-sdk
// https://en.wikipedia.org/wiki/Adapter_pattern
package adapter

import (
	"fmt"
	"github.com/urfave/cli"
)

type ProviderAdapter interface {
	GetClientAdapter() ClientAdapter
}

type ClientAdapter interface {
	GetFnsClient() FnClient
	GetAppsClient() AppClient
	GetTriggersClient() TriggerClient
}

type FnClient interface {
	CreateFn(c *cli.Context) error
	UpdateFn(c *cli.Context) error
	GetFn(c *cli.Context) error
	ListFn(c *cli.Context) error
	DeleteFn(c *cli.Context) error
}

type AppClient interface {
	CreateApp(c *cli.Context) (*App, error)
	GetApp(c *cli.Context) (*App, error)
	UpdateApp(c *cli.Context) error
	ListApp(c *cli.Context) ([]*App, error)
	DeleteApp(c *cli.Context) error
}

type TriggerClient interface {
}

type App struct {
	Name string
	ID   string
}

type Fn struct {
}

// NameNotFoundError error for app not found when looked up by name
type NameNotFoundError struct {
	Name string
}

func (n NameNotFoundError) Error() string {
	return fmt.Sprintf("app %s not found", n.Name)
}