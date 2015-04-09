package download

import (
	"time"

	"github.com/patdowney/downloaderd/api"
)

// ToAPIDownloadStats ...
func ToAPIDownloadStats(s *Stats) *api.DownloadStats {
	as := &api.DownloadStats{
		WaitTime: api.Stat{
			Min:   s.WaitTime.Min() / float64(time.Millisecond),
			Max:   s.WaitTime.Max() / float64(time.Millisecond),
			Mean:  s.WaitTime.Mean() / float64(time.Millisecond),
			Sum:   s.WaitTime.Sum() / float64(time.Millisecond),
			Count: s.WaitTime.Count()},
		DownloadTime: api.Stat{
			Min:   s.DownloadTime.Min() / float64(time.Millisecond),
			Max:   s.DownloadTime.Max() / float64(time.Millisecond),
			Mean:  s.DownloadTime.Mean() / float64(time.Millisecond),
			Sum:   s.DownloadTime.Sum() / float64(time.Millisecond),
			Count: s.DownloadTime.Count()},
		BytesRead: api.Stat{
			Min:   s.BytesRead.Min(),
			Max:   s.BytesRead.Max(),
			Mean:  s.BytesRead.Mean(),
			Sum:   s.BytesRead.Sum(),
			Count: s.BytesRead.Count()}}

	return as
}
