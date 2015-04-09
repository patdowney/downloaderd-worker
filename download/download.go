package download

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"strings"
	"time"
)

// Download ...
type Download struct {
	ID            string `gorethink:"id,omitempty"`
	URL           string
	Checksum      string
	ChecksumType  string
	Metadata      *Metadata
	Status        *Status
	TimeStarted   time.Time
	TimeRequested time.Time
	Finished      bool
	Errors        []Error
}

// NewDownload ...
func NewDownload(id string, request *Request, downloadTime time.Time) *Download {
	d := Download{
		ID:            id,
		URL:           request.URL,
		Checksum:      request.Checksum,
		ChecksumType:  request.ChecksumType,
		Status:        &Status{},
		Metadata:      &Metadata{},
		TimeRequested: downloadTime,
		Errors:        make([]Error, 0)}

	if request.ETag != "" {
		d.Metadata.ETag = request.ETag
	}

	if request.ContentLength > 0 {
		d.Metadata.Size = request.ContentLength
	}

	validatedChecksum, err := d.ValidateChecksum(d.ChecksumType)
	d.ChecksumType = validatedChecksum
	if err != nil {
		de := Error{DownloadID: id}
		de.Time = downloadTime
		de.OriginalError = err.Error()
		d.Errors = append(d.Errors, de)
	}

	return &d
}

// PercentComplete ...
func (d *Download) PercentComplete() float32 {
	if d.Metadata.Size > 0 {
		return float32(100 * (float64(d.Status.BytesRead) / float64(d.Metadata.Size)))
	}
	return 0
}

// Duration ...
func (d *Download) Duration() time.Duration {
	return d.Status.UpdateTime.Sub(d.TimeStarted)
}

// AverageBytesPerSecond ...
func (d *Download) AverageBytesPerSecond() float32 {
	return float32(float64(d.Status.BytesRead) / d.Duration().Seconds())
}

// ValidateChecksum ...
func (d *Download) ValidateChecksum(checksumType string) (string, error) {
	switch strings.ToLower(checksumType) {
	case "md5":
		return checksumType, nil
	case "sha1":
		return checksumType, nil
	case "sha256":
		return checksumType, nil
	case "sha512":
		return checksumType, nil
	}

	return "sha256", fmt.Errorf("No hash found for %s defaulting to %s", d.ChecksumType, "sha256")
}

// Hash ...
func (d *Download) Hash() (hash.Hash, error) {
	switch strings.ToLower(d.ChecksumType) {
	case "md5":
		return md5.New(), nil
	case "sha1":
		return sha1.New(), nil
	case "sha256":
		return sha256.New(), nil
	case "sha512":
		return sha512.New(), nil
	}

	return nil, fmt.Errorf("Invalid checksum type %s", d.ChecksumType)
}

// AddStatusUpdate ...
func (d *Download) AddStatusUpdate(statusUpdate *StatusUpdate) {
	var beginningOfTime time.Time
	if d.TimeStarted.UTC() == beginningOfTime.UTC() {
		d.TimeStarted = statusUpdate.Time
	}
	d.Checksum = statusUpdate.Checksum
	d.Finished = statusUpdate.Finished
	d.Status.AddStatusUpdate(statusUpdate)
}
