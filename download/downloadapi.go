package download

import (
	"encoding/json"
	"github.com/patdowney/downloaderd-worker/api"
	"log"
	"time"
)

func ToAPIDownloadList(origList *[]*Download) *[]*api.Download {
	rs := make([]*api.Download, len(*origList))

	for i, r := range *origList {
		rs[i] = ToAPIDownload(r)
	}

	return &rs
}

func ToAPIDownload(dd *Download) *api.Download {
	d := &api.Download{
		ID:            dd.ID,
		URL:           dd.URL,
		Checksum:      dd.Checksum,
		ChecksumType:  dd.ChecksumType,
		TimeStarted:   dd.TimeStarted,
		TimeRequested: dd.TimeRequested,
		Finished:      dd.Finished,
		Links:         make([]api.Link, 0)}

	if dd.Metadata != nil {
		d.Metadata = ToAPIMetadata(dd.Metadata)
	}

	if dd.Status != nil {
		d.BytesRead = dd.Status.BytesRead
		d.TimeUpdated = dd.Status.UpdateTime

		d.Duration = dd.Duration() / time.Millisecond
		d.PercentComplete = dd.PercentComplete()
		//		d.AverageBytesPerSecond = dd.AverageBytesPerSecond()
	}

	p, _ := json.MarshalIndent(dd, "", "  ")
	log.Printf("dd:%v", string(p))
	p, _ = json.MarshalIndent(d, "", "  ")
	log.Printf("d:%v", string(p))

	return d
}
