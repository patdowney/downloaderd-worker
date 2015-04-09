package download

import (
	"github.com/patdowney/downloaderd-common/common"
	"github.com/patdowney/downloaderd-worker/api"
)

func ToAPIError(e *common.TimestampedError) *api.Error {
	err := &api.Error{Time: e.Time}
	if e.OriginalError != "" {
		err.Error = e.OriginalError
	} else {
		err.Error = "error missing - weird."
	}

	return err
}
