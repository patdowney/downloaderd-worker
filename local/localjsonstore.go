package local

import (
	"encoding/json"
	"os"
	"sync"
)

type LocalJSONStore struct {
	sync.RWMutex
	DataFile string
}

func (l *LocalJSONStore) SaveToDisk(data interface{}) error {
	var err error
	if l.DataFile != "" {
		l.RLock()
		defer l.RUnlock()
		var openFile *os.File
		openFile, err = os.Create(l.DataFile)
		if err == nil {
			if openFile != nil {
				encoder := json.NewEncoder(openFile)
				err = encoder.Encode(data)
			}
		}
	}
	return err
}

func (l *LocalJSONStore) LoadFromDisk(data interface{}) error {
	var err error
	if l.DataFile != "" {
		l.Lock()
		defer l.Unlock()

		openFile, err := os.Open(l.DataFile)
		if err == nil {
			encoder := json.NewDecoder(openFile)
			err = encoder.Decode(data)
		}
	}

	return err
}
