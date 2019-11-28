package main

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"github.com/timrourke/funnel/s3"
	"github.com/timrourke/funnel/upload"
	"os"
	"strings"
)

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
var shouldWatchPaths bool

var rootCmd = &cobra.Command{
	Use:                        "funnel [OPTIONS] [PATHS]",
	Short:                      "Funnel is a tool for quickly saving files to AWS S3.",
	Example:                    "funnel --region=us-east-1 --bucket=some-cool-bucket /some/directory",
	Version:                    "0.0.1",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateCommandLineFlags(); err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		uploader := s3.NewS3Uploader(region, bucket)

		return upload.UploadFilesFromPathToBucket(args, shouldWatchPaths, uploader)
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
		&shouldWatchPaths,
		"watch",
		"w",
		false,
		"Whether to watch a path for changes",
	)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println("Upload failed:", err)
		os.Exit(1)
	}
}
