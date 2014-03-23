package rethinkdb

import (
	"log"

	r "github.com/dancannon/gorethink"
	"github.com/patdowney/downloaderd/download"
)

type RequestStore struct {
	DatabaseName string
	TableName    string
	Session      *r.Session
}

func (s *RequestStore) baseTerm() r.RqlTerm {
	return r.Db(s.DatabaseName).Table(s.TableName)
}

func (s *RequestStore) indexExists(name string) (bool, error) {
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

func (s *RequestStore) createIndexes() error {
	exists, err := s.indexExists("ResourceKey")
	if err != nil {
		return err
	}
	if !exists {
		_, err := s.baseTerm().IndexCreateFunc("ResourceKey", func(row r.RqlTerm) interface{} {
			return []interface{}{row.Field("URL"), row.Field("Metadata").Field("ETag")}
		}).RunWrite(s.Session)
		if err != nil {
			return err
		}
	}

	s.baseTerm().IndexWait().Exec(s.Session)
	return nil
}

func (s *RequestStore) Init() error {
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

func NewRequestStoreWithSession(s *r.Session, dbName string, tableName string) (*RequestStore, error) {
	requestStore := &RequestStore{
		Session:      s,
		DatabaseName: dbName,
		TableName:    tableName}

	err := requestStore.Init()
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
	_, err := s.baseTerm().Insert(request).RunWrite(s.Session)
	return err
}

func (s *RequestStore) FindByID(requestID string) (*download.Request, error) {
	idLookup := s.baseTerm().Get(requestID)

	return s.getSingleRequest(idLookup)
}

func (s *RequestStore) FindByResourceKey(resourceKey download.ResourceKey, offset uint, count uint) ([]*download.Request, error) {
	resourceKeyLookup := s.baseTerm().GetAllByIndex("ResourceKey", []interface{}{resourceKey.URL, resourceKey.ETag})

	return s.getMultiRequest(resourceKeyLookup, offset, count)
}

func (s *RequestStore) FindAll(offset uint, count uint) ([]*download.Request, error) {
	allLookup := s.baseTerm()
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
