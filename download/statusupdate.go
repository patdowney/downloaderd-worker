package download

import (
	"time"
)

type StatusUpdate struct {
	DownloadId string
	BytesRead  uint64
	Checksum   string
	Time       time.Time
	Finished   bool
}
