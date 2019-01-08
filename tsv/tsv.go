package tsv

import "strings"

type StringBasedTabularData struct {
	content string
}

func (td *StringBasedTabularData) AppendRow(row []string) {
	td.content += strings.Join(row, "\t") + "\n"
}
func (td *StringBasedTabularData) Rows() [][]string {
	var rows [][]string
	for _, line := range strings.Split((td.content), "\n") {
		rows = append(rows, strings.Split(line, "\t"))
	}
	return rows
}

func (td *StringBasedTabularData) Empty() bool {
	return td.content == ""
}

type TextFileLoader interface {
	Upload(content string)
	Download() string
}

type TextFileBackedTabularData struct {
	StringBasedTabularData
	TextFileLoader TextFileLoader
}

func (td *TextFileBackedTabularData) AppendRow(row []string) {
	td.cacheContent()
	td.StringBasedTabularData.AppendRow(row)
	td.TextFileLoader.Upload(td.content)
}
func (td *TextFileBackedTabularData) Rows() [][]string {
	td.cacheContent()
	return td.StringBasedTabularData.Rows()
}

func (td *TextFileBackedTabularData) Empty() bool {
	td.cacheContent()
	return td.StringBasedTabularData.Empty()
}

func (td *TextFileBackedTabularData) cacheContent() {
	if td.content == "" {
		td.content = td.TextFileLoader.Download()
	}
}
