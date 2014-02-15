package api

import (
	"github.com/patdowney/downloaderd/download"
	"time"
)

type Metadata struct {
	MimeType string `json:"mime_type,omitempty"`
	Size     uint64 `json:"size,omitempty"`

	// HTTP specific stuff
	Server       string    `json:"http_server,omitempty"`
	LastModified time.Time `json:"http_last_modified,omitempty"`
	ETag         string    `json:"http_etag,omitempty"`
	Expires      time.Time `json:"http_expires,omitempty"`
	StatusCode   int       `json:"http_status_code,omitempty"`
}

func (m *Metadata) ToDownloadMetadata() *download.Metadata {
	dm := &download.Metadata{
		MimeType:     m.MimeType,
		Size:         m.Size,
		Server:       m.Server,
		LastModified: m.LastModified,
		ETag:         m.ETag,
		Expires:      m.Expires,
		StatusCode:   m.StatusCode}
	return dm
}

func NewMetadata(dm *download.Metadata) *Metadata {
	m := &Metadata{
		MimeType:     dm.MimeType,
		Size:         dm.Size,
		Server:       dm.Server,
		LastModified: dm.LastModified,
		ETag:         dm.ETag,
		Expires:      dm.Expires,
		StatusCode:   dm.StatusCode}

	return m
}
