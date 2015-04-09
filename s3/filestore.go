package s3

import (
	"fmt"
	"io"
	"log"
	"net/url"
	"path/filepath"

	"gopkg.in/amz.v1/aws"
	"gopkg.in/amz.v1/s3"

	"github.com/patdowney/downloaderd/download"
)

// FileStore ...
type FileStore struct {
	Bucket *s3.Bucket
}

// Config ...
type Config struct {
	BucketName string
	RegionName string
	SecretKey  string
	AccessKey  string
}

func openBucket(auth aws.Auth, region aws.Region, bucketName string) *s3.Bucket {
	s3i := s3.New(auth, region)
	return s3i.Bucket(bucketName)
}

func authFromEnvOrConfig(c Config) (aws.Auth, error) {
	auth, err := aws.EnvAuth()
	if err != nil {
		return aws.Auth{}, err
	}
	if c.AccessKey != "" {
		auth.AccessKey = c.AccessKey
	}
	if c.SecretKey != "" {
		auth.SecretKey = c.SecretKey
	}

	return auth, err
}

// NewFileStore ...
func NewFileStore(c Config) (*FileStore, error) {
	auth, err := authFromEnvOrConfig(c)
	if err != nil {
		return nil, err
	}

	region := aws.Regions[c.RegionName]
	bucket := openBucket(auth, region, c.BucketName)

	return &FileStore{Bucket: bucket}, nil
}

// SavePathFromURL ...
func (s *FileStore) SavePathFromURL(sourceURL string) string {
	urlObj, _ := url.Parse(sourceURL)

	return filepath.Join(urlObj.Host, urlObj.Path)
}

// SavePathForDownload ...
func (s *FileStore) SavePathForDownload(download *download.Download) (string, error) {
	urlObj, err := url.Parse(download.URL)
	if err != nil {
		return "", err
	}
	p := filepath.Join(urlObj.Host, urlObj.Path, download.ID)
	return p, nil
}

// GetReader ...
func (s *FileStore) GetReader(download *download.Download) (io.ReadCloser, error) {
	dataPath, err := s.SavePathForDownload(download)
	if err != nil {
		return nil, err
	}

	return s.Bucket.GetReader(dataPath)
}

func (s *FileStore) s3upload(reader io.Reader, savePath string, length int64, contentType string) error {
	return s.Bucket.PutReader(savePath, reader, length, contentType, s3.BucketOwnerFull)
}

// GetWriter ...
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

	listResponse, err := s.Bucket.List(s3Key, "/", "", 1)
	if err != nil {
		return nil, err
	}

	if len(listResponse.Contents) != 1 {
		return nil, fmt.Errorf("key not found: %v", s3Key)
	}

	return &listResponse.Contents[0], nil
}

// Delete ...
func (s *FileStore) Delete(download *download.Download) (bool, error) {
	savePath, err := s.SavePathForDownload(download)
	if err != nil {
		return false, err
	}

	err = s.Bucket.Del(savePath)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Verify ...
func (s *FileStore) Verify(download *download.Download) (bool, error) {
	savePath, err := s.SavePathForDownload(download)
	if err != nil {
		return false, err
	}

	expectedSize := download.Metadata.Size

	fileKey, err := s.getFileInfo(savePath)
	if err != nil {
		return false, fmt.Errorf("verify(%v): %v", download.ID, err.Error())
	}

	sizeOnS3 := uint64(fileKey.Size)

	if sizeOnS3 != expectedSize {
		return false, fmt.Errorf("verify(%v):size mismatch (%v): expected=%d, actual=%d", download.ID, savePath, expectedSize, sizeOnS3)
	}

	// we're cheating - if we really meant it we'd compare checksums
	// s3 gives us the md5 checksum as fileKey.ETag

	return true, nil
}
