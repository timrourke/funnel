package main

import (
	"github.com/timrourke/funnel/s3"
	"log"
	"os"
	"path/filepath"
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

func watchDir(pending chan string) {
	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if "." == path {
			return nil
		}

		if info.IsDir() {
			return nil
		}

		pending <- path

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	uploader := s3.NewS3Uploader("us-west-2", "sdfhdfhsdfsdfh333")

	pending, completed := make(chan string), make(chan string)

	for i := 0; i < 10; i++ {
		go handlePending(pending, completed, uploader)
	}

	go func() {
		for output := range completed {
			log.Println("Processed:", output)
		}
	}()

	for {
		watchDir(pending)

		time.Sleep(1 * time.Second)
	}
}

