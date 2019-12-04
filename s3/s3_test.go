package s3

import (
	"errors"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"os"
	"testing"
)

type stubS3ManagerUploader struct {
	inputsPassed         []*s3manager.UploadInput
	expectedReturnValues []*s3manager.UploadOutput
	expectedErrorValues  []error
}

// Upload is a stubbed implementation of `s3manager.Uploader.Upload`
func (s *stubS3ManagerUploader) Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	s.inputsPassed = append(s.inputsPassed, input)
	ret := s.expectedReturnValues[len(s.expectedReturnValues)-1]
	err := s.expectedErrorValues[len(s.expectedErrorValues)-1]
	s.expectedReturnValues = s.expectedReturnValues[:len(s.expectedReturnValues)-1]
	s.expectedErrorValues = s.expectedErrorValues[:len(s.expectedErrorValues)-1]
	return ret, err
}

func TestNewS3Uploader(t *testing.T) {
	Convey("Should create new uploader", t, func() {
		stub := &stubS3ManagerUploader{}
		uploader := NewS3Uploader(stub, "some-bucket", logrus.New())

		So(uploader, ShouldNotBeNil)
	})
}

func TestS3Uploader_Upload(t *testing.T) {
	Convey("Should upload existing path to S3", t, func() {
		expectedPath := "/dev/null"
		expectedBucket := "some-bucket"
		stub := &stubS3ManagerUploader{
			inputsPassed:         nil,
			expectedReturnValues: []*s3manager.UploadOutput{nil},
			expectedErrorValues:  []error{nil},
		}

		uploader := NewS3Uploader(stub, expectedBucket, logrus.New())

		err := uploader.Upload(expectedPath, expectedPath)

		So(err, ShouldBeNil)

		inputPassed := stub.inputsPassed[0]

		Convey("Should upload to correct bucket", func() {
			So(*inputPassed.Bucket, ShouldEqual, expectedBucket)
		})

		Convey("Should name S3 key after file path", func() {
			So(*inputPassed.Key, ShouldEqual, expectedPath)
		})

		Convey("Should close file after upload", func() {
			_, err = ioutil.ReadAll(inputPassed.Body)
			So(err, ShouldNotBeNil)
			So(err, ShouldHaveSameTypeAs, &os.PathError{})
			So(err.Error(), ShouldContainSubstring, "already closed")
		})
	})

	Convey("Should fail to upload nonexistent file path", t, func() {
		stub := &stubS3ManagerUploader{
			inputsPassed:         nil,
			expectedReturnValues: []*s3manager.UploadOutput{nil},
			expectedErrorValues:  []error{nil},
		}

		uploader := NewS3Uploader(stub, "some-bucket", logrus.New())

		err := uploader.Upload("a nonexistent path", "unimportant")

		So(err, ShouldNotBeNil)
		So(err, ShouldHaveSameTypeAs, &os.PathError{})
		So(err.Error(), ShouldContainSubstring, "no such file or directory")
	})

	Convey("Should communicate upload failure", t, func() {
		expectedError := errors.New("unimportant")

		stub := &stubS3ManagerUploader{
			inputsPassed:         nil,
			expectedReturnValues: []*s3manager.UploadOutput{nil},
			expectedErrorValues:  []error{expectedError},
		}

		uploader := NewS3Uploader(stub, "some-bucket", logrus.New())

		err := uploader.Upload("/dev/null", "unimportant")

		So(err, ShouldEqual, expectedError)
	})
}
