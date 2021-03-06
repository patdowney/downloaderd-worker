package s3

import (
	"testing"

	"gopkg.in/amz.v1/aws"
	"gopkg.in/amz.v1/s3"
)

func TestS3Upload(t *testing.T) {
	auth, err := aws.EnvAuth()

	s := s3.New(auth, aws.USEast)
	bucket := s.Bucket("downloaderd")

	data := []byte("Hello, Goamz!!")
	err = bucket.Put("sample.txt", data, "text/plain", s3.BucketOwnerFull)
	if err != nil {
		t.Errorf("upload-failed: %v", err)
	}

	err = bucket.Put("test/sample.txt", data, "text/plain", s3.BucketOwnerFull)
	if err != nil {
		t.Errorf("upload-failed: %v", err)
	}
}
