package main

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/timrourke/funnel/s3"
	"log"
	"os"
	"path/filepath"
	"strings"
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

func watchDir(dirPathToWatch string, pending chan string) {
	err := filepath.Walk(dirPathToWatch, func(path string, info os.FileInfo, err error) error {
		if dirPathToWatch == path {
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

func validateCommandLineFlags() error {
	if "" == strings.Trim(region, " ") {
		return errors.New("must provide an AWS region where your S3 bucket exists")
	}

	if "" == strings.Trim(bucket, " ") {
		return errors.New("must specify an AWS S3 bucket to save files in")
	}

	return nil
}

var bucket string
var region string
var watch bool

var rootCmd = &cobra.Command{
	Use:                        "funnel",
	Short:                      "Funnel is a tool for quickly saving files in a given directory to AWS S3.",
	Example:                    "funnel --region=us-east-1 --bucket=some-cool-bucket",
	Version:                    "0.0.1",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateCommandLineFlags(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		if 0 == len(args) {
			return errors.New("must provide at least one path to a file or directory to upload to AWS S3")
		}

		uploader := s3.NewS3Uploader(region, bucket)

		pending, completed := make(chan string), make(chan string)

		for i := 0; i < 10; i++ {
			go handlePending(pending, completed, uploader)
		}

		go func() {
			for output := range completed {
				log.Println("Processed:", output)
			}
		}()

		if 1 == len(args) {
			info, err := os.Stat(args[0])
			if err != nil {
				return err
			}

			if info.IsDir() {
				if watch {
					for {
						watchDir(args[0], pending)

						time.Sleep(1 * time.Second)
					}
				} else {
					watchDir(args[0], pending)
				}
			} else {
				if watch {
					return errors.New("watching a single file not supported")
				}

				return uploader.Upload(args[0])
			}
		}

		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(
		&region,
		"region",
		"r",
		"",
		"The AWS region your S3 bucket is in, eg. \"us-east-1\"",
	)

	rootCmd.PersistentFlags().StringVarP(
		&bucket,
		"bucket",
		"b",
		"",
		"The AWS S3 bucket you want to save files to",
	)

	rootCmd.PersistentFlags().BoolVarP(
		&watch,
		"watch",
		"w",
		false,
		"Whether to watch a path for changes",
	)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

