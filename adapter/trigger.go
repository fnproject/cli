package adapter

import "github.com/go-openapi/strfmt"

// Copied from fn_go/modelsv2/trigger.go
// We need to maintain this struct since we have already exposed it to the user
type Trigger struct {

	// Trigger annotations - this is a map of annotations attached to this trigger, keys must not exceed 128 bytes and must consist of non-whitespace printable ascii characters, and the seralized representation of individual values must not exeed 512 bytes.
	Annotations map[string]interface{} `json:"annotations,omitempty"`

	// Opaque, unique Application identifier
	// Read Only: true
	AppID string `json:"app_id,omitempty"`

	// Time when trigger was created. Always in UTC.
	// Read Only: true
	// Format: date-time
	CreatedAt strfmt.DateTime `json:"created_at,omitempty"`

	// Opaque, unique Function identifier
	// Read Only: true
	FnID string `json:"fn_id,omitempty"`

	// Unique Trigger identifier.
	// Read Only: true
	ID string `json:"id,omitempty"`

	// Unique name for this trigger, used to identify this trigger.
	Name string `json:"name,omitempty"`

	// URI path for this trigger. e.g. `sayHello`, `say/hello`
	Source string `json:"source,omitempty"`

	// Class of trigger, e.g. schedule, http, queue
	Type string `json:"type,omitempty"`

	// Most recent time that trigger was updated. Always in UTC.
	// Read Only: true
	// Format: date-time
	UpdatedAt strfmt.DateTime `json:"updated_at,omitempty"`
}
