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
		err := s.baseTerm().IndexCreateFunc(
			"ResourceKey",
			URLETagIndex).Exec(s.Session)
		if err != nil {
			return err
		}
	}

	exists, err = IndexExists(s.Session, s.DatabaseName, s.TableName, "Finished")
	if !exists {
		err = s.baseTerm().IndexCreate("Finished").Exec(s.Session)
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

func (s *DownloadStore) purgeIncomplete() error {
	s.baseTerm().GetAllByIndex("Finished", true).Filter(IsIncomplete()).Delete().RunWrite(s.Session)

	return nil
}

func (s *DownloadStore) purgeUnfinished() error {
	s.baseTerm().GetAllByIndex("Finished", false).Delete().RunWrite(s.Session)

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

func (s *DownloadStore) getMultiDownload(term r.RqlTerm, offset uint, count uint) ([]*download.Download, error) {
	rows, err := term.Slice(offset, (offset + count)).Run(s.Session)
	if err != nil {
		results := make([]*download.Download, 0, 0)
		return results, err
	}

	resultCount, _ := rows.Count()
	results := make([]*download.Download, 0, resultCount)

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

func (s *DownloadStore) FindWaiting(offset uint, count uint) ([]*download.Download, error) {
	notStartedLookup := s.baseTerm().GetAllByIndex("Finished", false).Filter(NotStarted())

	return s.getMultiDownload(notStartedLookup, offset, count)
}

func (s *DownloadStore) FindNotFinished(offset uint, count uint) ([]*download.Download, error) {
	notFinishedLookup := s.baseTerm().GetAllByIndex("Finished", false)

	return s.getMultiDownload(notFinishedLookup, offset, count)
}

func (s *DownloadStore) FindFinished(offset uint, count uint) ([]*download.Download, error) {
	finishedLookup := s.baseTerm().GetAllByIndex("Finished", true)

	return s.getMultiDownload(finishedLookup, offset, count)
}

func (s *DownloadStore) FindInProgress(offset uint, count uint) ([]*download.Download, error) {
	inProgressLookup := s.baseTerm().GetAllByIndex("Finished", false).Filter(Started())

	return s.getMultiDownload(inProgressLookup, offset, count)
}

func (s *DownloadStore) FindAll(offset uint, count uint) ([]*download.Download, error) {
	allLookup := s.baseTerm()

	return s.getMultiDownload(allLookup, offset, count)
}

func (s *DownloadStore) Init() error {
	InitDatabaseAndTable(s.Session, s.DatabaseName, s.TableName)

	s.createIndexes()

	s.purgeUnfinished()
	s.purgeIncomplete()

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
