package download

import (
	"io"

	"github.com/patdowney/downloaderd/common"
)

type DownloadService struct {
	Clock       common.Clock
	IDGenerator IDGenerator

	updateChannel chan StatusUpdate
	errorChannel  chan DownloadError
	downloadQueue chan Download

	WorkerCount uint
	QueueLength uint

	HookService *HookService

	fileStore     FileStore
	downloadStore DownloadStore
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
	for workerID := uint(0); workerID < s.WorkerCount; workerID++ {
		w := NewWorker(workerID, s.downloadQueue, s.updateChannel, s.errorChannel, s.fileStore)
		w.start()
	}
}

func (s *DownloadService) ProcessError(downloadError *DownloadError) {
	download, _ := s.FindByID(downloadError.DownloadID)

	if download != nil {
		download.Errors = append(download.Errors, *downloadError)
	} else {
		e := DownloadError{DownloadID: downloadError.DownloadID}
		e.Time = s.Clock.Now()
		e.OriginalError = "status received before metadata"
		s.errorChannel <- e
	}
}

func (s *DownloadService) ProcessStatusUpdate(statusUpdate *StatusUpdate) {
	download, _ := s.FindByID(statusUpdate.DownloadID)

	if download != nil {
		download.AddStatusUpdate(statusUpdate)
		s.downloadStore.Update(download)

		if download.Finished && s.HookService != nil {
			s.HookService.Notify(download)
		}
	} else {
		e := DownloadError{DownloadID: statusUpdate.DownloadID}
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

func (s *DownloadService) createDownload(downloadRequest *Request) (*Download, error) {
	id, err := s.IDGenerator.GenerateID()
	if err != nil {
		return nil, err
	}

	download := NewDownload(id, downloadRequest, s.Clock.Now())
	if downloadRequest.Callback != "" && s.HookService != nil {
		s.HookService.Register(download.ID, downloadRequest.ID, downloadRequest.Callback)
	}
	err = s.downloadStore.Add(download)

	s.downloadQueue <- *download

	return download, err
}

func (s *DownloadService) ProcessRequest(downloadRequest *Request) (*Download, error) {
	download, err := s.downloadStore.FindByResourceKey(downloadRequest.ResourceKey())
	if err != nil {
		return nil, err
	}

	if download != nil {
		// notify request callback
		if downloadRequest.Callback != "" && s.HookService != nil {
			s.HookService.Register(download.ID, downloadRequest.ID, downloadRequest.Callback)
			s.HookService.Notify(download)
		}
		return download, err
	}

	download, err = s.createDownload(downloadRequest)

	return download, err
}

func (s *DownloadService) ListAll() ([]*Download, error) {
	return s.downloadStore.ListAll()
}

func (s *DownloadService) FindByID(id string) (*Download, error) {
	return s.downloadStore.FindByID(id)
}

func (s *DownloadService) GetReader(download *Download) (io.Reader, error) {
	return s.fileStore.GetReader(download)
}
