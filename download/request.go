package download

import (
	"time"
)

type Request struct {
	Id            string
	Url           string
	Checksum      string
	ChecksumType  string
	TimeRequested time.Time
	Callback      string
	DownloadId    string
	Errors        []*RequestError
	Metadata      *Metadata
}

func (r *Request) ResourceKey() ResourceKey {
	rk := ResourceKey{Url: r.Url}
	if r.Metadata != nil {
		rk.ETag = r.Metadata.ETag
	}
	return rk
}

func (r *Request) AddError(requestError error, errorTime time.Time) {
	r.Errors = append(r.Errors, NewRequestError(requestError, errorTime))
}
