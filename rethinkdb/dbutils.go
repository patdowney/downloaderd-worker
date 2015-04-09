package rethinkdb

import (
	r "github.com/dancannon/gorethink"
)

type IndexFunc func(row r.Term) interface{}

func IndexExists(s *r.Session, databaseName, tableName, indexName string) (bool, error) {
	return ListContains(s, r.Db(databaseName).Table(tableName).IndexList(), indexName)
}

func ListContains(s *r.Session, t r.Term, name string) (bool, error) {
	rows, err := t.Contains(name).Run(s)
	if err != nil {
		return false, err
	}
	var contains bool
	err = rows.One(&contains)
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
