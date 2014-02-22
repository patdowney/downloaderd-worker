package download

import (
	//	"errors"
	"github.com/patdowney/downloaderd/common"
	"io"
	"time"
)

type DownloadService struct {
	Clock       common.Clock
	IDGenerator IDGenerator

	updateChannel chan StatusUpdate
	errorChannel  chan DownloadError
	downloadQueue chan Download

	WorkerCount uint
	QueueLength uint

	fileStore     FileStore
	downloadStore DownloadStore
	requestStore  RequestStore
}

func NewDownloadService(downloadStore DownloadStore, fileStore FileStore, workerCount uint, queueLength uint) *DownloadService {
	s := DownloadService{
		IDGenerator:   &UUIDGenerator{},
		Clock:         &common.RealClock{},
		WorkerCount:   workerCount,
		QueueLength:   queueLength,
		updateChannel: make(chan StatusUpdate), //, queueLength),
		errorChannel:  make(chan DownloadError, workerCount),
		downloadQueue: make(chan Download, queueLength),
		fileStore:     fileStore,
		downloadStore: downloadStore}

	return &s
}

func (s *DownloadService) StartWorkers() {
	for workerId := uint(0); workerId < s.WorkerCount; workerId++ {
		w := NewWorker(workerId, s.downloadQueue, s.updateChannel, s.errorChannel, s.fileStore)
		w.start()
	}
}

func (s *DownloadService) ProcessError(downloadError *DownloadError) {
	download, _ := s.FindById(downloadError.DownloadId)

	if download != nil {
		download.Errors = append(download.Errors, *downloadError)
	} else {
		e := DownloadError{DownloadId: downloadError.DownloadId}
		e.Time = s.Clock.Now()
		e.OriginalError = "status received before metadata"
		s.errorChannel <- e
	}
}

func (s *DownloadService) ProcessStatusUpdate(statusUpdate *StatusUpdate) {
	download, _ := s.FindById(statusUpdate.DownloadId)

	if download != nil {
		var beginningOfTime time.Time
		if download.TimeStarted == beginningOfTime {
			download.TimeStarted = statusUpdate.Time
		}
		download.Checksum = statusUpdate.Checksum
		download.Finished = statusUpdate.Finished
		download.Status.AddStatusUpdate(statusUpdate)
		s.downloadStore.Update(download)
	} else {
		e := DownloadError{DownloadId: statusUpdate.DownloadId}
		e.Time = s.Clock.Now()
		e.OriginalError = "status received before metadata"
		s.errorChannel <- e
	}
}

func (s *DownloadService) StartEventHandlers() {
	go func() {
		for {
			select {
			case downloadError := <-s.errorChannel:
				s.ProcessError(&downloadError)
			case statusUpdate := <-s.updateChannel:
				s.ProcessStatusUpdate(&statusUpdate)
			}
		}
	}()
}

func (s *DownloadService) Start() {
	s.StartWorkers()
	s.StartEventHandlers()
}

func (s *DownloadService) ProcessRequest(downloadRequest *Request) (*Download, error) {
	id, err := s.IDGenerator.GenerateID()
	if err != nil {
		return nil, err
	}

	download := NewDownload(id, downloadRequest, s.Clock.Now())
	if downloadRequest.Callback != "" {
		download.AddRequestCallback(downloadRequest)
	}
	err = s.downloadStore.Add(download)

	s.downloadQueue <- *download

	return download, err
}

func (s *DownloadService) ListAll() ([]*Download, error) {
	return s.downloadStore.ListAll()
}

func (s *DownloadService) FindById(id string) (*Download, error) {
	return s.downloadStore.FindById(id)
}

func (s *DownloadService) GetReader(download *Download) (io.Reader, error) {
	return s.fileStore.GetReader(download)
}
