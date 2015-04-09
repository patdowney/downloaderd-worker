package local

import (
	"sync"

	"github.com/patdowney/downloaderd-common/local"
	"github.com/patdowney/downloaderd-worker/download"
)

type HookStore struct {
	local.JSONStore
	sync.RWMutex
	repository []*download.Hook
}

func NewHookStore(dataFile string) (*HookStore, error) {
	hookStore := &HookStore{
		repository: make([]*download.Hook, 0)}

	hookStore.DataFile = dataFile

	err := hookStore.LoadFromDisk(&hookStore.repository)

	return hookStore, err
}

func (s *HookStore) Add(hook *download.Hook) error {
	s.Lock()
	defer s.Unlock()
	s.repository = append(s.repository, hook)

	err := s.SaveToDisk(s.repository)

	return err
}

func (s *HookStore) Update(h *download.Hook) error {
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

func (s *HookStore) FindByDownloadID(downloadID string) ([]*download.Hook, error) {
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

func (s *HookStore) FindByRequestID(requestID string) ([]*download.Hook, error) {
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

func (s *HookStore) ListAll() ([]*download.Hook, error) {
	s.RLock()
	defer s.RUnlock()

	tmpRepository := make([]*download.Hook, len(s.repository), len(s.repository))
	copy(tmpRepository, s.repository)

	return tmpRepository, nil
}
