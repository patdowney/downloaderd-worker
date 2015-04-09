package download

import (
	"github.com/patdowney/downloaderd-worker/api"
	"github.com/patdowney/downloaderd-worker/common"
)

func ToAPIError(e *common.ErrorWrapper) *api.Error {
	err := &api.Error{Time: e.Time}
	if e.OriginalError != "" {
		err.Error = e.OriginalError
	} else {
		err.Error = "error missing - weird."
	}

	return err
}
