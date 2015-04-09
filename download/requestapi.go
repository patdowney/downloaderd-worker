package download

import (
	"github.com/patdowney/downloaderd-worker/api"
)

// FromAPIIncomingDownload ...
func FromAPIIncomingDownload(air *api.IncomingDownload) *Request {
	downloadReq := &Request{
		ID:           air.RequestID,
		URL:          air.URL,
		Checksum:     air.Checksum,
		ChecksumType: air.ChecksumType,
		Callback:     air.Callback,
		ETag:         air.ETag,
	}

	return downloadReq
}
