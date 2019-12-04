package tpl

import (
	"bytes"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"sync"
	"text/template"
)

// KeyTemplate generates keys for S3 objects based on parsing of a template
type KeyTemplate interface {
	KeyForFile(relativeFilePath string) (string, error)
}

type keyTemplate struct {
	mux         sync.Mutex
	tplFileData *tplFileData
	template    *template.Template
}

// NewKeyTemplate creates an instance of a KeyTemplate
func NewKeyTemplate(templateText string, logger *logrus.Logger) (KeyTemplate, error) {
	keyTemplate := &keyTemplate{mux: sync.Mutex{}}

	funcMap := template.FuncMap{
		"absoluteFilePath": func() string {
			abspath, err := keyTemplate.tplFileData.AbsoluteFilePath()
			if err != nil {
				panic(fmt.Errorf(
					"failed to parse absolute path for file: %s: %w",
					keyTemplate.tplFileData.filePath,
					err,
				))
			}

			return abspath
		},
		"dateWithFormat": func(layout string) string {
			return keyTemplate.tplFileData.DateWithFormat(layout)
		},
		"fileExtension": func() string {
			return keyTemplate.tplFileData.FileExtension()
		},
		"fileName": func() string {
			return keyTemplate.tplFileData.FileName()
		},
		"fileNameWithoutExtension": func() string {
			return keyTemplate.tplFileData.FileNameWithoutExtension()
		},
		"filePath": func() string {
			return keyTemplate.tplFileData.RelativeFilePath()
		},
	}

	tmpl, err := template.New("key").Funcs(funcMap).Parse(templateText)
	if err != nil {
		logger.Errorf("failed to parse template text: %w", err)
		return nil, err
	}

	keyTemplate.template = tmpl

	return keyTemplate, nil
}

// KeyForFile takes the path provided by the caller and parses the template with
// the provided path as template context
func (k *keyTemplate) KeyForFile(relativeFilePath string) (string, error) {
	info, err := os.Stat(relativeFilePath)
	if err != nil {
		panic(fmt.Errorf("failed to stat file: %s: %w", relativeFilePath, err))
	}

	k.mux.Lock()
	defer k.mux.Unlock()
	k.tplFileData = &tplFileData{
		filePath: relativeFilePath,
		fileInfo: info,
	}

	var b bytes.Buffer
	err = k.template.Execute(&b, k.tplFileData)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}
