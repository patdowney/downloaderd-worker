package download

import (
	"github.com/nu7hatch/gouuid"
	"time"
)

type DownloadService struct {
	downloadStore DownloadStore
}

func NewDownloadService(downloadStore DownloadStore) *DownloadService {
	s := DownloadService{
		downloadStore: downloadStore}

	return &s
}

func (d *DownloadService) ProcessRequest(downloadRequest *Request) (*Download, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	download := NewDownload(id.String(), downloadRequest, time.Now())
	err = d.downloadStore.Add(download)

	//	d.OrderQueue <- order

	return download, err
}

func (d *DownloadService) ListAll() ([]*Download, error) {
	return d.downloadStore.ListAll()
}

func (d *DownloadService) FindById(id string) (*Download, error) {
	return d.downloadStore.FindById(id)
}
