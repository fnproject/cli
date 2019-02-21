// Code generated by go-swagger; DO NOT EDIT.

package triggers

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"

	"github.com/go-openapi/runtime"

	strfmt "github.com/go-openapi/strfmt"

	modelsv2 "github.com/fnproject/fn_go/modelsv2"
)

// DeleteTriggerReader is a Reader for the DeleteTrigger structure.
type DeleteTriggerReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *DeleteTriggerReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {

	case 204:
		result := NewDeleteTriggerNoContent()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	case 404:
		result := NewDeleteTriggerNotFound()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return nil, result

	default:
		result := NewDeleteTriggerDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewDeleteTriggerNoContent creates a DeleteTriggerNoContent with default headers values
func NewDeleteTriggerNoContent() *DeleteTriggerNoContent {
	return &DeleteTriggerNoContent{}
}

/*DeleteTriggerNoContent handles this case with default header values.

Trigger successfully deleted.
*/
type DeleteTriggerNoContent struct {
}

func (o *DeleteTriggerNoContent) Error() string {
	return fmt.Sprintf("[DELETE /triggers/{triggerID}][%d] deleteTriggerNoContent ", 204)
}

func (o *DeleteTriggerNoContent) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	return nil
}

// NewDeleteTriggerNotFound creates a DeleteTriggerNotFound with default headers values
func NewDeleteTriggerNotFound() *DeleteTriggerNotFound {
	return &DeleteTriggerNotFound{}
}

/*DeleteTriggerNotFound handles this case with default header values.

The Trigger does not exist.
*/
type DeleteTriggerNotFound struct {
	Payload *modelsv2.Error
}

func (o *DeleteTriggerNotFound) Error() string {
	return fmt.Sprintf("[DELETE /triggers/{triggerID}][%d] deleteTriggerNotFound  %+v", 404, o.Payload)
}

func (o *DeleteTriggerNotFound) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(modelsv2.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewDeleteTriggerDefault creates a DeleteTriggerDefault with default headers values
func NewDeleteTriggerDefault(code int) *DeleteTriggerDefault {
	return &DeleteTriggerDefault{
		_statusCode: code,
	}
}

/*DeleteTriggerDefault handles this case with default header values.

An unexpected error occurred.
*/
type DeleteTriggerDefault struct {
	_statusCode int

	Payload *modelsv2.Error
}

// Code gets the status code for the delete trigger default response
func (o *DeleteTriggerDefault) Code() int {
	return o._statusCode
}

func (o *DeleteTriggerDefault) Error() string {
	return fmt.Sprintf("[DELETE /triggers/{triggerID}][%d] DeleteTrigger default  %+v", o._statusCode, o.Payload)
}

func (o *DeleteTriggerDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(modelsv2.Error)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}
