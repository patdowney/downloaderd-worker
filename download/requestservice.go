package download

import (
	"errors"
	"fmt"
	"github.com/nu7hatch/gouuid"
	"net/http"
	"time"
)

type RequestService struct {
	requestStore    RequestStore
	downloadService *DownloadService
}

func NewRequestService(requestStore RequestStore, downloadService *DownloadService) *RequestService {
	s := RequestService{
		requestStore:    requestStore,
		downloadService: downloadService}

	return &s
}

func (r *RequestService) ProcessNewRequest(downloadRequest *Request) (*Request, error) {
	id, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}

	downloadRequest.Id = id.String()
	downloadRequest.TimeRequested = time.Now()

	m, err := GetMetadataFromHead(downloadRequest)
	if err != nil {
		downloadRequest.AddError(err)
	} else {
		downloadRequest.Metadata = m
		if m.StatusCode == http.StatusOK {
			download, err := r.downloadService.ProcessRequest(downloadRequest)
			if err != nil {
				downloadRequest.AddError(err)
			}
			downloadRequest.Download = download
		} else {
			em := fmt.Sprintf("need-a-better-error: non-200 response from %s", downloadRequest.Url)
			err = errors.New(em)

			downloadRequest.AddError(err)
		}
	}

	err = r.requestStore.Add(downloadRequest)
	if err != nil {
		downloadRequest.AddError(err)
	}

	return downloadRequest, err
}

func (r *RequestService) ListAll() ([]*Request, error) {
	return r.requestStore.ListAll()
}

func (r *RequestService) FindById(id string) (*Request, error) {
	return r.requestStore.FindById(id)
}
