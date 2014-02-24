package local

import (
	"github.com/patdowney/downloaderd/download"
	"sync"
)

type LocalHookStore struct {
	LocalJSONStore
	sync.RWMutex
	repository []*download.Hook
}

func NewHookStore(dataFile string) (download.HookStore, error) {
	hookStore := &LocalHookStore{
		repository: make([]*download.Hook, 0)}

	hookStore.DataFile = dataFile

	err := hookStore.LoadFromDisk(&hookStore.repository)

	return hookStore, err
}

func (s *LocalHookStore) Add(hook *download.Hook) error {
	s.Lock()
	defer s.Unlock()
	s.repository = append(s.repository, hook)

	err := s.SaveToDisk(s.repository)

	return err
}

func (s *LocalHookStore) Update(h *download.Hook) error {
	s.Lock()
	defer s.Unlock()

	indexToDelete := -1

	for i, hook := range s.repository {
		if hook.DownloadID == h.DownloadID && hook.RequestID == h.RequestID {
			indexToDelete = i
		}
	}

	s.repository = append(s.repository[:indexToDelete], s.repository[indexToDelete+1:]...)
	s.repository = append(s.repository, h)

	err := s.SaveToDisk(s.repository)

	return err
}

func (s *LocalHookStore) FindByDownloadID(downloadID string) ([]*download.Hook, error) {
	s.RLock()
	defer s.RUnlock()
	results := make([]*download.Hook, 0, len(s.repository))
	for _, hook := range s.repository {
		if hook.DownloadID == downloadID {
			results = append(results, hook)
		}
	}
	return results, nil
}

func (s *LocalHookStore) FindByRequestID(requestID string) ([]*download.Hook, error) {
	s.RLock()
	defer s.RUnlock()
	results := make([]*download.Hook, 0, len(s.repository))
	for _, hook := range s.repository {
		if hook.RequestID == requestID {
			results = append(results, hook)
		}
	}
	return results, nil
}

func (s *LocalHookStore) ListAll() ([]*download.Hook, error) {
	s.RLock()
	defer s.RUnlock()

	tmpRepository := make([]*download.Hook, len(s.repository), len(s.repository))
	copy(tmpRepository, s.repository)

	return tmpRepository, nil
}
