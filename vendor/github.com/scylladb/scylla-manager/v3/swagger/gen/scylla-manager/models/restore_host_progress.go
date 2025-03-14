// Code generated by go-swagger; DO NOT EDIT.

package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"github.com/go-openapi/strfmt"
	"github.com/go-openapi/swag"
)

// RestoreHostProgress restore host progress
//
// swagger:model RestoreHostProgress
type RestoreHostProgress struct {

	// Total time spent by host on download in milliseconds (included in restore_duration)
	DownloadDuration int64 `json:"download_duration,omitempty"`

	// Total bytes downloaded by host (included in restored_bytes)
	DownloadedBytes int64 `json:"downloaded_bytes,omitempty"`

	// host
	Host string `json:"host,omitempty"`

	// Total time spent by host on restore in milliseconds
	RestoreDuration int64 `json:"restore_duration,omitempty"`

	// Total bytes restored by host
	RestoredBytes int64 `json:"restored_bytes,omitempty"`

	// Host shard count
	ShardCnt int64 `json:"shard_cnt,omitempty"`

	// Total time spent by host on load&stream in milliseconds (included in restore_duration)
	StreamDuration int64 `json:"stream_duration,omitempty"`

	// Total bytes load&streamed by host (included in restored_bytes)
	StreamedBytes int64 `json:"streamed_bytes,omitempty"`
}

// Validate validates this restore host progress
func (m *RestoreHostProgress) Validate(formats strfmt.Registry) error {
	return nil
}

// MarshalBinary interface implementation
func (m *RestoreHostProgress) MarshalBinary() ([]byte, error) {
	if m == nil {
		return nil, nil
	}
	return swag.WriteJSON(m)
}

// UnmarshalBinary interface implementation
func (m *RestoreHostProgress) UnmarshalBinary(b []byte) error {
	var res RestoreHostProgress
	if err := swag.ReadJSON(b, &res); err != nil {
		return err
	}
	*m = res
	return nil
}
