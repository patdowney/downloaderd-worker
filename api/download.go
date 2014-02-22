package api

import (
	"github.com/patdowney/downloaderd/download"
	"net/http"
	"time"
)

type Download struct {
	Id            string    `json:"id"`
	Url           string    `json:"url"`
	Checksum      string    `json:"checksum,omitempty"`
	ChecksumType  string    `json:"checksum_type,omitempty"`
	Metadata      *Metadata `json:"metadata"`
	BytesRead     uint64    `json:"bytes_read"`
	TimeStarted   time.Time `json:"time_started,omitempty"`
	TimeRequested time.Time `json:"time_requested"`
	TimeUpdated   time.Time `json:"time_updated,omitempty"`
	Finished      bool      `json:"finished"`

	Duration              time.Duration `json:"duration,omitempty"`
	PercentComplete       float32       `json:"percent_complete,omitempty"`
	AverageBytesPerSecond float32       `json:"avg_bytes_per_second,omitempty"`

	Links []Link `json:"links,omitempty"`
}

func NewDownloadList(origList *[]*download.Download) *[]*Download {
	rs := make([]*Download, len(*origList))

	for i, r := range *origList {
		rs[i] = NewDownload(r)
	}

	return &rs
}

func NewDownload(dd *download.Download) *Download {
	d := &Download{
		Id:            dd.Id,
		Url:           dd.Url,
		Checksum:      dd.Checksum,
		ChecksumType:  dd.ChecksumType,
		TimeStarted:   dd.TimeStarted,
		TimeRequested: dd.TimeRequested,
		Finished:      dd.Finished,
		Links:         make([]Link, 0)}

	if dd.Metadata != nil {
		d.Metadata = NewMetadata(dd.Metadata)
	}

	if dd.Status != nil {
		d.BytesRead = dd.Status.BytesRead
		d.TimeUpdated = dd.Status.UpdateTime

		d.Duration = dd.Duration() / time.Millisecond
		d.PercentComplete = dd.PercentComplete()
		d.AverageBytesPerSecond = dd.AverageBytesPerSecond()
	}

	// somehow populate links
	d.Links = append(d.Links,
		Link{Relation: "self", Value: d.Id,
			ValueId: "id", RouteName: "download"})
	d.Links = append(d.Links,
		Link{Relation: "data", Value: d.Id,
			ValueId: "id", RouteName: "download-data"})

	return d
}

func (d *Download) ResolveLinks(linkResolver *LinkResolver, req *http.Request) {
	linkResolver.ResolveLinks(req, &d.Links)
}
