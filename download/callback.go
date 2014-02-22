package download

import (
	"time"
)

type Callback struct {
	Id         string
	RequestId  string
	Url        string
	Errors     []error
	StatusCode int
	Time       time.Time
}

func NewCallback(requestId string, url string) *Callback {
	c := Callback{
		RequestId: requestId,
		Url:       url,
		Errors:    make([]error, 0)}

	return &c
}
