package api

import (
	"github.com/patdowney/downloaderd/download"
)

type IncomingRequest struct {
	URL          string `json:"url"`
	Checksum     string `json:"checksum,omitempty"`
	ChecksumType string `json:"checksum_type,omitempty"`
	Callback     string `json:"callback,omitempty"`
}

func (air *IncomingRequest) ToDownloadRequest() *download.Request {
	downloadReq := &download.Request{
		URL:          air.URL,
		Checksum:     air.Checksum,
		ChecksumType: air.ChecksumType,
		Callback:     air.Callback,
		Errors:       make([]*download.RequestError, 0)}

	return downloadReq
}
