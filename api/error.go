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
	if e.OriginalError != nil {
		err.Error = e.OriginalError.Error()
	} else {
		err.Error = "error missing - weird."
	}

	return err
}
