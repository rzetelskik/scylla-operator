// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// EndpointDetail endpoint_detail
//
// # Endpoint detail
//
// swagger:model endpoint_detail
type EndpointDetail struct {

	// The endpoint datacenter
	Datacenter string `json:"datacenter,omitempty"`

	// The endpoint host
	Host string `json:"host,omitempty"`

	// The endpoint rack
	Rack string `json:"rack,omitempty"`
}

// Validate validates this endpoint detail
func (m *EndpointDetail) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *EndpointDetail) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *EndpointDetail) UnmarshalBinary(b []byte) error {
	var res EndpointDetail
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
