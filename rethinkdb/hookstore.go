package rethinkdb

import (
	"log"

	r "github.com/dancannon/gorethink"
	"github.com/patdowney/downloaderd/download"
)

type HookStore struct {
	DatabaseName string
	TableName    string
	Session      *r.Session
}

func (s *HookStore) baseTerm() r.RqlTerm {
	return r.Db(s.DatabaseName).Table(s.TableName)
}

func (s *HookStore) indexExists(name string) (bool, error) {
	row, err := s.baseTerm().IndexList().Contains(name).RunRow(s.Session)
	if err != nil {
		return false, err
	}
	var exists bool
	err = row.Scan(&exists)
	if err != nil {
		return false, err
	}

	return exists, nil
}

func (s *HookStore) createIndexes() error {
	exists, err := s.indexExists("HookKeyIndex")
	if err != nil {
		return err
	}
	if !exists {
		s.baseTerm().IndexCreateFunc("HookKeyIndex",
			func(row r.RqlTerm) interface{} {
				return []interface{}{row.Field("DownloadID"), row.Field("RequestID")}
			}).Exec(s.Session)
	}

	exists, err = s.indexExists("DownloadID")
	if err != nil {
		return err
	}
	if !exists {
		s.baseTerm().IndexCreate("DownloadID").Exec(s.Session)
	}

	exists, err = s.indexExists("RequestID")
	if err != nil {
		return err
	}
	if !exists {
		s.baseTerm().IndexCreate("RequestID").Exec(s.Session)
	}

	s.baseTerm().IndexWait().Exec(s.Session)
	return nil
}

func (s *HookStore) Add(hook *download.Hook) error {
	_, err := s.baseTerm().Insert(hook).RunWrite(s.Session)
	return err
}

func (s *HookStore) AllByHookKey(downloadID string, requestID string) r.RqlTerm {
	return s.baseTerm().GetAllByIndex("HookKeyIndex",
		[]interface{}{downloadID, requestID})
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
	downloadIDLookup := s.baseTerm().GetAllByIndex("DownloadID", downloadID)

	return s.getMultiHook(downloadIDLookup)
}

func (s *HookStore) FindByRequestID(requestID string) ([]*download.Hook, error) {
	requestIDLookup := s.baseTerm().GetAllByIndex("RequestID", requestID)

	return s.getMultiHook(requestIDLookup)
}

func (s *HookStore) ListAll() ([]*download.Hook, error) {
	allLookup := s.baseTerm()

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
	err := r.DbCreate(s.DatabaseName).Exec(s.Session)
	if err != nil {
		log.Print(err)
	}
	_, err = r.Db(s.DatabaseName).TableCreate(s.TableName).RunWrite(s.Session)
	if err != nil {
		log.Print(err)
	}

	s.createIndexes()

	return nil
}

func NewHookStoreWithSession(s *r.Session, dbName string, tableName string) (*HookStore, error) {

	hookStore := &HookStore{
		Session:      s,
		DatabaseName: dbName,
		TableName:    tableName}

	err := hookStore.Init()
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
