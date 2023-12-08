// Code generated by go-swagger; DO NOT EDIT.

package operations

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

// NewColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams creates a new ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams object,
// with the default timeout for this client.
//
// Default values are not hydrated, since defaults are normally applied by the API server side.
//
// To enforce default values in parameter, use SetDefaults or WithDefaults.
func NewColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams() *ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams {
	return &ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams{
		timeout: cr.DefaultTimeout,
	}
}

// NewColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParamsWithTimeout creates a new ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams object
// with the ability to set a timeout on a request.
func NewColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParamsWithTimeout(timeout time.Duration) *ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams {
	return &ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams{
		timeout: timeout,
	}
}

// NewColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParamsWithContext creates a new ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams object
// with the ability to set a context for a request.
func NewColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParamsWithContext(ctx context.Context) *ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams {
	return &ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams{
		Context: ctx,
	}
}

// NewColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParamsWithHTTPClient creates a new ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams object
// with the ability to set a custom HTTPClient for a request.
func NewColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParamsWithHTTPClient(client *http.Client) *ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams {
	return &ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams{
		HTTPClient: client,
	}
}

/*
ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams contains all the parameters to send to the API endpoint

	for the column family metrics bloom filter off heap memory used by name get operation.

	Typically these are written to a http.Request.
*/
type ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams struct {

	/* Name.

	   The column family name in keyspace:name format
	*/
	Name string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithDefaults hydrates default values in the column family metrics bloom filter off heap memory used by name get params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams) WithDefaults() *ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams {
	o.SetDefaults()
	return o
}

// SetDefaults hydrates default values in the column family metrics bloom filter off heap memory used by name get params (not the query body).
//
// All values with no default are reset to their zero value.
func (o *ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams) SetDefaults() {
	// no default values defined for this parameter
}

// WithTimeout adds the timeout to the column family metrics bloom filter off heap memory used by name get params
func (o *ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams) WithTimeout(timeout time.Duration) *ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the column family metrics bloom filter off heap memory used by name get params
func (o *ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the column family metrics bloom filter off heap memory used by name get params
func (o *ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams) WithContext(ctx context.Context) *ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the column family metrics bloom filter off heap memory used by name get params
func (o *ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the column family metrics bloom filter off heap memory used by name get params
func (o *ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams) WithHTTPClient(client *http.Client) *ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the column family metrics bloom filter off heap memory used by name get params
func (o *ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithName adds the name to the column family metrics bloom filter off heap memory used by name get params
func (o *ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams) WithName(name string) *ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams {
	o.SetName(name)
	return o
}

// SetName adds the name to the column family metrics bloom filter off heap memory used by name get params
func (o *ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams) SetName(name string) {
	o.Name = name
}

// WriteToRequest writes these params to a swagger request
func (o *ColumnFamilyMetricsBloomFilterOffHeapMemoryUsedByNameGetParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param name
	if err := r.SetPathParam("name", o.Name); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}