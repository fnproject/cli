// Code generated by go-swagger; DO NOT EDIT.

package modelsv2

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"encoding/json"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
	"github.com/go-openapi/validate"
)

// Fn fn
//
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
	IdleTimeout *int32 `json:"idle_timeout,omitempty"`

	// Full container image name, e.g. hub.docker.com/fnproject/yo or fnproject/yo (default registry: hub.docker.com)
	Image string `json:"image,omitempty"`

	// Maximum usable memory given to function (MiB).
	Memory uint64 `json:"memory,omitempty"`

	// Unique name for this function.
	Name string `json:"name,omitempty"`

	// Valid values are GENERIC_X86, GENERIC_ARM and GENERIC_X86_ARM. Default is GENERIC_X86. Setting this to GENERIC_X86, will run the functions in the application on X86 processor architecture.
	// Setting this to GENERIC_ARM, will run the functions in the application on ARM processor architecture.
	// When set to 'GENERIC_X86_ARM', functions in the application are run on either X86 or ARM processor architecture.
	// Accepted values are:
	// GENERIC_X86, GENERIC_ARM, GENERIC_X86_ARM
	//
	// Enum: [GENERIC_X86 GENERIC_ARM GENERIC_X86_ARM]
	Shape string `json:"shape,omitempty"`

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

	if err := m.validateShape(formats); err != nil {
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

var fnTypeShapePropEnum []interface{}

func init() {
	var res []string
	if err := json.Unmarshal([]byte(`["GENERIC_X86","GENERIC_ARM","GENERIC_X86_ARM"]`), &res); err != nil {
		panic(err)
	}
	for _, v := range res {
		fnTypeShapePropEnum = append(fnTypeShapePropEnum, v)
	}
}

const (

	// FnShapeGENERICX86 captures enum value "GENERIC_X86"
	FnShapeGENERICX86 string = "GENERIC_X86"

	// FnShapeGENERICARM captures enum value "GENERIC_ARM"
	FnShapeGENERICARM string = "GENERIC_ARM"

	// FnShapeGENERICX86ARM captures enum value "GENERIC_X86_ARM"
	FnShapeGENERICX86ARM string = "GENERIC_X86_ARM"
)

// prop value enum
func (m *Fn) validateShapeEnum(path, location string, value string) error {
	if err := validate.EnumCase(path, location, value, fnTypeShapePropEnum, true); err != nil {
		return err
	}
	return nil
}

func (m *Fn) validateShape(formats strfmt.Registry) error {

	if swag.IsZero(m.Shape) { // not required
		return nil
	}

	// value enum
	if err := m.validateShapeEnum("shape", "body", m.Shape); err != nil {
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
