package download

import (
	"encoding/hex"
	"github.com/patdowney/downloaderd/common"
	"hash"
)

type StatusWriter struct {
	DownloadId    string
	Clock         common.Clock
	StatusChannel chan StatusUpdate
	Hash          hash.Hash
}

func NewStatusWriter(downloadId string, statusChannel chan StatusUpdate, hash hash.Hash) *StatusWriter {
	sw := StatusWriter{
		Clock:         &common.RealClock{},
		DownloadId:    downloadId,
		StatusChannel: statusChannel,
		Hash:          hash}

	sw.SendStartUpdate()

	return &sw
}

func (s *StatusWriter) Write(bytes []byte) (int, error) {
	if s.Hash != nil {
		s.Hash.Write(bytes)
	}
	byteCount := len(bytes)

	s.SendBytesWrittenUpdate(uint64(byteCount))

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
		DownloadId: s.DownloadId,
		Checksum:   s.ChecksumString(),
		Time:       s.Clock.Now(),
		BytesRead:  byteCount,
		Finished:   finished}

	s.StatusChannel <- statusUpdate
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
