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

type LocalFileStore struct {
	RootDirectory string
}

func NewFileStore(rootDirectory string) download.FileStore {
	return &LocalFileStore{RootDirectory: rootDirectory}
}

func (us *LocalFileStore) SavePathFromUrl(sourceUrl string) string {
	urlObj, _ := url.Parse(sourceUrl)

	return filepath.Join(urlObj.Host, urlObj.Path)
}

func (us *LocalFileStore) SavePathForOrder(order *download.Download) (string, error) {
	savePathFromUrl := us.SavePathFromUrl(order.Url)
	cleanRootDirectory := filepath.Clean(us.RootDirectory)
	dirtySavePath := filepath.Join(us.RootDirectory, savePathFromUrl)
	cleanSavePath := filepath.Clean(dirtySavePath)

	//ensure cleanSavePath starts with us.RootDirectory
	if !strings.HasPrefix(cleanSavePath, cleanRootDirectory) {
		return "", errors.New(fmt.Sprintf("localurlsaver: %s doesn't contain %s", cleanRootDirectory, cleanSavePath))
	}

	return cleanSavePath, nil
}

func (us *LocalFileStore) GetReader(order *download.Download) (io.ReadCloser, error) {
	dataPath, err := us.SavePathForOrder(order)
	if err != nil {
		return nil, err
	}

	openFile, err := os.Open(dataPath)
	if err != nil {
		return nil, err
	}

	return openFile, nil
}

func (us *LocalFileStore) GetWriter(order *download.Download) (io.WriteCloser, error) {
	savePath, err := us.SavePathForOrder(order)
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
