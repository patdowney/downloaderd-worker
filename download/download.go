package download

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"hash"
	"strings"
	"time"
)

type Download struct {
	Id            string
	Url           string
	Checksum      string
	ChecksumType  string
	Metadata      *Metadata
	Status        *Status
	TimeStarted   time.Time
	TimeRequested time.Time
	Finished      bool
	Errors        []DownloadError
	Callbacks     []Callback
}

func (d *Download) AddCallback(callback *Callback) {
	d.Callbacks = append(d.Callbacks, *callback)
}

func (d *Download) AddRequestCallback(request *Request) {
	c := NewCallback(request.Id, request.Callback)
	d.AddCallback(c)
}

func NewDownload(id string, request *Request, downloadTime time.Time) *Download {
	d := Download{
		Id:            id,
		Url:           request.Url,
		Checksum:      request.Checksum,
		ChecksumType:  request.ChecksumType,
		Metadata:      request.Metadata,
		Status:        &Status{},
		TimeRequested: downloadTime,
		Errors:        make([]DownloadError, 0),
		Callbacks:     make([]Callback, 0)}

	validatedChecksum, err := d.ValidateChecksum(d.ChecksumType)
	d.ChecksumType = validatedChecksum
	if err != nil {
		de := DownloadError{DownloadId: id}
		de.Time = downloadTime
		de.OriginalError = err.Error()
		d.Errors = append(d.Errors, de)
	}

	return &d
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

func (s *Download) ValidateChecksum(checksumType string) (string, error) {
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

	errMsg := fmt.Sprintf("No hash found for %s defaulting to %s", s.ChecksumType, "sha256")
	return "sha256", errors.New(errMsg)
}

func (s *Download) Hash() (hash.Hash, error) {
	switch strings.ToLower(s.ChecksumType) {
	case "md5":
		return md5.New(), nil
	case "sha1":
		return sha1.New(), nil
	case "sha256":
		return sha256.New(), nil
	case "sha512":
		return sha512.New(), nil
	}

	return nil, errors.New(fmt.Sprintf("Invalid checksum type %s", s.ChecksumType))
}
