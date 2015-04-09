package download

import (
	"io"
	//	"log"

	"github.com/patdowney/downloaderd/common"
)

// Service ...
type Service struct {
	Clock       common.Clock
	IDGenerator IDGenerator

	updateChannel chan StatusUpdate
	errorChannel  chan Error
	downloadQueue chan Download

	WorkerCount uint
	QueueLength uint

	HookService *HookService

	fileStore     FileStore
	downloadStore Store
}

// NewDownloadService ...
func NewDownloadService(downloadStore Store, fileStore FileStore, workerCount uint, queueLength uint) *Service {
	s := Service{
		IDGenerator:   &UUIDGenerator{},
		Clock:         &common.RealClock{},
		WorkerCount:   workerCount,
		QueueLength:   queueLength,
		updateChannel: make(chan StatusUpdate), //, queueLength),
		errorChannel:  make(chan Error, workerCount),
		downloadQueue: make(chan Download, queueLength),
		fileStore:     fileStore,
		downloadStore: downloadStore}

	return &s
}

// StartWorkers ...
func (s *Service) StartWorkers() {
	for workerID := uint(0); workerID < s.WorkerCount; workerID++ {
		w := NewWorker(workerID, s.downloadQueue, s.updateChannel, s.errorChannel, s.fileStore)
		w.start()
	}
}

// ProcessError ...
func (s *Service) ProcessError(downloadError *Error) {
	download, _ := s.FindByID(downloadError.DownloadID)

	if download != nil {
		download.Errors = append(download.Errors, *downloadError)
	} else {
		e := Error{DownloadID: downloadError.DownloadID}
		e.Time = s.Clock.Now()
		e.OriginalError = "status received before metadata"
		s.errorChannel <- e
	}
}

// ProcessStatusUpdate ...
func (s *Service) ProcessStatusUpdate(statusUpdate *StatusUpdate) {
	download, _ := s.FindByID(statusUpdate.DownloadID)

	if download != nil {
		download.AddStatusUpdate(statusUpdate)
		s.downloadStore.Update(download)

		if download.Finished && s.HookService != nil {
			s.HookService.Notify(download)
		}
	} else {
		e := Error{DownloadID: statusUpdate.DownloadID}
		e.Time = s.Clock.Now()
		e.OriginalError = "status received before metadata"
		s.errorChannel <- e
	}
}

// StartEventHandlers ...
func (s *Service) StartEventHandlers() {
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

// Start ...
func (s *Service) Start() {
	s.StartWorkers()
	s.StartEventHandlers()
}

func (s *Service) createDownload(downloadRequest *Request) (*Download, error) {
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

// ProcessRequest ...
func (s *Service) ProcessRequest(downloadRequest *Request) (*Download, error) {
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

// ListFinished ...
func (s *Service) ListFinished() ([]*Download, error) {
	return s.downloadStore.FindFinished(0, 25)
}

// ListNotFinished ...
func (s *Service) ListNotFinished() ([]*Download, error) {
	return s.downloadStore.FindNotFinished(0, 25)
}

// ListInProgress ...
func (s *Service) ListInProgress() ([]*Download, error) {
	return s.downloadStore.FindInProgress(0, 25)
}

// ListWaiting ...
func (s *Service) ListWaiting() ([]*Download, error) {
	return s.downloadStore.FindWaiting(0, 25)
}

// ListAll ...
func (s *Service) ListAll() ([]*Download, error) {
	return s.downloadStore.FindAll(0, 25)
}

// FindByID ...
func (s *Service) FindByID(id string) (*Download, error) {
	return s.downloadStore.FindByID(id)
}

// Delete ...
func (s *Service) Delete(download *Download) (bool, error) {
	_, err := s.fileStore.Delete(download)
	if err != nil {
		return false, err
	}

	err = s.downloadStore.Delete(download)
	if err != nil {
		return false, err
	}

	return true, nil
}

// DeleteByID ...
func (s *Service) DeleteByID(id string) (bool, error) {
	d, err := s.FindByID(id)
	if err != nil {
		return false, err
	}

	// needs to be handled better
	// in cases where id doesn't exist.
	if d == nil {
		return false, nil
	}

	return s.Delete(d)
}

// GetReader ...
func (s *Service) GetReader(download *Download) (io.Reader, error) {
	return s.fileStore.GetReader(download)
}

// Verify ...
func (s *Service) Verify(download *Download) (bool, error) {
	return s.fileStore.Verify(download)
}
