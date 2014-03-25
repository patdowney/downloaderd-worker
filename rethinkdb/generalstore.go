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

type IndexFunc func(row r.RqlTerm) interface{}

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

func IndexExists(s *r.Session, databaseName, tableName, indexName string) (bool, error) {
	return ListContains(s, r.Db(databaseName).Table(tableName).IndexList(), indexName)
}

func ListContains(s *r.Session, t r.RqlTerm, name string) (bool, error) {
	row, err := t.Contains(name).RunRow(s)
	if err != nil {
		return false, err
	}
	var contains bool
	err = row.Scan(&contains)
	if err != nil {
		return false, err
	}

	return contains, nil
}

func InitDatabaseAndTable(s *r.Session, databaseName string, tableName string) error {
	err := InitDatabase(s, databaseName)
	if err != nil {
		return err
	}

	err = InitTable(s, databaseName, tableName)
	if err != nil {
		return err
	}
	return nil
}

func InitDatabase(s *r.Session, databaseName string) error {
	exists, err := ListContains(s, r.DbList(), databaseName)
	if err != nil {
		return err
	}
	if !exists {
		err := r.DbCreate(databaseName).Exec(s)
		if err != nil {
			return err
		}
	}
	return nil
}

func InitTable(s *r.Session, databaseName string, tableName string) error {
	exists, err := ListContains(s, r.Db(databaseName).TableList(), tableName)
	if err != nil {
		return err
	}
	if !exists {
		_, err = r.Db(databaseName).TableCreate(tableName).RunWrite(s)
		if err != nil {
			return err
		}
	}
	return nil
}
