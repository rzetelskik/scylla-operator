// Code generated by go-swagger; DO NOT EDIT.

package config

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"context"
	"net/http"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"
)

// NewFindConfigListenAddressParams creates a new FindConfigListenAddressParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewFindConfigListenAddressParams() *FindConfigListenAddressParams {
	return &FindConfigListenAddressParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewFindConfigListenAddressParamsWithTimeout creates a new FindConfigListenAddressParams object
// with the ability to set a timeout on a request.
func NewFindConfigListenAddressParamsWithTimeout(timeout time.Duration) *FindConfigListenAddressParams {
	return &FindConfigListenAddressParams{
		timeout: timeout,
	}
}

// NewFindConfigListenAddressParamsWithContext creates a new FindConfigListenAddressParams object
// with the ability to set a context for a request.
func NewFindConfigListenAddressParamsWithContext(ctx context.Context) *FindConfigListenAddressParams {
	return &FindConfigListenAddressParams{
		Context: ctx,
	}
}

// NewFindConfigListenAddressParamsWithHTTPClient creates a new FindConfigListenAddressParams object
// with the ability to set a custom HTTPClient for a request.
func NewFindConfigListenAddressParamsWithHTTPClient(client *http.Client) *FindConfigListenAddressParams {
	return &FindConfigListenAddressParams{
		HTTPClient: client,
	}
}

/*
FindConfigListenAddressParams contains all the parameters to send to the API endpoint

	for the find config listen address operation.

	Typically these are written to a http.Request.
*/
type FindConfigListenAddressParams struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the find config listen address params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *FindConfigListenAddressParams) WithDefaults() *FindConfigListenAddressParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the find config listen address params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *FindConfigListenAddressParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the find config listen address params
func (o *FindConfigListenAddressParams) WithTimeout(timeout time.Duration) *FindConfigListenAddressParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the find config listen address params
func (o *FindConfigListenAddressParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the find config listen address params
func (o *FindConfigListenAddressParams) WithContext(ctx context.Context) *FindConfigListenAddressParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the find config listen address params
func (o *FindConfigListenAddressParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the find config listen address params
func (o *FindConfigListenAddressParams) WithHTTPClient(client *http.Client) *FindConfigListenAddressParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the find config listen address params
func (o *FindConfigListenAddressParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *FindConfigListenAddressParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}