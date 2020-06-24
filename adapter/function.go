package adapter

import "github.com/go-openapi/strfmt"

// Copied from fn_go/modelsv2/fn.go
// We need to maintain this struct since we have already exposed it to the user
type Fn struct {
	// Func annotations - this is a map of annotations attached to this func, keys must not exceed 128 bytes and must consist of non-whitespace printable ascii characters, and the serialized representation of individual values must not exceed 512 bytes.
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