package s3

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"log"
	"os"
)

type S3Uploader interface {
	Upload(path string) error
}

type S3ManagerUploader interface {
	Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
}

type s3Uploader struct {
	toBucket        string
	s3UploadManager S3ManagerUploader
}

func (s *s3Uploader) Upload(path string) error {
	file, err := os.Open(path)
	if err != nil {
		log.Println("Failed to open file:", path)
		return err
	}
	defer file.Close()

	input := &s3manager.UploadInput{
		Body:   file,
		Bucket: aws.String(s.toBucket),
		Key:    aws.String(path),
	}

	_, err = s.s3UploadManager.Upload(input)
	if err != nil {
		return err
	}

	return nil
}

func NewS3Uploader(s3UploadManager S3ManagerUploader, toBucket string) S3Uploader {
	return &s3Uploader{
		toBucket:        toBucket,
		s3UploadManager: s3UploadManager,
	}
}
