// Package adapter allows this CLI to switch between clients provided by fn_go and oci-go-sdk
// https://en.wikipedia.org/wiki/Adapter_pattern
package adapter

import (
	"fmt"
	"github.com/go-openapi/strfmt"
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

// Copied from fn_go/modelsv2/app.go
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

// Copied from fn_go/modelsv2/fn.go
// We need to maintain this interface since we have already exposed it to the user
type Fn struct {
	// Func annotations - this is a map of annotations attached to this func, keys must not exceed 128 bytes and must consist of non-whitespace printable ascii characters, and the seralized representation of individual values must not exeed 512 bytes.
	Annotations map[string]interface{} `json:"annotations,omitempty"`

	// App ID.
	AppID string `json:"app_id,omitempty"`

	// Function configuration key values.
	Config map[string]string `json:"config,omitempty"`

	// Time when function was created. Always in UTC RFC3339.
	// Read Only: true
	// Format: date-time
	CreatedAt strfmt.DateTime `json:"created_at,omitempty"`

	// Unique identifier
	// Read Only: true
	ID string `json:"id,omitempty"`

	// Hot functions idle timeout before container termination. Value in Seconds.
	IDLETimeout *int32 `json:"idle_timeout,omitempty"`

	// Full container image name, e.g. hub.docker.com/fnproject/yo or fnproject/yo (default registry: hub.docker.com)
	Image string `json:"image,omitempty"`

	// Maximum usable memory given to function (MiB).
	Memory uint64 `json:"memory,omitempty"`

	// Unique name for this function.
	Name string `json:"name,omitempty"`

	// Timeout for executions of a function. Value in Seconds.
	Timeout *int32 `json:"timeout,omitempty"`

	// Most recent time that function was updated. Always in UTC RFC3339.
	// Read Only: true
	// Format: date-time
	UpdatedAt strfmt.DateTime `json:"updated_at,omitempty"`
}

// NameNotFoundError error for app not found when looked up by name
type NameNotFoundError struct {
	Name string
}

func (n NameNotFoundError) Error() string {
	return fmt.Sprintf("app %s not found", n.Name)
}