package upload

import (
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/timrourke/funnel/s3"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func (u *uploader) handlePending(
	pending chan *fileUploadJob,
	completed chan<- *fileUploadJob,
	failed chan<- *fileUploadJob,
) {
	for input := range pending {
		err := u.s3Uploader.Upload(input.path)
		if err == nil {
			completed <- input
			continue
		}

		input.errors = append(input.errors, err)

		if len(input.errors) < 5 {
			go func() {
				pending <- input
			}()
		} else {
			failed <- input
		}
	}
}

func (u *uploader) enqueueDirContents(dirPathToWatch string, pending chan *fileUploadJob, wg *sync.WaitGroup) {
	err := filepath.Walk(dirPathToWatch, func(path string, info os.FileInfo, err error) error {
		if dirPathToWatch == path {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		wg.Add(1)
		pending <- &fileUploadJob{
			path:      path,
			errors:    []error{},
			startedAt: time.Now(),
		}

		return nil
	})
	if err != nil {
		u.logger.Fatal(err)
	}
}

func (u *uploader) uploadDir(filePath string, pending chan *fileUploadJob, wg *sync.WaitGroup) {
	if u.shouldWatchPaths {
		for {
			u.enqueueDirContents(filePath, pending, wg)

			time.Sleep(1 * time.Second)
		}
	} else {
		u.enqueueDirContents(filePath, pending, wg)
	}
}

func (u *uploader) processSingleFilePath(filePath string, pending chan *fileUploadJob, wg *sync.WaitGroup) error {
	filePathInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	if filePathInfo.IsDir() {
		u.uploadDir(filePath, pending, wg)
	} else {
		if u.shouldWatchPaths {
			for {
				wg.Add(1)
				pending <- &fileUploadJob{
					path:      filePath,
					errors:    []error{},
					startedAt: time.Now(),
				}

				time.Sleep(1 * time.Second)
			}
		} else {
			wg.Add(1)
			pending <- &fileUploadJob{
				path:      filePath,
				errors:    []error{},
				startedAt: time.Now(),
			}
		}
	}

	return nil
}

func (u *uploader) processMultipleFilePaths(filePaths []string, pending chan *fileUploadJob, wg *sync.WaitGroup) error {
	for _, filePath := range filePaths {
		filePathInfo, err := os.Stat(filePath)
		if err != nil {
			return err
		}

		if filePathInfo.IsDir() {
			u.uploadDir(filePath, pending, wg)
		} else {
			wg.Add(1)
			pending <- &fileUploadJob{
				path:      filePath,
				errors:    []error{},
				startedAt: time.Now(),
			}
		}
	}

	return nil
}

type fileUploadJob struct {
	path      string
	errors    []error
	startedAt time.Time
}

type Uploader interface {
	UploadFilesFromPathToBucket(filePaths []string) error
}

type uploader struct {
	logger               *logrus.Logger
	numConcurrentUploads int
	shouldWatchPaths     bool
	s3Uploader           s3.S3Uploader
}

func NewUploader(shouldWatchPaths bool, numConcurrentUploads int, s3Uploader s3.S3Uploader, logger *logrus.Logger) Uploader {
	return &uploader{
		logger:               logger,
		numConcurrentUploads: numConcurrentUploads,
		shouldWatchPaths:     shouldWatchPaths,
		s3Uploader:           s3Uploader,
	}
}

func (u *uploader) UploadFilesFromPathToBucket(filePaths []string) error {
	if 0 == len(filePaths) {
		return errors.New("must provide at least one path to a file or directory to upload to AWS S3")
	}

	var wg sync.WaitGroup

	pending, completed, failed := make(chan *fileUploadJob), make(chan *fileUploadJob), make(chan *fileUploadJob)

	for i := 0; i < u.numConcurrentUploads; i++ {
		go u.handlePending(pending, completed, failed)
	}

	go func() {
		for output := range completed {
			now := time.Now()
			uploadDuration := now.Sub(output.startedAt)

			u.logger.WithFields(logrus.Fields{
				"filename":            output.path,
				"startedAt":           output.startedAt.Format(time.RFC3339),
				"completedAt":         now.Format(time.RFC3339),
				"durationPretty":      uploadDuration.String(),
				"durationNanoseconds": uploadDuration.Nanoseconds(),
			}).Info(fmt.Sprintf("Uploaded file %s", output.path))

			wg.Done()
		}
	}()

	go func() {
		for failure := range failed {
			now := time.Now()
			failedDuration := now.Sub(failure.startedAt)
			var errorStrings []string

			for _, err := range failure.errors {
				errorStrings = append(errorStrings, err.Error())
			}

			u.logger.WithFields(logrus.Fields{
				"filename":            failure.path,
				"startedAt":           failure.startedAt.Format(time.RFC3339),
				"failedAt":            now.Format(time.RFC3339),
				"durationPretty":      failedDuration.String(),
				"durationNanoseconds": failedDuration.Nanoseconds(),
				"errors":              errorStrings,
			}).Info(fmt.Sprintf("Failed to upload file %s", failure.path))

			wg.Done()
		}
	}()

	if 1 == len(filePaths) {
		err := u.processSingleFilePath(filePaths[0], pending, &wg)
		if err != nil {
			return err
		}
	} else {
		if u.shouldWatchPaths {
			return errors.New("watching multiple paths not supported")
		}

		err := u.processMultipleFilePaths(filePaths, pending, &wg)
		if err != nil {
			return err
		}
	}

	wg.Wait()

	return nil
}
