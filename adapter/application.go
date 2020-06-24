package adapter

import "github.com/go-openapi/strfmt"

// Copied from fn_go/modelsv2/app.go
// We need to maintain this struct since we have already exposed it to the user
type App struct {
	// Application annotations - this is a map of annotations attached to this app, keys must not exceed 128 bytes and must consist of non-whitespace printable ascii characters, and the serialized representation of individual values must not exceed 512 bytes.
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

