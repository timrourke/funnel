package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"os"
	"testing"
)

var (
	fixture1                     *os.File
	fixture2                     *os.File
	fixture3                     *os.File
	fixture4                     *os.File
	fixtureDir1                  string
	funnelTestAwsAccessKeyId     = os.Getenv("FUNNEL_TEST_AWS_ACCESS_KEY_ID")
	funnelTestAwsDefaultRegion   = os.Getenv("FUNNEL_TEST_AWS_DEFAULT_REGION")
	funnelTestAwsSecretAccessKey = os.Getenv("FUNNEL_TEST_AWS_SECRET_ACCESS_KEY")
	funnelTestAwsS3Bucket        = os.Getenv("FUNNEL_TEST_AWS_S3_BUCKET")
	s3Client                     *s3.S3
)

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	os.Exit(code)
}

func setup() {
	creds := credentials.NewStaticCredentials(
		funnelTestAwsAccessKeyId,
		funnelTestAwsSecretAccessKey,
		"",
	)

	config := aws.NewConfig().
		WithRegion(funnelTestAwsDefaultRegion).
		WithCredentials(creds)

	sess := session.Must(session.NewSession(config))

	s3Client = s3.New(sess)

	resetCliFlags()
	createFixtures()
}

func resetCliFlags() {
	bucket = ""
	numConcurrentUploads = 0
	region = ""
}

func cleanUpBucket() {
	deleteAllObjsInBucket(
		&s3.ListObjectsInput{
			Bucket: aws.String(funnelTestAwsS3Bucket),
		},
		s3Client,
	)
}

func createFixtures() {
	file, err := ioutil.TempFile(os.TempDir(), "somefile.txt")
	if err != nil {
		panic(err)
	}

	fixture1 = file

	file, err = ioutil.TempFile(os.TempDir(), "somefile2.txt")
	if err != nil {
		panic(err)
	}

	fixture2 = file

	file, err = ioutil.TempFile(os.TempDir(), "somefile3.txt")
	if err != nil {
		panic(err)
	}

	fixture3 = file

	dir, err := ioutil.TempDir(os.TempDir(), "somedir")
	if err != nil {
		panic(err)
	}

	fixtureDir1 = dir

	file, err = ioutil.TempFile(dir, "somefileindir1.txt")
	if err != nil {
		panic(err)
	}

	fixture4 = file
}

func deleteAllObjsInBucket(listObjectsInput *s3.ListObjectsInput, s3Client *s3.S3) {
	objectListing, err := s3Client.ListObjects(listObjectsInput)
	if err != nil {
		panic(err)
	}

	for _, obj := range objectListing.Contents {
		_, err := s3Client.DeleteObject(&s3.DeleteObjectInput{
			Bucket: aws.String(funnelTestAwsS3Bucket),
			Key:    obj.Key,
		})
		if err != nil {
			panic(err)
		}
	}

	if *objectListing.IsTruncated {
		deleteAllObjsInBucket(
			&s3.ListObjectsInput{
				Bucket: aws.String(funnelTestAwsS3Bucket),
				Marker: objectListing.Marker,
			},
			s3Client,
		)
	}
}

func funnelTestBucketContainsObjectWithKey(key string) bool {
	_, err := s3Client.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(funnelTestAwsS3Bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		logger.Error(err)
		return false
	}

	return true
}

func TestExecute(t *testing.T) {
	Convey("Uploading files", t, func() {
		Convey("Should upload a single file", func() {
			cleanUpBucket()
			defer resetCliFlags()

			bucket = funnelTestAwsS3Bucket
			numConcurrentUploads = 10
			region = funnelTestAwsDefaultRegion

			err := Execute(rootCmd, []string{fixture1.Name()})
			if err != nil {
				t.Fatal(err)
			}

			So(funnelTestBucketContainsObjectWithKey(fixture1.Name()), ShouldBeTrue)
		})

		Convey("Should upload multiple files", func() {
			cleanUpBucket()
			defer resetCliFlags()

			bucket = funnelTestAwsS3Bucket
			numConcurrentUploads = 10
			region = funnelTestAwsDefaultRegion

			err := Execute(rootCmd, []string{fixture2.Name(), fixture3.Name()})
			if err != nil {
				t.Fatal(err)
			}

			So(funnelTestBucketContainsObjectWithKey(fixture2.Name()), ShouldBeTrue)
			So(funnelTestBucketContainsObjectWithKey(fixture3.Name()), ShouldBeTrue)
		})

		Convey("Should upload directory with a single file in it", func() {
			cleanUpBucket()
			defer resetCliFlags()

			bucket = funnelTestAwsS3Bucket
			numConcurrentUploads = 10
			region = funnelTestAwsDefaultRegion

			err := Execute(rootCmd, []string{fixtureDir1})
			if err != nil {
				t.Fatal(err)
			}

			So(funnelTestBucketContainsObjectWithKey(fixture4.Name()), ShouldBeTrue)
		})
	})

	Convey("Command line flag validation", t, func() {
		Convey("Should fail if region is empty", func() {
			defer resetCliFlags()

			err := Execute(rootCmd, []string{})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "must provide an AWS region where your S3 bucket exists")
		})

		Convey("Should fail if region is just whitespace", func() {
			defer resetCliFlags()

			region = `

  `

			err := Execute(rootCmd, []string{})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "must provide an AWS region where your S3 bucket exists")
		})

		Convey("Should fail if bucket is empty", func() {
			defer resetCliFlags()

			region = "us-east-1"

			err := Execute(rootCmd, []string{})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "must specify an AWS S3 bucket to save files in")
		})

		Convey("Should fail if bucket is just whitespace", func() {
			defer resetCliFlags()

			region = "us-east-1"
			bucket = `

  `

			err := Execute(rootCmd, []string{})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "must specify an AWS S3 bucket to save files in")
		})

		Convey("Should fail if numConcurrentUploads is less than zero", func() {
			defer resetCliFlags()

			region = "us-east-1"
			bucket = "unimportant"
			numConcurrentUploads = -1

			err := Execute(rootCmd, []string{})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "number of concurrent uploads must be within the range 1-100")
		})

		Convey("Should fail if numConcurrentUploads is zero", func() {
			defer resetCliFlags()

			region = "us-east-1"
			bucket = "unimportant"
			numConcurrentUploads = 0

			err := Execute(rootCmd, []string{})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "number of concurrent uploads must be within the range 1-100")
		})

		Convey("Should fail if numConcurrentUploads is greater than 100", func() {
			defer resetCliFlags()

			region = "us-east-1"
			bucket = "unimportant"
			numConcurrentUploads = 101

			err := Execute(rootCmd, []string{})

			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "number of concurrent uploads must be within the range 1-100")
		})
	})
}
