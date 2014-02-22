package api

import (
	"github.com/patdowney/downloaderd/download"
	"net/http"
	"time"
)

type Request struct {
	Id                   string    `json:"id"`
	Url                  string    `json:"url"`
	ExpectedChecksum     string    `json:"expected_checksum,omitempty"`
	ExpectedChecksumType string    `json:"expected_checksum_type,omitempty"`
	TimeRequested        time.Time `json:"time_requested"`
	Callback             string    `json:"callback,omitempty"`
	DownloadId           string    `json:"download_id,omitempty"`
	Errors               []*Error  `json:"errors,omitempty"`
	Metadata             *Metadata `json:"metadata,omitempty"`
	Links                []Link    `json:"links"`
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
		DownloadId:           orig.DownloadId,
		TimeRequested:        orig.TimeRequested,
		Callback:             orig.Callback,
		Errors:               make([]*Error, 0, len(orig.Errors)),
		Links:                make([]Link, 0)}

	if orig.Metadata != nil {
		apiRequest.Metadata = NewMetadata(orig.Metadata)
	}

	if len(orig.Errors) > 0 {
		for _, e := range orig.Errors {
			if e.OriginalError != "" {
				apiRequest.Errors = append(apiRequest.Errors, NewError(&e.ErrorWrapper))
			}
		}
	}

	// add links here
	apiRequest.Links = append(apiRequest.Links,
		Link{Relation: "self", Value: apiRequest.Id,
			ValueId: "id", RouteName: "request"})

	/*
		if apiRequest.Callback != "" {
			apiRequest.Links = append(apiRequest.Links,
				Link{Relation: "callback-status", Value: apiRequest.Callback, ValueId: "id", RouteName: "callback-status"})
		}
	*/

	if apiRequest.DownloadId != "" {
		apiRequest.Links = append(apiRequest.Links,
			Link{Relation: "download", Value: apiRequest.DownloadId,
				ValueId: "id", RouteName: "download"})
		apiRequest.Links = append(apiRequest.Links,
			Link{Relation: "data", Value: apiRequest.DownloadId,
				ValueId: "id", RouteName: "download-data"})
	}

	return apiRequest
}

func (r *Request) ResolveLinks(linkResolver *LinkResolver, req *http.Request) {
	linkResolver.ResolveLinks(req, &r.Links)
}
