// Code generated by go-swagger; DO NOT EDIT.

package operations

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"fmt"
	"io"
	"strings"

	"github.com/go-openapi/runtime"
	"github.com/go-openapi/strfmt"

	"github.com/scylladb/scylla-manager/v3/swagger/gen/scylla/v1/models"
)

// MessagingServiceVersionGetReader is a Reader for the MessagingServiceVersionGet structure.
type MessagingServiceVersionGetReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *MessagingServiceVersionGetReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewMessagingServiceVersionGetOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil
	default:
		result := NewMessagingServiceVersionGetDefault(response.Code())
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		if response.Code()/100 == 2 {
			return result, nil
		}
		return nil, result
	}
}

// NewMessagingServiceVersionGetOK creates a MessagingServiceVersionGetOK with default headers values
func NewMessagingServiceVersionGetOK() *MessagingServiceVersionGetOK {
	return &MessagingServiceVersionGetOK{}
}

/*
MessagingServiceVersionGetOK handles this case with default header values.

Success
*/
type MessagingServiceVersionGetOK struct {
	Payload int32
}

func (o *MessagingServiceVersionGetOK) GetPayload() int32 {
	return o.Payload
}

func (o *MessagingServiceVersionGetOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	// response payload
	if err := consumer.Consume(response.Body(), &o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

// NewMessagingServiceVersionGetDefault creates a MessagingServiceVersionGetDefault with default headers values
func NewMessagingServiceVersionGetDefault(code int) *MessagingServiceVersionGetDefault {
	return &MessagingServiceVersionGetDefault{
		_statusCode: code,
	}
}

/*
MessagingServiceVersionGetDefault handles this case with default header values.

internal server error
*/
type MessagingServiceVersionGetDefault struct {
	_statusCode int

	Payload *models.ErrorModel
}

// Code gets the status code for the messaging service version get default response
func (o *MessagingServiceVersionGetDefault) Code() int {
	return o._statusCode
}

func (o *MessagingServiceVersionGetDefault) GetPayload() *models.ErrorModel {
	return o.Payload
}

func (o *MessagingServiceVersionGetDefault) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.ErrorModel)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

func (o *MessagingServiceVersionGetDefault) Error() string {
	return fmt.Sprintf("agent [HTTP %d] %s", o._statusCode, strings.TrimRight(o.Payload.Message, "."))
}