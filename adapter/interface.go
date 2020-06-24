// Package adapter allows this CLI to switch between clients provided by fn_go and oci-go-sdk
// https://en.wikipedia.org/wiki/Adapter_pattern
package adapter

import (
	"fmt"
)

type ProviderAdapter interface {
	APIClientAdapter() APIClientAdapter
	VersionClientAdapter() VersionClientAdapter
	FunctionInvokeClientAdapter() FunctionInvokeClientAdapter
}

type APIClientAdapter interface {
	FnClient() FnClient
	AppClient() AppClient
	TriggerClient() TriggerClient
}

type FnClient interface {
	CreateFn(fn *Fn) (*Fn, error)
	UpdateFn(fn *Fn) (*Fn, error)
	GetFn(appID string, fnName string) (*Fn, error)
	ListFn(appID string, limit int64) error
	DeleteFn(fnID string) error
}

type AppClient interface {
	CreateApp(app *App) (*App, error)
	GetApp(appName string) (*App, error)
	UpdateApp(app *App) (*App, error)
	ListApp(limit int64) ([]*App, error)
	DeleteApp(appID string) error
}

type TriggerClient interface {
}

type VersionClientAdapter interface {
	VersionClient() VersionClient
}

type VersionClient interface {
	GetVersion() string
}

type FunctionInvokeClientAdapter interface {
	FunctionInvokeClient() FunctionInvokeClient
}

type FunctionInvokeClient interface {
	InvokeFunction(fn string)
}

// NameNotFoundError error for app not found when looked up by name
type NameNotFoundError struct {
	Name string
}

func (n NameNotFoundError) Error() string {
	return fmt.Sprintf("app %s not found", n.Name)
}
