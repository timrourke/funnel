package upload

import (
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/timrourke/funnel/s3"
	"github.com/timrourke/funnel/tpl"
	"io/ioutil"
	"os"
	"testing"
)

type stubS3ManagerUploader struct {
	inputsPassed         chan *s3manager.UploadInput
	expectedReturnValues chan *s3manager.UploadOutput
	expectedErrorValues  chan error
}

// Upload is a stubbed implementation of `s3manager.Uploader.Upload`
func (s *stubS3ManagerUploader) Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	s.inputsPassed <- input

	return <-s.expectedReturnValues, <-s.expectedErrorValues
}

func TestUploadFilesFromPathToBucket(t *testing.T) {
	Convey("Should upload single file", t, func(c C) {
		file, err := ioutil.TempFile("", "somefile")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(file.Name())

		stub := &stubS3ManagerUploader{
			inputsPassed:         make(chan *s3manager.UploadInput),
			expectedReturnValues: make(chan *s3manager.UploadOutput),
			expectedErrorValues:  make(chan error),
		}

		go func() {
			inputPassed := <-stub.inputsPassed
			stub.expectedReturnValues <- nil
			stub.expectedErrorValues <- nil
			c.So(*inputPassed.Key, ShouldEqual, file.Name())
		}()

		logger := logrus.New()

		s3Uploader := s3.NewS3Uploader(stub, "unimportant", logger)

		keyTemplate, err := tpl.NewKeyTemplate("{{ filePath }}", logger)
		if err != nil {
			t.Fatal(err)
		}

		uploader := NewUploader(false, false, 10, s3Uploader, keyTemplate, logger)

		err = uploader.UploadFilesFromPathToBucket([]string{file.Name()})

		c.So(err, ShouldBeNil)
	})

	Convey("Should upload a directory", t, func(c C) {
		dirname, err := ioutil.TempDir("", "somedir")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(dirname)

		expectedFilePath1 := dirname + "/somefile1"
		err = ioutil.WriteFile(expectedFilePath1, nil, 0644)
		if err != nil {
			t.Fatal(err)
		}

		expectedFilePath2 := dirname + "/somefile2"
		err = ioutil.WriteFile(expectedFilePath2, nil, 0644)
		if err != nil {
			t.Fatal(err)
		}

		stub := &stubS3ManagerUploader{
			inputsPassed:         make(chan *s3manager.UploadInput),
			expectedReturnValues: make(chan *s3manager.UploadOutput),
			expectedErrorValues:  make(chan error),
		}

		go func() {
			var uploadedKeys []string
			for input := range stub.inputsPassed {
				uploadedKeys = append(uploadedKeys, *input.Key)
				stub.expectedReturnValues <- nil
				stub.expectedErrorValues <- nil
			}

			c.So(uploadedKeys, ShouldContain, expectedFilePath1)
			c.So(uploadedKeys, ShouldContain, expectedFilePath2)
			c.So(len(uploadedKeys), ShouldEqual, 2)
		}()

		logger := logrus.New()

		s3Uploader := s3.NewS3Uploader(stub, "unimportant", logger)

		keyTemplate, err := tpl.NewKeyTemplate("{{ filePath }}", logger)
		if err != nil {
			t.Fatal(err)
		}

		uploader := NewUploader(false, false, 10, s3Uploader, keyTemplate, logger)

		err = uploader.UploadFilesFromPathToBucket([]string{dirname})

		c.So(err, ShouldBeNil)
	})

	Convey("Should fail if no file paths provided", t, func() {
		stub := &stubS3ManagerUploader{
			inputsPassed:         nil,
			expectedReturnValues: nil,
			expectedErrorValues:  nil,
		}

		logger := logrus.New()

		s3Uploader := s3.NewS3Uploader(stub, "unimportant", logger)

		keyTemplate, err := tpl.NewKeyTemplate("{{ filePath }}", logger)
		if err != nil {
			t.Fatal(err)
		}

		uploader := NewUploader(false, false, 10, s3Uploader, keyTemplate, logger)

		err = uploader.UploadFilesFromPathToBucket([]string{})

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "must provide at least one path to a file or directory to upload to AWS S3")
	})
}
