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
	session, err := r.Connect(r.ConnectOpts{
		Address: c.Address,
		MaxIdle: c.MaxIdle,
		MaxOpen: c.MaxOpen,
	})

	if err != nil {
		return nil, err
	}

	return NewGeneralStoreWithSession(session, c.Database, tableName)
}

func (s *GeneralStore) Get(key interface{}) r.Term {
	return s.BaseTerm().Get(key)
}

func (s *GeneralStore) IndexWait() {
	s.BaseTerm().IndexWait().Exec(s.Session)
}

func (s *GeneralStore) Insert(arg interface{}, optArgs ...r.InsertOpts) error {
	_, err := s.BaseTerm().Insert(arg, optArgs...).RunWrite(s.Session)
	return err
}

func (s *GeneralStore) DeleteByKey(key interface{}, optArgs ...r.DeleteOpts) error {
	_, err := s.BaseTerm().Get(key).Delete(optArgs...).RunWrite(s.Session)
	return err
}

func (s *GeneralStore) GetAllByIndex(index interface{}, keys ...interface{}) r.Term {
	return s.BaseTerm().GetAllByIndex(index, keys...)
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

func (s *GeneralStore) BaseTerm() r.Term {
	return r.Db(s.DatabaseName).Table(s.TableName)
}
