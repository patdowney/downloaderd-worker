package api

import (
	"github.com/patdowney/downloaderd/download"
	"time"
)

type Request struct {
	Id                   string    `json:"id"`
	Url                  string    `json:"url"`
	ExpectedChecksum     string    `json:"expected_checksum,omitempty"`
	ExpectedChecksumType string    `json:"expected_checksum_type,omitempty"`
	TimeRequested        time.Time `json:"time_requested"`
	Callback             string    `json:"callback,omitempty"`
	Download             *Download `json:"download,omitempty"`
	Errors               []*Error  `json:"errors,omitempty"`
	Metadata             *Metadata `json:"metadata,omitempty"`
	Link                 []Link    `json:"links"`
}

func NewRequestList(origList *[]*download.Request) *[]*Request {
	rs := make([]*Request, len(*origList))

	for i, r := range *origList {
		rs[i] = NewRequest(r)
	}

	return &rs
}

func NewRequest(orig *download.Request) *Request {
	apiRequest := &Request{
		Id:                   orig.Id,
		Url:                  orig.Url,
		ExpectedChecksum:     orig.Checksum,
		ExpectedChecksumType: orig.ChecksumType,
		TimeRequested:        orig.TimeRequested,
		Callback:             orig.Callback,
		Errors:               make([]*Error, len(orig.Errors)),
		Link:                 make([]Link, 0)}

	if orig.Metadata != nil {
		apiRequest.Metadata = NewMetadata(orig.Metadata)
	}

	if orig.Download != nil {
		apiRequest.Download = NewDownload(orig.Download)
	}

	if len(orig.Errors) > 0 {
		for i, e := range orig.Errors {
			apiRequest.Errors[i] = NewError(e.OriginalError)
		}
	}

	// add links here

	return apiRequest
}
