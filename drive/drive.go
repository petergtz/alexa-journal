package journaldrive

import (
	"bytes"
	"context"
	"io/ioutil"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
)

type Journal struct {
	files *drive.FilesService
}

type JournalProvider struct {
}

func (jp *JournalProvider) Get(accessToken string) (*Journal, error) {
	return NewJournal(accessToken)
}

func NewJournal(accessToken string) (*Journal, error) {
	d, e := drive.New(
		oauth2.NewClient(
			context.TODO(),
			oauth2.StaticTokenSource(&oauth2.Token{
				AccessToken: accessToken,
			})))
	if e != nil {
		return nil, errors.Wrap(e, "Could not instantiate drive")
	}
	return &Journal{files: d.Files}, nil
}

func (j *Journal) AddEntry(date time.Time, text string) error {
	const filename = "my-journal.tsv"

	fileList, e := j.files.List().Q("name = '" + filename + "' and trashed = false").Do()
	if e != nil {
		return errors.Wrap(e, "Could not list files")
	}
	fileId := ""
	content := []byte("")
	switch len(fileList.Files) {
	case 0:
		file, e := j.files.Create(&drive.File{Name: filename}).Do()
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

		content, e = ioutil.ReadAll(download.Body)
		if e != nil {
			return errors.Wrap(e, "Could not read body")
		}
	default:
		return errors.New("Expected exactly 0 or 1 file in drive")
	}

	updatedContent := []byte(string(content) + date.String() + "\t" + text + "\n")

	_, e = j.files.Update(fileId, &drive.File{}).Media(bytes.NewReader(updatedContent)).Do()
	if e != nil {
		return errors.Wrap(e, "Could not upload updated content")
	}

	return nil
}
