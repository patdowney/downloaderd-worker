package local

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/patdowney/downloaderd-worker/download"
)

// FileStore ...
type FileStore struct {
	RootDirectory string
}

// NewFileStore ...
func NewFileStore(rootDirectory string) *FileStore {
	return &FileStore{RootDirectory: rootDirectory}
}

// SavePathFromURL ...
func (us *FileStore) SavePathFromURL(sourceURL string) string {
	urlObj, _ := url.Parse(sourceURL)

	return filepath.Join(urlObj.Host, urlObj.Path)
}

// SavePathForDownload ...
func (us *FileStore) SavePathForDownload(download *download.Download) (string, error) {
	savePathFromURL := us.SavePathFromURL(download.URL)
	cleanRootDirectory := filepath.Clean(us.RootDirectory)
	dirtySavePath := filepath.Join(us.RootDirectory, savePathFromURL)
	cleanSavePath := filepath.Clean(dirtySavePath)

	//ensure cleanSavePath starts with us.RootDirectory
	if !strings.HasPrefix(cleanSavePath, cleanRootDirectory) {
		return "", fmt.Errorf("localurlsaver: %s doesn't contain %s", cleanRootDirectory, cleanSavePath)
	}

	return cleanSavePath, nil
}

// GetReader ...
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

// GetWriter ...
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

// Delete ...
func (us *FileStore) Delete(download *download.Download) (bool, error) {
	dataPath, err := us.SavePathForDownload(download)
	if err != nil {
		return false, err
	}

	err = os.Remove(dataPath)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Verify ...
func (us *FileStore) Verify(download *download.Download) (bool, error) {
	savePath, err := us.SavePathForDownload(download)
	if err != nil {
		return false, err
	}

	fileInfo, err := os.Stat(savePath)
	if err != nil {
		return false, err
	}

	expectedSize := download.Metadata.Size
	sizeOnDisk := uint64(fileInfo.Size())

	if sizeOnDisk != expectedSize {
		return false, fmt.Errorf("size mismatch: expected=%d, actual=%d", expectedSize, sizeOnDisk)
	}

	// we're cheating - if we really meant it we'd compare checksums

	return true, nil
}
