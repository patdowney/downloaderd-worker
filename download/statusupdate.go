package download

import (
	"time"
)

type StatusUpdate struct {
	DownloadID string
	BytesRead  uint64
	Checksum   string
	Time       time.Time
	Finished   bool
}
