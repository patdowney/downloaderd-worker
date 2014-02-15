package common

import (
	"time"
)

type ErrorWrapper struct {
	Time          time.Time
	OriginalError error
}

func (e *ErrorWrapper) Error() string {
	return e.OriginalError.Error()
}

func NewErrorWrapper(err error, time time.Time) *ErrorWrapper {
	return &ErrorWrapper{OriginalError: err, Time: time}
}
