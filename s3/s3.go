package s3

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/sirupsen/logrus"
	"os"
)

// S3Uploader uploads files to AWS S3
type S3Uploader interface {
	Upload(path string) error
}

// S3ManagerUploader knows how to use the AWS S3 SDK to upload files. This more
// narrow interface definition replaces the dependency on the `s3manager.Uploader`
// concrete type, and aids primarily in defining simple test doubles
type S3ManagerUploader interface {
	Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error)
}

type s3Uploader struct {
	toBucket        string
	s3UploadManager S3ManagerUploader
	logger          *logrus.Logger
}

// Upload a file with a given path to AWS S3
func (s *s3Uploader) Upload(path string) error {
	file, err := os.Open(path)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		s.logger.WithFields(logrus.Fields{
			"filename": path,
			"error":    err.Error(),
		}).Warnf(
			"Tried uploading file that does not exist, did another worker upload and then delete it?: %s: %w",
			path,
			err,
		)
		return err
	}
	if err != nil {
		s.logger.WithFields(logrus.Fields{
			"filename": path,
			"error":    err.Error(),
		}).Errorf("Failed to open file: %s: %w", path, err)
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

// NewS3Uploader creates a new uploader service for a given destination bucket
// in AWS S3
func NewS3Uploader(
	s3UploadManager S3ManagerUploader,
	toBucket string,
	logger *logrus.Logger,
) S3Uploader {
	return &s3Uploader{
		toBucket:        toBucket,
		s3UploadManager: s3UploadManager,
		logger:          logger,
	}
}
