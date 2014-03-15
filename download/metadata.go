package download

import (
	"net/http"
	"strconv"
	"time"
)

type Metadata struct {
	RequestID     string
	TimeRequested time.Time
	MimeType      string
	Size          uint64

	// HTTP specific stuff
	Server       string
	LastModified time.Time
	ETag         string
	Expires      time.Time
	StatusCode   int

	Errors []string
}

func GetMetadataFromHead(requestTime time.Time, request *Request) (*Metadata, error) {
	res, err := http.Head(request.URL)
	if err != nil {
		return nil, err
	}
	metadata := NewMetadata(request, res, requestTime)

	return metadata, nil
}

func ParseTime(timeHeader string) (time.Time, error) {
	return time.Parse(time.RFC1123, timeHeader)
}

func NewMetadata(request *Request, res *http.Response, requestTime time.Time) *Metadata {

	m := &Metadata{
		RequestID:     request.ID,
		TimeRequested: requestTime,
		MimeType:      res.Header.Get("Content-Type"),
		ETag:          res.Header.Get("ETag"),
		Server:        res.Header.Get("Server"),
		StatusCode:    res.StatusCode,
		Errors:        make([]string, 0)}

	var err error
	// reference time: Mon Jan 2 15:04:05 -0700 MST 2006
	contentLengthHeader := res.Header.Get("Content-Length")
	m.Size, err = strconv.ParseUint(contentLengthHeader, 10, 64)
	if err != nil {
		m.Errors = append(m.Errors, err.Error())
	}

	m.LastModified, err = ParseTime(res.Header.Get("Last-Modified"))
	if err != nil {
		m.Errors = append(m.Errors, err.Error())
	}

	m.Expires, err = ParseTime(res.Header.Get("Expires"))
	if err != nil {
		m.Errors = append(m.Errors, err.Error())
	}

	return m
}
