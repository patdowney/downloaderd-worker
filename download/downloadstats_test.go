package download

import (
	"testing"
	"time"

	"github.com/patdowney/downloaderd/common"
	//	"log"
)

func createStatTestDownload(requested, started, updated string, bytesRead uint64) *Download {
	r, _ := time.Parse(time.RFC3339, requested)
	s, _ := time.Parse(time.RFC3339, started)
	u, _ := time.Parse(time.RFC3339, updated)

	d := &Download{
		TimeRequested: r,
		TimeStarted:   s,
		Status: &Status{
			BytesRead:  bytesRead,
			UpdateTime: u}}
	return d
}

func TestCalculateWaitTimeEasy(t *testing.T) {
	requested := "2014-03-16T15:04:05Z"
	started := "2014-03-16T16:04:05Z"
	var updated string
	var bytesRead uint64

	d := createStatTestDownload(requested, started, updated, bytesRead)
	s := Stats{}

	expected := time.Hour
	actual := s.calculateWaitTime(d)
	if actual != expected {
		t.Errorf("download-wait-time, expected = %v, got=%v", expected, actual)
	}

}

func TestCalculateWaitTimeNotStarted(t *testing.T) {
	requested := "2014-03-16T15:04:05Z"
	var started string
	var updated string
	var bytesRead uint64

	d := createStatTestDownload(requested, started, updated, bytesRead)

	fakeTime, _ := time.Parse(time.RFC3339, "2014-03-16T15:44:05Z")
	c := &common.FakeClock{FakeTime: fakeTime}
	s := Stats{Clock: c}

	expected := (40 * time.Minute)
	actual := s.calculateWaitTime(d)
	if actual != expected {
		t.Errorf("download-wait-time, expected = %v, got=%v", expected, actual)
	}

}

func TestCalculateDownloadTimeEasy(t *testing.T) {
	var requested string
	started := "2014-03-16T16:04:05Z"
	updated := "2014-03-16T16:14:05Z"
	var bytesRead uint64

	d := createStatTestDownload(requested, started, updated, bytesRead)
	s := Stats{}

	expected := (10 * time.Minute)
	actual := s.calculateDownloadTime(d)
	if actual != expected {
		t.Errorf("download-time, expected = %v, got=%v", expected, actual)
	}

}

func TestCalculateDownloadTimeNotStarted(t *testing.T) {
	var requested string
	var started string
	var updated string
	var bytesRead uint64

	d := createStatTestDownload(requested, started, updated, bytesRead)

	fakeTime, _ := time.Parse(time.RFC3339, "2014-03-16T15:44:05Z")
	c := &common.FakeClock{FakeTime: fakeTime}
	s := Stats{Clock: c}

	expected := 0 * time.Hour
	actual := s.calculateDownloadTime(d)
	if actual != expected {
		t.Errorf("download-time, expected = %v, got=%v", expected, actual)
	}

}

func TestWaitTimeStats(t *testing.T) {
	requested := "2014-03-16T15:04:05Z"
	started := "2014-03-16T16:04:05Z"
	updated := "2014-03-16T16:14:05Z"
	bytesRead := uint64(14)

	d1 := createStatTestDownload(requested, started, updated, bytesRead)
	d2 := createStatTestDownload(requested, started, updated, bytesRead)

	s := Stats{}
	s.Add(d1)
	s.Add(d2)

	expected := float64(time.Hour)
	actual := s.WaitTime.Mean()
	if actual != expected {
		t.Errorf("mean-wait-time, expected = %f, got=%f", expected, actual)
	}
}

func TestDownloadTimeStats(t *testing.T) {
	bytesRead := uint64(14)

	requested := "2014-03-16T16:00:05Z"
	started := "2014-03-16T16:04:05Z"
	updated := "2014-03-16T16:14:05Z"
	d1 := createStatTestDownload(requested, started, updated, bytesRead)

	started2 := "2014-03-16T16:04:05Z"
	updated2 := "2014-03-16T16:25:05Z"
	d2 := createStatTestDownload(requested, started2, updated2, bytesRead)

	s := Stats{}
	s.Add(d1)
	s.Add(d2)

	expected := (float64(10*time.Minute) + float64(21*time.Minute)) / 2.0
	actual := s.DownloadTime.Mean()
	if actual != expected {
		t.Errorf("mean-download-time, expected = %f, got=%f", expected, actual)
	}

	expected = float64(10 * time.Minute)
	actual = s.DownloadTime.Min()
	if actual != expected {
		t.Errorf("min-download-time, expected = %f, got=%f", expected, actual)
	}
}

func TestBytesReadStats(t *testing.T) {
	requested := "2014-03-16T16:00:05Z"
	started := "2014-03-16T16:04:05Z"
	updated := "2014-03-16T16:14:05Z"

	bytesRead := uint64(14)
	d1 := createStatTestDownload(requested, started, updated, bytesRead)

	bytesRead2 := uint64(26)
	d2 := createStatTestDownload(requested, started, updated, bytesRead2)

	s := Stats{}
	s.Add(d1)
	s.Add(d2)

	expected := float64(20)
	actual := s.BytesRead.Mean()
	if actual != expected {
		t.Errorf("mean-bytes-read, expected = %f, got=%f", expected, actual)
	}
}

func TestBytesReadStatsFromList(t *testing.T) {
	requested := "2014-03-16T16:00:05Z"
	started := "2014-03-16T16:04:05Z"
	updated := "2014-03-16T16:14:05Z"

	bytesRead := uint64(14)
	d1 := createStatTestDownload(requested, started, updated, bytesRead)

	bytesRead2 := uint64(26)
	d2 := createStatTestDownload(requested, started, updated, bytesRead2)

	dl := []*Download{d1, d2}

	s := Stats{}
	s.AddList(dl)
	//s.Add(d1)
	//s.Add(d2)

	expected := float64(20)
	actual := s.BytesRead.Mean()
	if actual != expected {
		t.Errorf("mean-bytes-read, expected = %f, got=%f", expected, actual)
	}
}
