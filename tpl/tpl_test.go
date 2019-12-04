package tpl

import (
	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"os"
	"path"
	"regexp"
	"testing"
)

var logger logrus.Logger

func TestNewKeyTemplate(t *testing.T) {
	Convey("should fail if unable to parse template text", t, func() {
		_, err := NewKeyTemplate("{{", &logger)

		So(err, ShouldNotBeNil)
	})

	Convey("should instantiate if template is valid", t, func() {
		tpl, err := NewKeyTemplate("something valid", &logger)

		So(err, ShouldBeNil)
		So(tpl, ShouldNotBeNil)
	})
}

func TestKeyTemplate_KeyForFile(t *testing.T) {
	tempFile, err := os.Create(path.Join(os.TempDir(), "somefile.go"))
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tempFile.Name())

	Convey("Template functions", t, func() {
		Convey("should interpolate file's abs path", func() {
			tpl, err := NewKeyTemplate("{{ absoluteFilePath }}", &logger)
			if err != nil {
				t.Fatal(err)
			}

			actual, err := tpl.KeyForFile(tempFile.Name())
			if err != nil {
				t.Errorf("failed to interpolate abs path: %v", err)
			}

			So(actual, ShouldEqual, path.Join(os.TempDir(), "somefile.go"))
		})

		Convey("should interpolate formatted date", func() {
			tpl, err := NewKeyTemplate(`{{ dateWithFormat "2006-01-02" }}`, &logger)
			if err != nil {
				t.Fatal(err)
			}

			actual, err := tpl.KeyForFile(tempFile.Name())
			if err != nil {
				t.Errorf("failed to interpolate date with format: %v", err)
			}

			// Specific date is unimportant here
			matches, err := regexp.MatchString(`[\d]{4}-[\d]{2}-[\d]{2}`, actual)
			if err != nil {
				t.Fatal(err)
			}

			So(matches, ShouldBeTrue)
		})

		Convey("should interpolate file's extension", func() {
			tpl, err := NewKeyTemplate("{{ fileExtension }}", &logger)
			if err != nil {
				t.Fatal(err)
			}

			actual, err := tpl.KeyForFile(tempFile.Name())
			if err != nil {
				t.Errorf("failed to interpolate extension: %v", err)
			}

			So(actual, ShouldEqual, ".go")
		})

		Convey("should interpolate file's name", func() {
			tpl, err := NewKeyTemplate("{{ fileName }}", &logger)
			if err != nil {
				t.Fatal(err)
			}

			actual, err := tpl.KeyForFile(tempFile.Name())
			if err != nil {
				t.Errorf("failed to interpolate filename: %v", err)
			}

			So(actual, ShouldEqual, "somefile.go")
		})

		Convey("should interpolate file's name without extension", func() {
			tpl, err := NewKeyTemplate("{{ fileNameWithoutExtension }}", &logger)
			if err != nil {
				t.Fatal(err)
			}

			actual, err := tpl.KeyForFile(tempFile.Name())
			if err != nil {
				t.Errorf("failed to interpolate file name without extension: %v", err)
			}

			So(actual, ShouldEqual, "somefile")
		})

		Convey("should interpolate file's path as originally provided", func() {
			tpl, err := NewKeyTemplate("{{ filePath }}", &logger)
			if err != nil {
				t.Fatal(err)
			}

			actual, err := tpl.KeyForFile(tempFile.Name())
			if err != nil {
				t.Errorf("failed to interpolate file path as originally provided: %v", err)
			}

			So(actual, ShouldEqual, path.Join(os.TempDir(), "somefile.go"))
		})
	})
}
