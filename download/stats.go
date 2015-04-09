package download

import (
	"time"

	"github.com/GaryBoone/GoStats/stats"
	"github.com/patdowney/downloaderd-common/common"
)

// Stats ...
type Stats struct {
	Clock        common.Clock
	WaitTime     stats.Stats
	DownloadTime stats.Stats
	BytesRead    stats.Stats
}

func (s *Stats) calculateWaitTime(d *Download) time.Duration {
	var zeroTime time.Time
	if d.TimeStarted.UTC() == zeroTime.UTC() {
		return s.Clock.Now().Sub(d.TimeRequested)
	}
	return d.TimeStarted.Sub(d.TimeRequested)
}

func (s *Stats) calculateDownloadTime(d *Download) time.Duration {
	var zeroTime time.Time
	if d.TimeStarted.UTC() == zeroTime.UTC() {
		return time.Duration(0)
	}

	return d.Status.UpdateTime.Sub(d.TimeStarted)
}

// Add ...
func (s *Stats) Add(d *Download) {
	s.WaitTime.Update(float64(s.calculateWaitTime(d)))

	s.DownloadTime.Update(float64(s.calculateDownloadTime(d)))

	s.BytesRead.Update(float64(d.Status.BytesRead))
}

// AddList ...
func (s *Stats) AddList(dl []*Download) {
	for _, download := range dl {
		s.Add(download)
	}
}
