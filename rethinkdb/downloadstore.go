package rethinkdb

import (
	r "github.com/dancannon/gorethink"
	"github.com/patdowney/downloaderd/download"
)

type DownloadStore struct {
	GeneralStore
}

func (s *DownloadStore) createIndexes() error {
	err := s.IndexCreateWithFunc("ResourceKey", URLETagIndex)
	if err != nil {
		return err
	}

	err = s.IndexCreate("Finished")
	if err != nil {
		return err
	}

	s.IndexWait()

	return nil
}

func (s *DownloadStore) Delete(download *download.Download) error {
	err := s.DeleteByKey(download.ID)
	return err
}

func (s *DownloadStore) Add(download *download.Download) error {
	err := s.Insert(download)
	return err
}

func (s *DownloadStore) Update(download *download.Download) error {
	_, err := s.Get(download.ID).Update(download).RunWrite(s.Session)
	return err
}

func (s *DownloadStore) purgeIncomplete() error {
	s.GetAllByIndex("Finished", true).Filter(IsIncomplete()).Delete().RunWrite(s.Session)

	return nil
}

func (s *DownloadStore) purgeUnfinished() error {
	s.GetAllByIndex("Finished", false).Delete().RunWrite(s.Session)

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

func (s *DownloadStore) FindByID(downloadID string) (*download.Download, error) {
	idLookup := s.Get(downloadID)

	return s.getSingleDownload(idLookup)
}

func (s *DownloadStore) FindByResourceKey(resourceKey download.ResourceKey) (*download.Download, error) {
	resourceKeyLookup := s.GetAllByIndex("ResourceKey", []interface{}{resourceKey.URL, resourceKey.ETag})

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
	notStartedLookup := s.GetAllByIndex("Finished", false).Filter(NotStarted())

	return s.getMultiDownload(notStartedLookup, offset, count)
}

func (s *DownloadStore) FindNotFinished(offset uint, count uint) ([]*download.Download, error) {
	notFinishedLookup := s.GetAllByIndex("Finished", false)

	return s.getMultiDownload(notFinishedLookup, offset, count)
}

func (s *DownloadStore) FindFinished(offset uint, count uint) ([]*download.Download, error) {
	finishedLookup := s.GetAllByIndex("Finished", true)

	return s.getMultiDownload(finishedLookup, offset, count)
}

func (s *DownloadStore) FindInProgress(offset uint, count uint) ([]*download.Download, error) {
	inProgressLookup := s.GetAllByIndex("Finished", false).Filter(Started())

	return s.getMultiDownload(inProgressLookup, offset, count)
}

func (s *DownloadStore) FindAll(offset uint, count uint) ([]*download.Download, error) {
	allLookup := s.BaseTerm()

	return s.getMultiDownload(allLookup, offset, count)
}

func (s *DownloadStore) Init() error {
	s.createIndexes()

	s.purgeUnfinished()
	s.purgeIncomplete()

	return nil
}

func NewDownloadStoreWithSession(s *r.Session, dbName string, tableName string) (*DownloadStore, error) {

	generalStore, err := NewGeneralStoreWithSession(s, dbName, tableName)
	if err != nil {
		return nil, err
	}

	downloadStore := &DownloadStore{}
	downloadStore.GeneralStore = *generalStore

	err = downloadStore.Init()
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
