package main

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func resetCliFlags() {
	bucket = ""
	numConcurrentUploads = 0
	region = ""
}

func TestExecute(t *testing.T) {
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
