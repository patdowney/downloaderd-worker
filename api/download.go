package api

import (
	"net/http"
	"time"
)

type Download struct {
	ID            string    `json:"id"`
	URL           string    `json:"url"`
	Checksum      string    `json:"checksum,omitempty"`
	ChecksumType  string    `json:"checksum_type,omitempty"`
	Metadata      *Metadata `json:"metadata"`
	BytesRead     uint64    `json:"bytes_read"`
	TimeStarted   time.Time `json:"time_started,omitempty"`
	TimeRequested time.Time `json:"time_requested"`
	TimeUpdated   time.Time `json:"time_updated,omitempty"`
	Finished      bool      `json:"finished"`

	Duration        time.Duration `json:"duration,omitempty"`
	PercentComplete float32       `json:"percent_complete,omitempty"`
	Links           []Link        `json:"links,omitempty"`
}

func (d *Download) ResolveLinks(linkResolver *LinkResolver, req *http.Request) {
	// somehow populate links
	d.Links = append(d.Links,
		Link{Relation: "self", Value: d.ID,
			ValueID: "id", RouteName: "download"})
	d.Links = append(d.Links,
		Link{Relation: "data", Value: d.ID,
			ValueID: "id", RouteName: "download-data"})
	d.Links = append(d.Links,
		Link{Relation: "verify", Value: d.ID,
			ValueID: "id", RouteName: "download-verify"})

	linkResolver.ResolveLinks(req, &d.Links)
}
