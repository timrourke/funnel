package main

import (
	"errors"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/timrourke/funnel/s3"
	"github.com/timrourke/funnel/upload"
	"golang.org/x/crypto/ssh/terminal"
	"os"
	"strings"
)

func validateCommandLineFlags() error {
	if "" == strings.TrimSpace(region) {
		return errors.New("must provide an AWS region where your S3 bucket exists")
	}

	if "" == strings.TrimSpace(bucket) {
		return errors.New("must specify an AWS S3 bucket to save files in")
	}

	if numConcurrentUploads <= 0 || numConcurrentUploads > 100 {
		return errors.New("number of concurrent uploads must be within the range 1-100")
	}

	return nil
}

var (
	bucket                      string
	log                         = logrus.New()
	numConcurrentUploads        int
	shouldDeleteFileAfterUpload bool
	shouldWatchPaths            bool
	region                      string

	rootCmd = &cobra.Command{
		Use:     "funnel [OPTIONS] [PATHS]",
		Short:   "Funnel is a tool for quickly saving files to AWS S3.",
		Example: "funnel --region=us-east-1 --bucket=some-cool-bucket /some/directory",
		Version: "0.0.1",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := validateCommandLineFlags()
			if err != nil {
				return err
			}

			config := aws.NewConfig().
				WithRegion(region).
				WithMaxRetries(3)

			sess := session.Must(session.NewSession(config))

			s3UploadManager := s3manager.NewUploader(sess)

			s3Uploader := s3.NewS3Uploader(s3UploadManager, bucket, log)

			uploader := upload.NewUploader(
				shouldDeleteFileAfterUpload,
				shouldWatchPaths,
				numConcurrentUploads,
				s3Uploader,
				log,
			)

			return uploader.UploadFilesFromPathToBucket(args)
		},
	}
)

func configureLogger() {
	if !terminal.IsTerminal(int(os.Stdout.Fd())) {
		log.SetFormatter(&logrus.JSONFormatter{})
	}
}

func configureRootCmd() {
	rootCmd.PersistentFlags().StringVarP(
		&region,
		"region",
		"r",
		"",
		"The AWS region your S3 bucket is in, eg. \"us-east-1\"",
	)
	if "" == strings.TrimSpace(region) {
		region = os.Getenv("AWS_DEFAULT_REGION")
	}

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

	rootCmd.PersistentFlags().IntVarP(
		&numConcurrentUploads,
		"num-concurrent-uploads",
		"n",
		10,
		"Number of concurrent uploads",
	)

	rootCmd.PersistentFlags().BoolVarP(
		&shouldDeleteFileAfterUpload,
		"delete-file-after-upload",
		"",
		false,
		"Whether to delete the uploaded file after a successful upload",
	)

	rootCmd.DisableFlagsInUseLine = true
}

func init() {
	configureLogger()
	configureRootCmd()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
