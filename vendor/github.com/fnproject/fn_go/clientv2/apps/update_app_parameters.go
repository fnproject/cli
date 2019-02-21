// Code generated by go-swagger; DO NOT EDIT.

package apps

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"

	strfmt "github.com/go-openapi/strfmt"

	modelsv2 "github.com/fnproject/fn_go/modelsv2"
)

// NewUpdateAppParams creates a new UpdateAppParams object
// with the default values initialized.
func NewUpdateAppParams() *UpdateAppParams {
	var ()
	return &UpdateAppParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewUpdateAppParamsWithTimeout creates a new UpdateAppParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewUpdateAppParamsWithTimeout(timeout time.Duration) *UpdateAppParams {
	var ()
	return &UpdateAppParams{

		timeout: timeout,
	}
}

// NewUpdateAppParamsWithContext creates a new UpdateAppParams object
// with the default values initialized, and the ability to set a context for a request
func NewUpdateAppParamsWithContext(ctx context.Context) *UpdateAppParams {
	var ()
	return &UpdateAppParams{

		Context: ctx,
	}
}

// NewUpdateAppParamsWithHTTPClient creates a new UpdateAppParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewUpdateAppParamsWithHTTPClient(client *http.Client) *UpdateAppParams {
	var ()
	return &UpdateAppParams{
		HTTPClient: client,
	}
}

/*UpdateAppParams contains all the parameters to send to the API endpoint
for the update app operation typically these are written to a http.Request
*/
type UpdateAppParams struct {

	/*AppID
	  Opaque, unique Application ID.

	*/
	AppID string
	/*Body
	  Application data to merge with current values.

	*/
	Body *modelsv2.App

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the update app params
func (o *UpdateAppParams) WithTimeout(timeout time.Duration) *UpdateAppParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the update app params
func (o *UpdateAppParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the update app params
func (o *UpdateAppParams) WithContext(ctx context.Context) *UpdateAppParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the update app params
func (o *UpdateAppParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the update app params
func (o *UpdateAppParams) WithHTTPClient(client *http.Client) *UpdateAppParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the update app params
func (o *UpdateAppParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithAppID adds the appID to the update app params
func (o *UpdateAppParams) WithAppID(appID string) *UpdateAppParams {
	o.SetAppID(appID)
	return o
}

// SetAppID adds the appId to the update app params
func (o *UpdateAppParams) SetAppID(appID string) {
	o.AppID = appID
}

// WithBody adds the body to the update app params
func (o *UpdateAppParams) WithBody(body *modelsv2.App) *UpdateAppParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the update app params
func (o *UpdateAppParams) SetBody(body *modelsv2.App) {
	o.Body = body
}

// WriteToRequest writes these params to a swagger request
func (o *UpdateAppParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param appID
	if err := r.SetPathParam("appID", o.AppID); err != nil {
		return err
	}

	if o.Body != nil {
		if err := r.SetBodyParam(o.Body); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
