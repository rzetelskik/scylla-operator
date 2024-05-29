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
	"github.com/go-openapi/swag"
)

// NewStorageServiceCompactionThroughputPostParams creates a new StorageServiceCompactionThroughputPostParams object
// with the default values initialized.
func NewStorageServiceCompactionThroughputPostParams() *StorageServiceCompactionThroughputPostParams {
	var ()
	return &StorageServiceCompactionThroughputPostParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewStorageServiceCompactionThroughputPostParamsWithTimeout creates a new StorageServiceCompactionThroughputPostParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewStorageServiceCompactionThroughputPostParamsWithTimeout(timeout time.Duration) *StorageServiceCompactionThroughputPostParams {
	var ()
	return &StorageServiceCompactionThroughputPostParams{

		timeout: timeout,
	}
}

// NewStorageServiceCompactionThroughputPostParamsWithContext creates a new StorageServiceCompactionThroughputPostParams object
// with the default values initialized, and the ability to set a context for a request
func NewStorageServiceCompactionThroughputPostParamsWithContext(ctx context.Context) *StorageServiceCompactionThroughputPostParams {
	var ()
	return &StorageServiceCompactionThroughputPostParams{

		Context: ctx,
	}
}

// NewStorageServiceCompactionThroughputPostParamsWithHTTPClient creates a new StorageServiceCompactionThroughputPostParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewStorageServiceCompactionThroughputPostParamsWithHTTPClient(client *http.Client) *StorageServiceCompactionThroughputPostParams {
	var ()
	return &StorageServiceCompactionThroughputPostParams{
		HTTPClient: client,
	}
}

/*
StorageServiceCompactionThroughputPostParams contains all the parameters to send to the API endpoint
for the storage service compaction throughput post operation typically these are written to a http.Request
*/
type StorageServiceCompactionThroughputPostParams struct {

	/*Value
	  compaction throughput

	*/
	Value int32

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the storage service compaction throughput post params
func (o *StorageServiceCompactionThroughputPostParams) WithTimeout(timeout time.Duration) *StorageServiceCompactionThroughputPostParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the storage service compaction throughput post params
func (o *StorageServiceCompactionThroughputPostParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the storage service compaction throughput post params
func (o *StorageServiceCompactionThroughputPostParams) WithContext(ctx context.Context) *StorageServiceCompactionThroughputPostParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the storage service compaction throughput post params
func (o *StorageServiceCompactionThroughputPostParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the storage service compaction throughput post params
func (o *StorageServiceCompactionThroughputPostParams) WithHTTPClient(client *http.Client) *StorageServiceCompactionThroughputPostParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the storage service compaction throughput post params
func (o *StorageServiceCompactionThroughputPostParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithValue adds the value to the storage service compaction throughput post params
func (o *StorageServiceCompactionThroughputPostParams) WithValue(value int32) *StorageServiceCompactionThroughputPostParams {
	o.SetValue(value)
	return o
}

// SetValue adds the value to the storage service compaction throughput post params
func (o *StorageServiceCompactionThroughputPostParams) SetValue(value int32) {
	o.Value = value
}

// WriteToRequest writes these params to a swagger request
func (o *StorageServiceCompactionThroughputPostParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// query param value
	qrValue := o.Value
	qValue := swag.FormatInt32(qrValue)
	if qValue != "" {
		if err := r.SetQueryParam("value", qValue); err != nil {
			return err
		}
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}