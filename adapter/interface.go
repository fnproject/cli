// Package adapter allows this CLI to switch between clients provided by fn_go and oci-go-sdk
// https://en.wikipedia.org/wiki/Adapter_pattern
package adapter

import (
	"fmt"
	"github.com/go-openapi/strfmt"
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

// Copied from fnproject/cli/vendor/github.com/fnproject/fn_go/modelsv2/app.go
// We need to maintain this interface since we have already exposed it to the user
type App struct {
	// Application annotations - this is a map of annotations attached to this app, keys must not exceed 128 bytes and must consist of non-whitespace printable ascii characters, and the seralized representation of individual values must not exeed 512 bytes.
	Annotations map[string]interface{} `json:"annotations,omitempty"`

	// Application function configuration, applied to all Functions.
	Config map[string]string `json:"config,omitempty"`

	// Time when app was created. Always in UTC.
	// Read Only: true
	// Format: date-time
	CreatedAt strfmt.DateTime `json:"created_at,omitempty"`

	// App ID
	// Read Only: true
	ID string `json:"id,omitempty"`

	// Name of this app. Must be different than the image name. Can ony contain alphanumeric, -, and _.
	// Read Only: true
	Name string `json:"name,omitempty"`

	// A comma separated list of syslog urls to send all function logs to. supports tls, udp or tcp. e.g. tls://logs.papertrailapp.com:1
	SyslogURL *string `json:"syslog_url,omitempty"`

	// Most recent time that app was updated. Always in UTC.
	// Read Only: true
	// Format: date-time
	UpdatedAt strfmt.DateTime `json:"updated_at,omitempty"`
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