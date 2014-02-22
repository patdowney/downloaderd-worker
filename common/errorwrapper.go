package common

import (
	"time"
)

type ErrorWrapper struct {
	Time          time.Time
	OriginalError string
}

func (e *ErrorWrapper) Error() string {
	return e.OriginalError
}

func NewErrorWrapper(err error, time time.Time) *ErrorWrapper {
	return &ErrorWrapper{OriginalError: err.Error(), Time: time}
}
