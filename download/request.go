package download

import (
	"time"
)

type Request struct {
	ID            string
	URL           string
	Checksum      string
	ChecksumType  string
	TimeRequested time.Time
	Callback      string
	DownloadID    string
	Errors        []*RequestError
	Metadata      *Metadata
}

func (r *Request) ResourceKey() ResourceKey {
	rk := ResourceKey{URL: r.URL}
	if r.Metadata != nil {
		rk.ETag = r.Metadata.ETag
	}
	return rk
}

func (r *Request) AddError(requestError error, errorTime time.Time) {
	r.Errors = append(r.Errors, NewRequestError(requestError, errorTime))
}
