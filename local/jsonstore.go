package local

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"
)

// JSONStore ...
type JSONStore struct {
	sync.RWMutex
	DataFile string
}

func (l *JSONStore) writeToWriter(w io.Writer, data interface{}) error {
	encoder := json.NewEncoder(w)
	return encoder.Encode(data)

}

// SaveToDisk ...
func (l *JSONStore) SaveToDisk(data interface{}) error {
	if l.DataFile == "" {
		return errors.New("no datafile specified for json store")
	}

	l.RLock()
	defer l.RUnlock()

	openFile, err := os.Create(l.DataFile)
	if err != nil {
		return err
	}
	defer openFile.Close()

	return l.writeToWriter(openFile, data)
}

// LoadFromDisk ...
func (l *JSONStore) LoadFromDisk(data interface{}) error {
	if l.DataFile == "" {
		return nil
	}
	l.Lock()
	defer l.Unlock()

	openFile, err := os.Open(l.DataFile)
	if err != nil {
		return err
	}
	defer openFile.Close()

	encoder := json.NewDecoder(openFile)

	return encoder.Decode(data)
}
