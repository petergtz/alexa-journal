package journaldrive

import (
	"context"
	"io/ioutil"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/rickb777/date"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
)

type Journal struct {
	files    *drive.FilesService
	filename string
}

type JournalProvider struct {
}

func (jp *JournalProvider) Get(accessToken string) *Journal {
	return NewJournal(accessToken, "my-journal.tsv")
}

func NewJournal(accessToken string, filename string) *Journal {
	d, e := drive.New(
		oauth2.NewClient(
			context.TODO(),
			oauth2.StaticTokenSource(&oauth2.Token{
				AccessToken: accessToken,
			})))
	if e != nil {
		panic(errors.Wrap(e, "Could not instantiate drive"))
	}
	return &Journal{files: d.Files, filename: filename}
}

func (j *Journal) AddEntry(entryDate date.Date, text string) error {
	fileList, e := j.files.List().Q("name = '" + j.filename + "' and trashed = false").Do()
	if e != nil {
		return errors.Wrap(e, "Could not list files")
	}
	fileId := ""
	content := ""
	switch len(fileList.Files) {
	case 0:
		file, e := j.files.Create(&drive.File{Name: j.filename}).Media(strings.NewReader("timestamp\tdate\ttext\n")).Do()
		if e != nil {
			return errors.Wrap(e, "Could not create file to store journal")
		}
		fileId = file.Id
	case 1:
		fileId = fileList.Files[0].Id
		download, e := j.files.Get(fileId).Download()
		if e != nil {
			return errors.Wrap(e, "Could not download file")
		}
		defer download.Body.Close()

		byteContent, e := ioutil.ReadAll(download.Body)
		if e != nil {
			return errors.Wrap(e, "Could not read body")
		}
		content = string(byteContent)
	default:
		return errors.New("Expected exactly 0 or 1 file in drive")
	}

	if string(content) == "" {
		content += "timestamp\tdate\ttext\n"
	}
	content += time.Now().Format("2006-01-02 15:04:05 -0700 MST") + "\t" + entryDate.String() + "\t" + text + "\n"

	_, e = j.files.Update(fileId, &drive.File{}).Media(strings.NewReader(content)).Do()
	if e != nil {
		return errors.Wrap(e, "Could not upload updated content")
	}

	return nil
}

func (j *Journal) GetEntry(entryDate date.Date) (string, error) {
	fileList, e := j.files.List().Q("name = '" + j.filename + "' and trashed = false").Do()
	if e != nil {
		return "", errors.Wrap(e, "Could not list files")
	}

	switch len(fileList.Files) {
	case 0:
		return "Keinen Eintrag gefunden", nil
	case 1:
		download, e := j.files.Get(fileList.Files[0].Id).Download()
		if e != nil {
			return "", errors.Wrap(e, "Could not download file")
		}
		defer download.Body.Close()

		content, e := ioutil.ReadAll(download.Body)
		if e != nil {
			return "", errors.Wrap(e, "Could not read body")
		}
		for _, line := range strings.Split(string(content), "\n") {
			parts := strings.Split(line, "\t")
			if len(parts) != 3 {
				continue
			}
			if parts[1] == entryDate.String() {
				return parts[2], nil
			}
		}
		return "Keinen Eintrag gefunden", nil
	default:
		return "", errors.New("Expected exactly 0 or 1 file in drive")
	}
}
