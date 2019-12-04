// Package tpl provides tools for customizing the key for the uploaded S3 objects
package tpl

import (
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"
)

// TplFileData exposes data to be used as template context when parsing templates
// for keys of S3 objects
type TplFileData interface {
	AbsoluteFilePath() (string, error)
	DateWithFormat(layout string) string
	FileExtension() string
	FileName() string
	FileNameWithoutExtension() string
	RelativeFilePath() string
}

type tplFileData struct {
	filePath string
	fileInfo os.FileInfo
}

// AbsoluteFilePath determines the absolute file path on your local computer
// and uses that for the key
func (t *tplFileData) AbsoluteFilePath() (string, error) {
	if filepath.IsAbs(t.filePath) {
		return t.filePath, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return filepath.Abs(path.Join(cwd, t.filePath))
}

// DateWithFormat formats the current time (eg. `time.Now()`) and formats it
// with the provided layout string
func (t *tplFileData) DateWithFormat(layout string) string {
	now := time.Now()
	return now.Format(layout)
}

// FileExtension returns the file extension, eg. `.txt`
func (t *tplFileData) FileExtension() string {
	return path.Ext(t.fileInfo.Name())
}

// FileName returns the whole filename, without preceding directories
func (t *tplFileData) FileName() string {
	return t.fileInfo.Name()
}

// FileNameWithoutExtension returns the filename, but without the file extension
func (t *tplFileData) FileNameWithoutExtension() string {
	ext := path.Ext(t.fileInfo.Name())

	return strings.TrimSuffix(t.fileInfo.Name(), ext)
}

// RelativeFilePath returns the unmodified path as stored under the field `tplFileData.filePath`
// TODO: Improve the name of this method
func (t *tplFileData) RelativeFilePath() string {
	return t.filePath
}
