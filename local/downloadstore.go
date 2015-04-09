package local

import (
	"sync"
	"time"

	"github.com/patdowney/downloaderd-worker/download"
)

// DownloadStore ...
type DownloadStore struct {
	JSONStore
	sync.RWMutex
	repository []*download.Download
}

// Delete ...
func (s *DownloadStore) Delete(d *download.Download) error {
	s.Lock()
	newRepository := make([]*download.Download, 0, len(s.repository))

	for _, download := range s.repository {
		if download.ID != d.ID {
			newRepository = append(newRepository, download)
		}
	}
	s.repository = newRepository
	s.Unlock()

	err := s.Commit()

	return err
}

// Add ...
func (s *DownloadStore) Add(download *download.Download) error {
	s.Lock()
	s.repository = append(s.repository, download)
	s.Unlock()

	err := s.Commit()

	return err
}

// Update ...
func (s *DownloadStore) Update(download *download.Download) error {
	s.Lock()
	d, err := s.FindByID(download.ID)
	if err == nil {
		*d = *download
	}
	s.Unlock()
	err = s.Commit()

	return err
}

// Commit ...
func (s *DownloadStore) Commit() error {
	return s.SaveToDisk(s.repository)
}

func (s *DownloadStore) purgeUnfinished() error {
	newRepository := make([]*download.Download, 0, len(s.repository))

	for _, download := range s.repository {
		if download.Finished {
			newRepository = append(newRepository, download)
		}
	}
	s.repository = newRepository
	return nil
}

func (s *DownloadStore) load() error {
	err := s.LoadFromDisk(&s.repository)

	s.purgeUnfinished()

	return err
}

// FindByID ...
func (s *DownloadStore) FindByID(downloadID string) (*download.Download, error) {
	s.RLock()
	defer s.RUnlock()

	for _, download := range s.repository {
		if download.ID == downloadID {
			return download, nil
		}
	}
	return nil, nil
}

// FindByResourceKey ...
func (s *DownloadStore) FindByResourceKey(resourceKey download.ResourceKey) (*download.Download, error) {
	s.RLock()
	defer s.RUnlock()

	for _, download := range s.repository {
		if download.URL == resourceKey.URL {
			if download.Metadata != nil {
				if download.Metadata.ETag == resourceKey.ETag {
					return download, nil

				}
			}
			return download, nil
		}
	}
	return nil, nil
}

// FindAll ...
func (s *DownloadStore) FindAll(offset uint, count uint) ([]*download.Download, error) {
	s.RLock()
	defer s.RUnlock()

	tmpRepository := make([]*download.Download, len(s.repository), len(s.repository))
	copy(tmpRepository, s.repository)

	return tmpRepository, nil
}

// FindFinished ...
func (s *DownloadStore) FindFinished(offset uint, count uint) ([]*download.Download, error) {
	s.RLock()
	defer s.RUnlock()

	var tmpRepository []*download.Download
	//tmpRepository := make([]*download.Download, 0)

	repoSlice := s.sliceRepository(offset, count)
	for _, download := range repoSlice {
		if download.Finished {
			tmpRepository = append(tmpRepository, download)
		}
	}

	return tmpRepository, nil
}

// FindNotFinished ...
func (s *DownloadStore) FindNotFinished(offset uint, count uint) ([]*download.Download, error) {
	s.RLock()
	defer s.RUnlock()

	var tmpRepository []*download.Download

	repoSlice := s.sliceRepository(offset, count)
	for _, download := range repoSlice {
		if !download.Finished {
			tmpRepository = append(tmpRepository, download)
		}
	}

	return tmpRepository, nil
}

// FindInProgress ...
func (s *DownloadStore) FindInProgress(offset uint, count uint) ([]*download.Download, error) {
	s.RLock()
	defer s.RUnlock()

	var tmpRepository []*download.Download
	var beginningOfTime time.Time

	repoSlice := s.sliceRepository(offset, count)
	for _, download := range repoSlice {
		if !download.Finished {
			if download.TimeStarted.After(beginningOfTime) {
				tmpRepository = append(tmpRepository, download)
			}
		}
	}

	return tmpRepository, nil
}

func (s *DownloadStore) sliceRepository(offset uint, count uint) []*download.Download {
	return s.repository[offset:(offset + count)]
}

// FindWaiting ...
func (s *DownloadStore) FindWaiting(offset uint, count uint) ([]*download.Download, error) {
	s.RLock()
	defer s.RUnlock()

	var tmpRepository []*download.Download
	var beginningOfTime time.Time

	repoSlice := s.sliceRepository(offset, count)
	for _, download := range repoSlice {
		if !download.Finished {
			if download.TimeStarted.UTC() == beginningOfTime.UTC() {
				tmpRepository = append(tmpRepository, download)
			}
		}
	}

	return tmpRepository, nil
}

// NewDownloadStore ...
func NewDownloadStore(dataFile string) (*DownloadStore, error) {
	downloadStore := &DownloadStore{
		repository: make([]*download.Download, 0)}

	downloadStore.DataFile = dataFile
	err := downloadStore.load()

	return downloadStore, err
}
