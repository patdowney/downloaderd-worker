package api

import (
	"net/http"
	"time"
)

type Request struct {
	ID                   string    `json:"id"`
	URL                  string    `json:"url"`
	ExpectedChecksum     string    `json:"expected_checksum,omitempty"`
	ExpectedChecksumType string    `json:"expected_checksum_type,omitempty"`
	TimeRequested        time.Time `json:"time_requested"`
	Callback             string    `json:"callback,omitempty"`
	DownloadID           string    `json:"download_id,omitempty"`
	Errors               []*Error  `json:"errors,omitempty"`
	Metadata             *Metadata `json:"metadata,omitempty"`
	Links                []Link    `json:"links"`
}

func (r *Request) ResolveLinks(linkResolver *LinkResolver, req *http.Request) {
	// add links here
	r.Links = append(r.Links,
		Link{Relation: "self", Value: r.ID,
			ValueID: "id", RouteName: "request"})

	/*
		if r.Callback != "" {
			r.Links = append(r.Links,
				Link{Relation: "callback-status", Value: r.Callback, ValueID: "id", RouteName: "callback-status"})
		}
	*/

	if r.DownloadID != "" {
		r.Links = append(r.Links,
			Link{Relation: "download", Value: r.DownloadID,
				ValueID: "id", RouteName: "download"})
		r.Links = append(r.Links,
			Link{Relation: "data", Value: r.DownloadID,
				ValueID: "id", RouteName: "download-data"})
	}

	linkResolver.ResolveLinks(req, &r.Links)
}
