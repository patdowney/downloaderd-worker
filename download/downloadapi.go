package download

import (
	"time"

	"github.com/patdowney/downloaderd/api"
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

	return d
}
