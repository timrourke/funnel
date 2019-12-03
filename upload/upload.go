package upload

import (
	"errors"
	"github.com/timrourke/funnel/s3"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func handlePending(
	pending chan *fileUploadJob,
	completed chan<- *fileUploadJob,
	failed chan<- *fileUploadJob,
	uploader s3.S3Uploader,
) {
	for input := range pending {
		err := uploader.Upload(input.path)
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

func enqueueDirContents(dirPathToWatch string, pending chan *fileUploadJob, wg *sync.WaitGroup) {
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
		log.Fatal(err)
	}
}

func uploadDir(filePath string, pending chan *fileUploadJob, shouldWatchPaths bool, wg *sync.WaitGroup) {
	if shouldWatchPaths {
		for {
			enqueueDirContents(filePath, pending, wg)

			time.Sleep(1 * time.Second)
		}
	} else {
		enqueueDirContents(filePath, pending, wg)
	}
}

func processSingleFilePath(filePath string, pending chan *fileUploadJob, shouldWatchPaths bool, wg *sync.WaitGroup) error {
	filePathInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	if filePathInfo.IsDir() {
		uploadDir(filePath, pending, shouldWatchPaths, wg)
	} else {
		if shouldWatchPaths {
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

func processMultipleFilePaths(filePaths []string, pending chan *fileUploadJob, shouldWatchPaths bool, wg *sync.WaitGroup) error {
	for _, filePath := range filePaths {
		filePathInfo, err := os.Stat(filePath)
		if err != nil {
			return err
		}

		if filePathInfo.IsDir() {
			uploadDir(filePath, pending, shouldWatchPaths, wg)
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
	path        string
	errors      []error
	startedAt   time.Time
}

func UploadFilesFromPathToBucket(filePaths []string, shouldWatchPaths bool, uploader s3.S3Uploader) error {
	if 0 == len(filePaths) {
		return errors.New("must provide at least one path to a file or directory to upload to AWS S3")
	}

	var wg sync.WaitGroup

	pending, completed, failed := make(chan *fileUploadJob), make(chan *fileUploadJob), make(chan *fileUploadJob)

	for i := 0; i < 20; i++ {
		go handlePending(pending, completed, failed, uploader)
	}

	go func() {
		for output := range completed {
			uploadDuration := time.Now().Sub(output.startedAt)
			log.Printf("Uploaded file %s, took: %v", output.path, uploadDuration)
			wg.Done()
		}
	}()

	go func() {
		for failure := range failed {
			failedDuration := time.Now().Sub(failure.startedAt)
			log.Printf("Failed to upload file %s with errors %s, took: %v", failure.path, failure.errors, failedDuration)
			wg.Done()
		}
	}()

	if 1 == len(filePaths) {
		err := processSingleFilePath(filePaths[0], pending, shouldWatchPaths, &wg)
		if err != nil {
			return err
		}
	} else {
		if shouldWatchPaths {
			return errors.New("watching multiple paths not supported")
		}

		err := processMultipleFilePaths(filePaths, pending, shouldWatchPaths, &wg)
		if err != nil {
			return err
		}
	}

	wg.Wait()

	return nil
}
