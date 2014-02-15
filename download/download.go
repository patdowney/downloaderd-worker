package download

import (
	"time"
)

type Download struct {
	Id                   string
	Url                  string
	ResourceKey          ResourceKey
	ExpectedChecksum     string
	ExpectedChecksumType string
	Metadata             *Metadata
	Status               *Status
	TimeStarted          time.Time
	TimeRequested        time.Time
	Finished             bool
	Errors               []DownloadError
}

func NewDownload(id string, request *Request, downloadTime time.Time) *Download {
	return &Download{
		Id:                   id,
		Url:                  request.Url,
		ExpectedChecksum:     request.Checksum,
		ExpectedChecksumType: request.ChecksumType,
		ResourceKey:          request.ResourceKey,
		Metadata:             request.Metadata,
		Status:               &Status{},
		TimeRequested:        downloadTime,
		Errors:               make([]DownloadError, 0)}

}

func (s *Download) PercentComplete() float32 {
	return float32(100 * (float64(s.Status.BytesRead) / float64(s.Metadata.Size)))
}

func (s *Download) Duration() time.Duration {
	return s.Status.UpdateTime.Sub(s.TimeStarted)
}

func (s *Download) AverageBytesPerSecond() float32 {
	return float32(float64(s.Status.BytesRead) / s.Duration().Seconds())
}
