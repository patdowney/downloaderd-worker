package download

import (
	"github.com/patdowney/downloaderd/common"
)

type DownloadService struct {
	Clock       common.Clock
	IDGenerator IDGenerator

	downloadStore DownloadStore
}

func NewDownloadService(downloadStore DownloadStore) *DownloadService {
	s := DownloadService{
		IDGenerator:   &UUIDGenerator{},
		Clock:         &common.RealClock{},
		downloadStore: downloadStore}

	return &s
}

func (s *DownloadService) ProcessRequest(downloadRequest *Request) (*Download, error) {
	id, err := s.IDGenerator.GenerateID()
	if err != nil {
		return nil, err
	}

	download := NewDownload(id, downloadRequest, s.Clock.Now())
	err = s.downloadStore.Add(download)

	//	s.OrderQueue <- order

	return download, err
}

func (s *DownloadService) ListAll() ([]*Download, error) {
	return s.downloadStore.ListAll()
}

func (s *DownloadService) FindById(id string) (*Download, error) {
	return s.downloadStore.FindById(id)
}
