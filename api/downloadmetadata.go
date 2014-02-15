package api

import (
	"github.com/patdowney/downloaderd/download"
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

func NewMetadata(dm *download.Metadata) *Metadata {
	m := &Metadata{
		TimeRequested: dm.TimeRequested,
		MimeType:      dm.MimeType,
		Size:          dm.Size,
		Server:        dm.Server,
		LastModified:  dm.LastModified,
		ETag:          dm.ETag,
		Expires:       dm.Expires,
		StatusCode:    dm.StatusCode}

	return m
}
