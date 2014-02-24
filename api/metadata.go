package api

import (
	"time"
)

type Metadata struct {
	MimeType      string    `json:"mime_type,omitempty"`
	Size          uint64    `json:"size,omitempty"`
	TimeRequested time.Time `json:"time_requested,omitempty"`

	// HTTP specific stuff
	Server       string    `json:"http_server,omitempty"`
	LastModified time.Time `json:"http_last_modified,omitempty"`
	ETag         string    `json:"http_etag,omitempty"`
	Expires      time.Time `json:"http_expires,omitempty"`
	StatusCode   int       `json:"http_status_code,omitempty"`
}
