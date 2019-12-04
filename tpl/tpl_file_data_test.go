package tpl

import (
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"path"
	"testing"
	"time"
)

func TestTplFileData_AbsoluteFilePath(t *testing.T) {
	tempDir := os.TempDir()
	tempFile, err := os.Create(path.Join(os.TempDir(), "somefile.txt"))
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	fileInfo, err := os.Stat(tempFile.Name())
	if err != nil {
		t.Fatal(err)
	}

	Convey("AbsoluteFilePath", t, func() {
		expectedAbsolutePath := tempDir + "/somefile.txt"

		tplFileData := &tplFileData{
			filePath: tempFile.Name(),
			fileInfo: fileInfo,
		}

		actual, err := tplFileData.AbsoluteFilePath()

		So(err, ShouldBeNil)
		So(actual, ShouldEqual, expectedAbsolutePath)
	})

	Convey("DateWithFormat", t, func() {
		tplFileData := &tplFileData{
			filePath: tempFile.Name(),
			fileInfo: fileInfo,
		}

		// Go uses a reference date to perform date formatting
		// @see https://golang.org/pkg/time/#Time.Format
		actual := tplFileData.DateWithFormat("2006-01-02")

		So(actual, ShouldEqual, time.Now().Format("2006-01-02"))
	})

	Convey("FileExtension", t, func() {
		tplFileData := &tplFileData{
			filePath: tempFile.Name(),
			fileInfo: fileInfo,
		}

		actual := tplFileData.FileExtension()

		So(actual, ShouldEqual, ".txt")
	})

	Convey("FileName", t, func() {
		tplFileData := &tplFileData{
			filePath: tempFile.Name(),
			fileInfo: fileInfo,
		}

		actual := tplFileData.FileName()

		So(actual, ShouldEqual, "somefile.txt")
	})

	Convey("FileNameWithoutExtension", t, func() {
		tplFileData := &tplFileData{
			filePath: tempFile.Name(),
			fileInfo: fileInfo,
		}

		actual := tplFileData.FileNameWithoutExtension()

		So(actual, ShouldEqual, "somefile")
	})

	Convey("RelativeFilePath", t, func() {
		tplFileData := &tplFileData{
			filePath: tempFile.Name(),
			fileInfo: fileInfo,
		}

		actual := tplFileData.RelativeFilePath()

		So(actual, ShouldEqual, tempFile.Name())
	})
}
