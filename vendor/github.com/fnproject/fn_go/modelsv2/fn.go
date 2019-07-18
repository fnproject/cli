// Code generated by go-swagger; DO NOT EDIT.

package modelsv2

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// Fn fn
// swagger:model Fn
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

// Validate validates this fn
func (m *Fn) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateCreatedAt(formats); err != nil {
		res = append(res, err)
	}

	if err := m.validateUpdatedAt(formats); err != nil {
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *Fn) validateCreatedAt(formats strfmt.Registry) error {

	if swag.IsZero(m.CreatedAt) { // not required
		return nil
	}

	if err := validate.FormatOf("created_at", "body", "date-time", m.CreatedAt.String(), formats); err != nil {
		return err
	}

	return nil
}

func (m *Fn) validateUpdatedAt(formats strfmt.Registry) error {

	if swag.IsZero(m.UpdatedAt) { // not required
		return nil
	}

	if err := validate.FormatOf("updated_at", "body", "date-time", m.UpdatedAt.String(), formats); err != nil {
		return err
	}

	return nil
}

// MarshalBinary interface implementation
func (m *Fn) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *Fn) UnmarshalBinary(b []byte) error {
	var res Fn
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
