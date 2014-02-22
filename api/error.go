package api

import (
	"github.com/patdowney/downloaderd/common"
	"time"
)

type Error struct {
	Time  time.Time `json:"time"`
	Error string    `json:"error"`
}

func NewError(e *common.ErrorWrapper) *Error {
	err := &Error{Time: e.Time}
	if e.OriginalError != "" {
		err.Error = e.OriginalError
	} else {
		err.Error = "error missing - weird."
	}

	return err
}
