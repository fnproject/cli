// Code generated by go-swagger; DO NOT EDIT.

package fns

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/fnproject/fn_go/modelsv2"
)

// GetFnReader is a Reader for the GetFn structure.
type GetFnReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *GetFnReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewGetFnOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	case 404:
		result := NewGetFnNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result
	default:
		result := NewGetFnDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewGetFnOK creates a GetFnOK with default headers values
func NewGetFnOK() *GetFnOK {
	return &GetFnOK{}
}

/*GetFnOK handles this case with default header values.

Function definition
*/
type GetFnOK struct {
	Payload *modelsv2.Fn
}

func (o *GetFnOK) Error() string {
	return fmt.Sprintf("[GET /fns/{fnID}][%d] getFnOK  %+v", 200, o.Payload)
}

func (o *GetFnOK) GetPayload() *modelsv2.Fn {
	return o.Payload
}

func (o *GetFnOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(modelsv2.Fn)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetFnNotFound creates a GetFnNotFound with default headers values
func NewGetFnNotFound() *GetFnNotFound {
	return &GetFnNotFound{}
}

/*GetFnNotFound handles this case with default header values.

Function does not exist.
*/
type GetFnNotFound struct {
	Payload *modelsv2.Error
}

func (o *GetFnNotFound) Error() string {
	return fmt.Sprintf("[GET /fns/{fnID}][%d] getFnNotFound  %+v", 404, o.Payload)
}

func (o *GetFnNotFound) GetPayload() *modelsv2.Error {
	return o.Payload
}

func (o *GetFnNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(modelsv2.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewGetFnDefault creates a GetFnDefault with default headers values
func NewGetFnDefault(code int) *GetFnDefault {
	return &GetFnDefault{
		_statusCode: code,
	}
}

/*GetFnDefault handles this case with default header values.

Error
*/
type GetFnDefault struct {
	_statusCode int

	Payload *modelsv2.Error
}

// Code gets the status code for the get fn default response
func (o *GetFnDefault) Code() int {
	return o._statusCode
}

func (o *GetFnDefault) Error() string {
	return fmt.Sprintf("[GET /fns/{fnID}][%d] GetFn default  %+v", o._statusCode, o.Payload)
}

func (o *GetFnDefault) GetPayload() *modelsv2.Error {
	return o.Payload
}

func (o *GetFnDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(modelsv2.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
