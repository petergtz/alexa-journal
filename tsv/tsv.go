package tsv

import (
	"strings"

	"github.com/pkg/errors"
)

type StringBasedTabularData struct {
	content string
}

func (td *StringBasedTabularData) AppendRow(row []string) error {
	td.content += strings.Join(row, "\t") + "\n"
	return nil
}
func (td *StringBasedTabularData) Rows() ([][]string, error) {
	var rows [][]string
	for _, line := range strings.Split((td.content), "\n") {
		rows = append(rows, strings.Split(line, "\t"))
	}
	return rows, nil
}

func (td *StringBasedTabularData) Empty() (bool, error) {
	return td.content == "", nil
}

type TextFileLoader interface {
	Upload(content string) error
	Download() (string, error)
}

type TextFileBackedTabularData struct {
	StringBasedTabularData
	TextFileLoader TextFileLoader
}

func (td *TextFileBackedTabularData) AppendRow(row []string) error {
	if e := td.cacheContent(); e != nil {
		return e
	}
	td.StringBasedTabularData.AppendRow(row)
	e := td.TextFileLoader.Upload(td.content)
	if e != nil {
		return errors.Wrap(e, "Could not upload file content")
	}
	return nil
}
func (td *TextFileBackedTabularData) Rows() ([][]string, error) {
	if e := td.cacheContent(); e != nil {
		return nil, e
	}
	return td.StringBasedTabularData.Rows()
}

func (td *TextFileBackedTabularData) Empty() (bool, error) {
	if e := td.cacheContent(); e != nil {
		return false, e
	}
	return td.StringBasedTabularData.Empty()
}

func (td *TextFileBackedTabularData) cacheContent() error {
	if td.content == "" {
		var e error
		td.content, e = td.TextFileLoader.Download()
		if e != nil {
			return errors.Wrap(e, "Could not download file content")
		}
	}
	return nil
}
