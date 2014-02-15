package download

import (
	"time"
)

type Status struct {
	BytesRead  uint64
	UpdateTime time.Time
}

func (s *Status) AddStatusUpdate(statusUpdate *StatusUpdate) {
	s.BytesRead += statusUpdate.BytesRead
	s.UpdateTime = statusUpdate.Time
}

func NewStatus(metadata *Metadata) *Status {
	return &Status{}
}
