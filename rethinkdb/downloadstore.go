package rethinkdb

import (
	r "github.com/dancannon/gorethink"
	"github.com/patdowney/downloaderd/download"

	"log"
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
	log.Printf("insert: %v", download.TimeStarted)
	err := s.Insert(download)

	d, _ := s.FindByID(download.ID)
	log.Printf("get: %v", d.TimeStarted)

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

func (s *DownloadStore) getSingleDownload(term r.Term) (*download.Download, error) {
	row, err := term.Run(s.Session)

	if err != nil {
		return nil, err
	}

	if row.IsNil() {
		return nil, nil
	}

	var download download.Download
	err = row.One(&download)
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

func (s *DownloadStore) getMultiDownload(term r.Term, offset uint, count uint) ([]*download.Download, error) {
	var results []*download.Download

	rows, err := term.Slice(offset, (offset + count)).Run(s.Session)
	if err != nil {
		return results, err
	}

	err = rows.All(&results)
	if err != nil {
		return nil, err
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
	session, err := r.Connect(r.ConnectOpts{
		Address: c.Address,
		MaxIdle: c.MaxIdle,
		MaxOpen: c.MaxOpen,
	})

	if err != nil {
		return nil, err
	}

	return NewDownloadStoreWithSession(session, c.Database, "DownloadStore")
}
