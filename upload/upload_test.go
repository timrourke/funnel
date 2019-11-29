package upload

import (
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/timrourke/funnel/s3"
	"io/ioutil"
	"os"
	"testing"
)

type stubS3ManagerUploader struct {
	inputsPassed []*s3manager.UploadInput
	expectedReturnValues []*s3manager.UploadOutput
	expectedErrorValues []error
}

func (s *stubS3ManagerUploader) Upload(input *s3manager.UploadInput, options ...func(*s3manager.Uploader)) (*s3manager.UploadOutput, error) {
	s.inputsPassed = append(s.inputsPassed, input)
	ret := s.expectedReturnValues[len(s.expectedReturnValues)-1]
	err := s.expectedErrorValues[len(s.expectedErrorValues)-1]
	s.expectedReturnValues = s.expectedReturnValues[:len(s.expectedReturnValues)-1]
	s.expectedErrorValues = s.expectedErrorValues[:len(s.expectedErrorValues)-1]
	return ret, err
}

func TestUploadFilesFromPathToBucket(t *testing.T) {
	Convey("Should upload single file", t, func() {
		file, err := ioutil.TempFile("", "somefile")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(file.Name())

		stub := &stubS3ManagerUploader{
			inputsPassed:         []*s3manager.UploadInput{},
			expectedReturnValues: []*s3manager.UploadOutput{nil},
			expectedErrorValues:  []error{nil},
		}

		uploader := s3.NewS3Uploader(stub, "unimportant")

		err = UploadFilesFromPathToBucket([]string{file.Name()}, false, uploader)

		So(err, ShouldBeNil)
		So(len(stub.inputsPassed), ShouldEqual, 1)
		So(*stub.inputsPassed[0].Key, ShouldEqual, file.Name())
	})

	Convey("Should upload a directory", t, func() {
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
			inputsPassed:         []*s3manager.UploadInput{},
			expectedReturnValues: []*s3manager.UploadOutput{nil, nil},
			expectedErrorValues:  []error{nil, nil},
		}

		uploader := s3.NewS3Uploader(stub, "unimportant")

		err = UploadFilesFromPathToBucket([]string{dirname}, false, uploader)

		So(err, ShouldBeNil)
		So(len(stub.inputsPassed), ShouldEqual, 2)

		var uploadedKeys []string
		for _, input := range stub.inputsPassed {
			uploadedKeys = append(uploadedKeys, *input.Key)
		}

		So(uploadedKeys, ShouldContain, expectedFilePath1)
		So(uploadedKeys, ShouldContain, expectedFilePath2)
	})

	Convey("Should fail if no file paths provided", t, func() {
		stub := &stubS3ManagerUploader{
			inputsPassed:         nil,
			expectedReturnValues: nil,
			expectedErrorValues:  nil,
		}

		uploader := s3.NewS3Uploader(stub, "unimportant")

		err := UploadFilesFromPathToBucket([]string{}, false, uploader)

		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldEqual, "must provide at least one path to a file or directory to upload to AWS S3")
	})
}
