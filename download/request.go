package download

import (
	"time"
)

type Request struct {
	Id            string
	Url           string
	ResourceKey   ResourceKey
	Checksum      string
	ChecksumType  string
	TimeRequested time.Time
	Callback      string
	DownloadId    string
	Errors        []*RequestError
	Metadata      *Metadata
}

func (r *Request) AddError(requestError error, errorTime time.Time) {
	r.Errors = append(r.Errors, NewRequestError(requestError, errorTime))
}
