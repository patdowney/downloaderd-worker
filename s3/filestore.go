package s3

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"path/filepath"

	"github.com/patdowney/downloaderd/download"
	"launchpad.net/goamz/aws"
	"launchpad.net/goamz/s3"
)

type FileStore struct {
	Region aws.Region

	auth aws.Auth
}

type Config struct {
	SecretKey string
	AccessKey string
}

func (s *FileStore) bucket(bucketName string) *s3.Bucket {
	s3i := s3.New(s.auth, s.Region)
	return s3i.Bucket(bucketName)
}

func NewFileStore(c Config) (*FileStore, error) {

	auth, err := aws.EnvAuth()
	if err != nil {
		return nil, err
	}
	if c.AccessKey != "" {
		auth.AccessKey = c.AccessKey
	}
	if c.SecretKey != "" {
		auth.SecretKey = c.SecretKey
	}

	return &FileStore{auth: auth, Region: aws.Regions["us-east-1"]}, nil
}

func (s *FileStore) SavePathFromURL(sourceURL string) string {
	urlObj, _ := url.Parse(sourceURL)

	return filepath.Join(urlObj.Host, urlObj.Path)
}

func (s *FileStore) SavePathForDownload(download *download.Download) (string, error) {
	urlObj, err := url.Parse(download.URL)
	if err != nil {
		return "", err
	}
	p := filepath.Join(urlObj.Host, urlObj.Path, download.ID)
	return p, nil
}

func (s *FileStore) GetReader(download *download.Download) (io.ReadCloser, error) {
	dataPath, err := s.SavePathForDownload(download)
	if err != nil {
		return nil, err
	}

	return s.bucket("downloaderd").GetReader(dataPath)
}

func (s *FileStore) s3upload(reader io.Reader, savePath string, length int64, contentType string) error {
	return s.bucket("downloaderd").PutReader(savePath, reader, length, contentType, s3.BucketOwnerFull)
}

func (s *FileStore) GetWriter(download *download.Download) (io.WriteCloser, error) {
	savePath, err := s.SavePathForDownload(download)
	if err != nil {
		return nil, err
	}

	pipeReader, pipeWriter := io.Pipe()

	go func() {
		err := s.s3upload(pipeReader, savePath, int64(download.Metadata.Size), download.Metadata.MimeType)
		if err != nil {
			log.Printf("s3-upload-failed for download:%s, savePath:%s, error:%s", download.ID, savePath, err.Error())
		}
	}()

	return pipeWriter, nil
}

func (s *FileStore) getFileInfo(s3Key string) (*s3.Key, error) {

	listResponse, err := s.bucket("downloaderd").List(s3Key, "/", "", 1)
	if err != nil {
		return nil, err
	}

	if len(listResponse.Contents) != 1 {
		return nil, errors.New(fmt.Sprintf("key not found: %v", s3Key))
	}

	return &listResponse.Contents[0], nil
}

func (s *FileStore) Verify(download *download.Download) (bool, error) {
	savePath, err := s.SavePathForDownload(download)
	if err != nil {
		return false, err
	}

	expectedSize := download.Metadata.Size

	fileKey, err := s.getFileInfo(savePath)
	if err != nil {
		return false, errors.New(fmt.Sprintf("verify(%v): %v", download.ID, err.Error()))
	}

	sizeOnS3 := uint64(fileKey.Size)

	if sizeOnS3 != expectedSize {
		return false, errors.New(fmt.Sprintf("verify(%v):size mismatch (%v): expected=%d, actual=%d", download.ID, savePath, expectedSize, sizeOnS3))
	}

	// we're cheating - if we really meant it we'd compare checksums
	// s3 gives us the md5 checksum as fileKey.ETag

	return true, nil
}
