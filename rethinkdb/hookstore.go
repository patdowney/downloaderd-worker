package rethinkdb

import (
	"log"

	r "github.com/dancannon/gorethink"
	"github.com/patdowney/downloaderd/download"
)

type HookStore struct {
	GeneralStore
}

func HookKeyIndex(row r.RqlTerm) interface{} {
	return []interface{}{row.Field("DownloadID"), row.Field("RequestID")}
}

func (s *HookStore) createIndexes() error {
	err := s.IndexCreateWithFunc("HookKeyIndex", HookKeyIndex)
	if err != nil {
		return err
	}

	err = s.IndexCreate("DownloadID")
	if err != nil {
		return err
	}

	err = s.IndexCreate("RequestID")
	if err != nil {
		return err
	}

	s.IndexWait()
	return nil
}

func (s *HookStore) Add(hook *download.Hook) error {
	err := s.Insert(hook)
	return err
}

func (s *HookStore) AllByHookKey(downloadID string, requestID string) r.RqlTerm {
	return s.GetAllByIndex("HookKeyIndex", []interface{}{downloadID, requestID})
}

func (s *HookStore) Update(h *download.Hook) error {
	_, err := s.AllByHookKey(h.DownloadID, h.RequestID).Update(h).RunWrite(s.Session)
	return err
}

func (s *HookStore) FindByHookKey(downloadID string, requestID string) ([]*download.Hook, error) {
	hookLookup := s.AllByHookKey(downloadID, requestID)

	return s.getMultiHook(hookLookup)
}

func (s *HookStore) FindByDownloadID(downloadID string) ([]*download.Hook, error) {
	downloadIDLookup := s.GetAllByIndex("DownloadID", downloadID)

	return s.getMultiHook(downloadIDLookup)
}

func (s *HookStore) FindByRequestID(requestID string) ([]*download.Hook, error) {
	requestIDLookup := s.GetAllByIndex("RequestID", requestID)

	return s.getMultiHook(requestIDLookup)
}

func (s *HookStore) ListAll() ([]*download.Hook, error) {
	allLookup := s.BaseTerm()

	return s.getMultiHook(allLookup)
}

func (s *HookStore) getMultiHook(term r.RqlTerm) ([]*download.Hook, error) {
	rows, err := term.Run(s.Session)
	if err != nil {
		results := make([]*download.Hook, 0, 0)
		log.Print(err)
		return results, err
	}

	count, _ := rows.Count()
	results := make([]*download.Hook, 0, count)

	for rows.Next() {
		var hook download.Hook
		err = rows.Scan(&hook)
		if err != nil {
			return nil, err
		}
		results = append(results, &hook)
	}
	return results, nil
}

func (s *HookStore) Init() error {
	return s.createIndexes()
}

func NewHookStoreWithSession(s *r.Session, dbName string, tableName string) (*HookStore, error) {

	generalStore, err := NewGeneralStoreWithSession(s, dbName, tableName)
	if err != nil {
		return nil, err
	}

	hookStore := &HookStore{}
	hookStore.GeneralStore = *generalStore

	err = hookStore.Init()
	if err != nil {
		return nil, err
	}

	return hookStore, nil
}

func NewHookStore(c Config) (*HookStore, error) {
	session, err := r.Connect(map[string]interface{}{
		"address":   c.Address,
		"maxIdle":   c.MaxIdle,
		"maxActive": c.MaxActive,
	})

	if err != nil {
		return nil, err
	}

	return NewHookStoreWithSession(session, c.Database, "HookStore")
}
