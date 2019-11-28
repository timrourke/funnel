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

func handlePending(pending <-chan string, completed chan<- string, uploader s3.S3Uploader) {
	for input := range pending {
		err := uploader.Upload(input)
		if err != nil {
			log.Println("Failed to upload file:", err)
		}
		completed <- input
	}
}

func enqueueDirContents(dirPathToWatch string, pending chan string, wg *sync.WaitGroup) {
	err := filepath.Walk(dirPathToWatch, func(path string, info os.FileInfo, err error) error {
		if dirPathToWatch == path {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		wg.Add(1)
		pending <- path

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func uploadDir(filePath string, pending chan string, shouldWatchPaths bool, wg *sync.WaitGroup) {
	if shouldWatchPaths {
		for {
			enqueueDirContents(filePath, pending, wg)

			time.Sleep(1 * time.Second)
		}
	} else {
		enqueueDirContents(filePath, pending, wg)
	}
}

func processSingleFilePath(filePath string, pending chan string, shouldWatchPaths bool, wg *sync.WaitGroup) error {
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
				pending <- filePath

				time.Sleep(1 * time.Second)
			}
		} else {
			wg.Add(1)
			pending <- filePath
		}
	}

	return nil
}

func processMultipleFilePaths(filePaths []string, pending chan string, shouldWatchPaths bool, wg *sync.WaitGroup) error {
	for _, filePath := range filePaths {
		filePathInfo, err := os.Stat(filePath)
		if err != nil {
			return err
		}

		if filePathInfo.IsDir() {
			uploadDir(filePath, pending, shouldWatchPaths, wg)
		} else {
			wg.Add(1)
			pending <- filePath
		}
	}

	return nil
}

func UploadFilesFromPathToBucket(filePaths []string, shouldWatchPaths bool, uploader s3.S3Uploader) error {
	if 0 == len(filePaths) {
		return errors.New("must provide at least one path to a file or directory to upload to AWS S3")
	}

	var wg sync.WaitGroup

	pending, completed := make(chan string), make(chan string)

	for i := 0; i < 10; i++ {
		go handlePending(pending, completed, uploader)
	}

	go func() {
		for output := range completed {
			log.Println("Processed:", output)
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
