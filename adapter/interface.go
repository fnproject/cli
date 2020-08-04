// Package adapter allows this CLI to switch between clients provided by fn_go and oci-go-sdk
// https://en.wikipedia.org/wiki/Adapter_pattern
package adapter

import (
	"fmt"
)

type Provider interface {
	APIClient() APIClient
	VersionClient() VersionClient
	FunctionInvokeClient() FunctionInvokeClient
}

type APIClient interface {
	FnClient() FnClient
	AppClient() AppClient
	TriggerClient() TriggerClient
}

type FnClient interface {
	CreateFn(fn *Fn) (*Fn, error)
	UpdateFn(fn *Fn) (*Fn, error)
	GetFn(appID string, fnName string) (*Fn, error)
	ListFn(appID string, limit int64) ([]*Fn, error)
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


type VersionClient interface {
	GetVersion() string
}

type FunctionInvokeClient interface {
	InvokeFunction(fn string)
}

// NameNotFoundError error for app not found when looked up by name
type AppNameNotFoundError struct {
	Name string
}

func (n AppNameNotFoundError) Error() string {
	return fmt.Sprintf("app %s not found", n.Name)
}

// NameNotFoundError error for function not found when looked up by name
type FunctionNameNotFoundError struct {
	Name string
}

func (n FunctionNameNotFoundError) Error() string {
	return fmt.Sprintf("function %s not found", n.Name)
}