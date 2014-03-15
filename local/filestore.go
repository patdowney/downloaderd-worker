package local

import (
	//"bufio"
	"errors"
	"fmt"
	"github.com/patdowney/downloaderd/download"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type FileStore struct {
	RootDirectory string
}

func NewFileStore(rootDirectory string) *FileStore {
	return &FileStore{RootDirectory: rootDirectory}
}

func (us *FileStore) SavePathFromURL(sourceURL string) string {
	urlObj, _ := url.Parse(sourceURL)

	return filepath.Join(urlObj.Host, urlObj.Path)
}

func (us *FileStore) SavePathForDownload(download *download.Download) (string, error) {
	savePathFromURL := us.SavePathFromURL(download.URL)
	cleanRootDirectory := filepath.Clean(us.RootDirectory)
	dirtySavePath := filepath.Join(us.RootDirectory, savePathFromURL)
	cleanSavePath := filepath.Clean(dirtySavePath)

	//ensure cleanSavePath starts with us.RootDirectory
	if !strings.HasPrefix(cleanSavePath, cleanRootDirectory) {
		return "", errors.New(fmt.Sprintf("localurlsaver: %s doesn't contain %s", cleanRootDirectory, cleanSavePath))
	}

	return cleanSavePath, nil
}

func (us *FileStore) GetReader(download *download.Download) (io.ReadCloser, error) {
	dataPath, err := us.SavePathForDownload(download)
	if err != nil {
		return nil, err
	}

	openFile, err := os.Open(dataPath)
	if err != nil {
		return nil, err
	}

	return openFile, nil
}

func (us *FileStore) GetWriter(download *download.Download) (io.WriteCloser, error) {
	savePath, err := us.SavePathForDownload(download)
	if err != nil {
		return nil, err
	}

	err = os.MkdirAll(filepath.Dir(savePath), os.ModeDir|0755)
	if err != nil {
		return nil, err
	}

	saveFile, err := os.Create(savePath)
	if err != nil {
		return nil, err
	}

	return saveFile, nil
}
