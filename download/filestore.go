package download

import (
	//"bufio"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type FileStore interface {
	GetWriter(*Download) (io.WriteCloser, error)
	GetReader(*Download) (io.ReadCloser, error)
}

type LocalFileStore struct {
	RootDirectory string
}

func NewLocalFileStore(rootDirectory string) FileStore {
	return &LocalFileStore{RootDirectory: rootDirectory}
}

func (us *LocalFileStore) SavePathFromUrl(sourceUrl string) string {
	urlObj, _ := url.Parse(sourceUrl)
	//return urlObj.Path[1:]
	return filepath.Join(urlObj.Host, urlObj.Path)
}

func (us *LocalFileStore) SavePathForOrder(download *Download) (string, error) {
	savePathFromUrl := us.SavePathFromUrl(download.Url)
	cleanRootDirectory := filepath.Clean(us.RootDirectory)
	dirtySavePath := filepath.Join(us.RootDirectory, savePathFromUrl)
	cleanSavePath := filepath.Clean(dirtySavePath)

	//ensure cleanSavePath starts with us.RootDirectory
	if !strings.HasPrefix(cleanSavePath, cleanRootDirectory) {
		return "", errors.New(fmt.Sprintf("localurlsaver: %s doesn't contain %s", cleanRootDirectory, cleanSavePath))
	}

	return cleanSavePath, nil
}

func (us *LocalFileStore) GetReader(download *Download) (io.ReadCloser, error) {
	dataPath, err := us.SavePathForOrder(download)
	if err != nil {
		return nil, err
	}

	openFile, err := os.Open(dataPath)
	if err != nil {
		return nil, err
	}

	return openFile, nil
}

func (us *LocalFileStore) GetWriter(download *Download) (io.WriteCloser, error) {
	savePath, err := us.SavePathForOrder(download)
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
