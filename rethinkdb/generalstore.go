package rethinkdb

import (
	r "github.com/dancannon/gorethink"
)

type GeneralStore struct {
	DatabaseName string
	TableName    string
	Session      *r.Session
}

func NewGeneralStoreWithSession(s *r.Session, databaseName string, tableName string) (*GeneralStore, error) {

	generalStore := &GeneralStore{
		Session:      s,
		DatabaseName: databaseName,
		TableName:    tableName}

	err := generalStore.Init()
	if err != nil {
		return nil, err
	}

	return generalStore, nil

}

func NewGeneralStore(c Config, tableName string) (*GeneralStore, error) {
	session, err := r.Connect(map[string]interface{}{
		"address":   c.Address,
		"maxIdle":   c.MaxIdle,
		"maxActive": c.MaxActive,
	})

	if err != nil {
		return nil, err
	}

	return NewGeneralStoreWithSession(session, c.Database, tableName)
}

func (s *GeneralStore) Init() error {
	return InitDatabaseAndTable(s.Session, s.DatabaseName, s.TableName)
}

func (s *GeneralStore) IndexCreateWithFunc(name string, indexFunc IndexFunc) error {
	exists, err := IndexExists(s.Session, s.DatabaseName, s.TableName, name)
	if err != nil {
		return err
	}
	if !exists {
		err = s.BaseTerm().IndexCreateFunc(name, indexFunc).Exec(s.Session)
		if err != nil {
			return err
		}
	}

	s.BaseTerm().IndexWait().Exec(s.Session)

	return nil
}

func (s *GeneralStore) IndexCreate(field string) error {
	exists, err := IndexExists(s.Session, s.DatabaseName, s.TableName, field)
	if err != nil {
		return err
	}
	if !exists {
		err = s.BaseTerm().IndexCreate(field).Exec(s.Session)
		if err != nil {
			return err
		}
	}

	s.BaseTerm().IndexWait().Exec(s.Session)

	return nil
}

func (s *GeneralStore) BaseTerm() r.RqlTerm {
	return r.Db(s.DatabaseName).Table(s.TableName)
}
