package local

import (
	"github.com/patdowney/downloaderd/download"
	"sync"
)

type LocalRequestStore struct {
	LocalJSONStore
	sync.RWMutex
	repository []*download.Request
}

func NewRequestStore(dataFile string) (download.RequestStore, error) {
	requestStore := &LocalRequestStore{
		repository: make([]*download.Request, 0)}

	requestStore.DataFile = dataFile

	err := requestStore.LoadFromDisk(&requestStore.repository)

	return requestStore, err
}

func (s *LocalRequestStore) Add(request *download.Request) error {
	s.Lock()
	defer s.Unlock()
	s.repository = append(s.repository, request)

	err := s.SaveToDisk(s.repository)

	return err
}

func (s *LocalRequestStore) FindById(requestId string) (*download.Request, error) {
	s.RLock()
	defer s.RUnlock()
	for _, request := range s.repository {
		if request.Id == requestId {
			return request, nil
		}
	}
	return nil, nil
}

func (s *LocalRequestStore) FindByResourceKey(resourceKey download.ResourceKey) ([]*download.Request, error) {
	s.RLock()
	defer s.RUnlock()
	results := make([]*download.Request, 0, len(s.repository))
	for _, request := range s.repository {
		if request.ResourceKey() == resourceKey {
			results = append(results, request)
		}
	}
	return results, nil
}

func (s *LocalRequestStore) ListAll() ([]*download.Request, error) {
	s.RLock()
	defer s.RUnlock()

	tmpRepository := make([]*download.Request, len(s.repository), len(s.repository))
	copy(tmpRepository, s.repository)

	return tmpRepository, nil
}
