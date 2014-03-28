package download

import (
	"hash/crc32"
	"testing"
)

type DummyStatusSender struct {
	UpdatesSent int
}

func (s *DummyStatusSender) SendUpdate(update StatusUpdate) {
	s.UpdatesSent += 1
}

func TestAccumulateTotal(t *testing.T) {
	byteDifference := 10

	sender := &DummyStatusSender{}

	w := NewStatusWriter("some-dummy-downloadid", sender, crc32.NewIEEE(), byteDifference)

	w.Write(make([]byte, 5))
	w.Write(make([]byte, 6))

	actual := w.TotalBytesRead
	expected := 11

	if actual != expected {
		t.Errorf("TotalBytesRead: expected %d updates, got %d", expected, actual)
	}

}

func TestOnlySendUpdateOnDifference(t *testing.T) {
	byteDifference := 10
	sender := &DummyStatusSender{}

	w := NewStatusWriter("some-dummy-downloadid", sender, crc32.NewIEEE(), byteDifference)

	w.Write(make([]byte, 5))
	if !(sender.UpdatesSent == 0) {
		t.Errorf("sent less than difference: expected %d updates, got %d", 0, sender.UpdatesSent)
	}

	w.Write(make([]byte, 6))

	if !(sender.UpdatesSent == 1) {
		t.Errorf("sent more than difference: expected %d updates, got %d", 1, sender.UpdatesSent)
	}

}
