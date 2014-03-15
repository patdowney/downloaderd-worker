package rethinkdb

import (
	r "github.com/dancannon/gorethink"
	"github.com/patdowney/downloaderd/download"
)

type DownloadStore struct {
	DatabaseName string
	TableName    string
	Session      *r.Session
}

func (s *DownloadStore) createIndexes() error {
	exists, err := IndexExists(s.Session, s.DatabaseName, s.TableName, "ResourceKey")
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

func (s *DownloadStore) Add(download *download.Download) error {
	_, err := s.baseTerm().Insert(download).RunWrite(s.Session)
	return err
}

func (s *DownloadStore) Update(download *download.Download) error {
	_, err := s.baseTerm().Get(download.ID).Update(download).RunWrite(s.Session)
	return err
}

func (s *DownloadStore) purgeUnfinished() error {
	s.baseTerm().Filter(IsNotFinished()).Delete().Run(s.Session)
	return nil
}

func (s *DownloadStore) getSingleDownload(term r.RqlTerm) (*download.Download, error) {
	row, err := term.RunRow(s.Session)

	if err != nil {
		return nil, err
	}

	if row.IsNil() {
		return nil, nil
	}

	var download download.Download
	err = row.Scan(&download)
	if err != nil {
		return nil, err
	}
	return &download, nil
}

func (s *DownloadStore) baseTerm() r.RqlTerm {
	return r.Db(s.DatabaseName).Table(s.TableName)
}

func (s *DownloadStore) FindByID(downloadID string) (*download.Download, error) {
	idLookup := s.baseTerm().Get(downloadID)

	return s.getSingleDownload(idLookup)
}

func (s *DownloadStore) FindByResourceKey(resourceKey download.ResourceKey) (*download.Download, error) {
	resourceKeyLookup := s.baseTerm().GetAllByIndex("ResourceKey", []interface{}{resourceKey.URL, resourceKey.ETag})

	return s.getSingleDownload(resourceKeyLookup)
}

func (s *DownloadStore) getMultiDownload(term r.RqlTerm) ([]*download.Download, error) {
	rows, err := term.Run(s.Session)
	if err != nil {
		results := make([]*download.Download, 0, 0)
		return results, err
	}

	count, _ := rows.Count()
	results := make([]*download.Download, 0, count)

	for rows.Next() {
		var download download.Download
		err = rows.Scan(&download)
		if err != nil {
			return nil, err
		}
		results = append(results, &download)
	}
	return results, nil
}

func (s *DownloadStore) FindWaiting() ([]*download.Download, error) {
	notStartedLookup := s.baseTerm().Filter(IsWaiting())

	return s.getMultiDownload(notStartedLookup)
}

func (s *DownloadStore) FindNotFinished() ([]*download.Download, error) {
	notFinishedLookup := s.baseTerm().Filter(IsNotFinished())

	return s.getMultiDownload(notFinishedLookup)
}

func (s *DownloadStore) FindFinished() ([]*download.Download, error) {
	finishedLookup := s.baseTerm().Filter(IsFinished())

	return s.getMultiDownload(finishedLookup)
}

func (s *DownloadStore) FindInProgress() ([]*download.Download, error) {
	inProgressLookup := s.baseTerm().Filter(InProgress())

	return s.getMultiDownload(inProgressLookup)
}

func (s *DownloadStore) FindAll() ([]*download.Download, error) {
	allLookup := s.baseTerm()

	return s.getMultiDownload(allLookup)
}

func (s *DownloadStore) Init() error {
	InitDatabaseAndTable(s.Session, s.DatabaseName, s.TableName)

	s.createIndexes()

	s.purgeUnfinished()

	return nil
}

func NewDownloadStoreWithSession(s *r.Session, dbName string, tableName string) (*DownloadStore, error) {

	downloadStore := &DownloadStore{
		Session:      s,
		DatabaseName: dbName,
		TableName:    tableName}

	err := downloadStore.Init()
	if err != nil {
		return nil, err
	}

	return downloadStore, nil
}

func NewDownloadStore(c Config) (*DownloadStore, error) {
	session, err := r.Connect(map[string]interface{}{
		"address":   c.Address,
		"maxIdle":   c.MaxIdle,
		"maxActive": c.MaxActive,
	})

	if err != nil {
		return nil, err
	}

	return NewDownloadStoreWithSession(session, c.Database, "DownloadStore")
}
