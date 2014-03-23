package download

import (
	"encoding/hex"
	"hash"

	"github.com/patdowney/downloaderd/common"
)

type StatusWriter struct {
	DownloadID   string
	Clock        common.Clock
	StatusSender StatusSender
	Hash         hash.Hash

	TotalBytesRead       int
	UpdateByteDifference int
}

func NewStatusWriter(downloadID string, statusSender StatusSender, hash hash.Hash, byteDifference int) *StatusWriter {
	sw := StatusWriter{
		Clock:                &common.RealClock{},
		DownloadID:           downloadID,
		StatusSender:         statusSender,
		Hash:                 hash,
		UpdateByteDifference: byteDifference,
		TotalBytesRead:       0}

	return &sw
}

func (s *StatusWriter) ShouldSendUpdate(oldTotalByteCount int, newTotalByteCount int) bool {
	previousDiv := int(oldTotalByteCount / s.UpdateByteDifference)
	currentDiv := int(newTotalByteCount / s.UpdateByteDifference)

	return (currentDiv > previousDiv)
}

func (s *StatusWriter) Write(bytes []byte) (int, error) {
	if s.Hash != nil {
		s.Hash.Write(bytes)
	}
	byteCount := len(bytes)

	oldTotalByteCount := s.TotalBytesRead

	s.TotalBytesRead += byteCount

	if s.ShouldSendUpdate(oldTotalByteCount, s.TotalBytesRead) {
		s.SendBytesWrittenUpdate(uint64(byteCount))
	}

	return byteCount, nil
}

func (s *StatusWriter) SendBytesWrittenUpdate(byteCount uint64) {
	s.SendUpdate(byteCount, false)
}

func (s *StatusWriter) SendStartUpdate() {
	s.SendUpdate(uint64(0), false)
}

func (s *StatusWriter) SendFinishedUpdate() {
	s.SendUpdate(uint64(0), true)
}

func (s *StatusWriter) SendUpdate(byteCount uint64, finished bool) {
	statusUpdate := StatusUpdate{
		DownloadID: s.DownloadID,
		Checksum:   s.ChecksumString(),
		Time:       s.Clock.Now(),
		BytesRead:  byteCount,
		Finished:   finished}

	s.StatusSender.SendUpdate(statusUpdate)
}

func (s *StatusWriter) Checksum() []byte {
	checksum := make([]byte, 0)
	if s.Hash != nil {
		checksum = s.Hash.Sum(checksum)
	}
	return checksum
}

func (s *StatusWriter) ChecksumString() string {
	return hex.EncodeToString(s.Checksum())
}

func (s *StatusWriter) Close() error {
	s.SendFinishedUpdate()
	return nil
}
