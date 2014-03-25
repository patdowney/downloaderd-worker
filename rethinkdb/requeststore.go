package rethinkdb

import (
	r "github.com/dancannon/gorethink"
	"github.com/patdowney/downloaderd/download"
)

type RequestStore struct {
	GeneralStore
}

func (s *RequestStore) indexExists(name string) (bool, error) {
	row, err := s.BaseTerm().IndexList().Contains(name).RunRow(s.Session)
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

func ResourceKeyIndex(row r.RqlTerm) interface{} {
	return []interface{}{row.Field("URL"), row.Field("Metadata").Field("ETag")}
}

func (s *RequestStore) createIndexes() error {
	err := s.IndexCreateWithFunc(
		"ResourceKey",
		ResourceKeyIndex)
	if err != nil {
		return err
	}

	s.BaseTerm().IndexWait().Exec(s.Session)
	return nil
}

func (s *RequestStore) Init() error {
	return s.createIndexes()
}

func NewRequestStoreWithSession(s *r.Session, dbName string, tableName string) (*RequestStore, error) {

	generalStore, err := NewGeneralStoreWithSession(s, dbName, tableName)
	if err != nil {
		return nil, err
	}

	requestStore := &RequestStore{}
	requestStore.GeneralStore = *generalStore

	err = requestStore.Init()
	if err != nil {
		return nil, err
	}
	return requestStore, nil
}

func NewRequestStore(c Config) (*RequestStore, error) {
	session, err := r.Connect(map[string]interface{}{
		"address":   c.Address,
		"maxIdle":   c.MaxIdle,
		"maxActive": c.MaxActive,
	})
	if err != nil {
		return nil, err
	}

	return NewRequestStoreWithSession(session, c.Database, "RequestStore")
}

func (s *RequestStore) Add(request *download.Request) error {
	_, err := s.BaseTerm().Insert(request).RunWrite(s.Session)
	return err
}

func (s *RequestStore) FindByID(requestID string) (*download.Request, error) {
	idLookup := s.BaseTerm().Get(requestID)

	return s.getSingleRequest(idLookup)
}

func (s *RequestStore) FindByResourceKey(resourceKey download.ResourceKey, offset uint, count uint) ([]*download.Request, error) {
	resourceKeyLookup := s.BaseTerm().GetAllByIndex("ResourceKey", []interface{}{resourceKey.URL, resourceKey.ETag})

	return s.getMultiRequest(resourceKeyLookup, offset, count)
}

func (s *RequestStore) FindAll(offset uint, count uint) ([]*download.Request, error) {
	allLookup := s.BaseTerm()
	return s.getMultiRequest(allLookup, offset, count)
}

func (s *RequestStore) getMultiRequest(term r.RqlTerm, offset uint, count uint) ([]*download.Request, error) {
	rows, err := term.Slice(offset, (offset + count)).Run(s.Session)
	if err != nil {
		results := make([]*download.Request, 0, 0)
		return results, err
	}

	resultCount, _ := rows.Count()
	results := make([]*download.Request, 0, resultCount)

	for rows.Next() {
		var request download.Request
		err = rows.Scan(&request)
		if err != nil {
			return nil, err
		}
		results = append(results, &request)
	}
	return results, nil
}

func (s *RequestStore) getSingleRequest(term r.RqlTerm) (*download.Request, error) {
	row, err := term.RunRow(s.Session)

	if err != nil {
		return nil, err
	}

	if row.IsNil() {
		return nil, nil
	}

	var request download.Request
	err = row.Scan(&request)
	if err != nil {
		return nil, err
	}
	return &request, nil
}
