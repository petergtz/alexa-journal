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
	files   *drive.FilesService
	fileID  string
	content string
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
	fileID, e := getOrCreateJournalFile(d.Files, filename)
	if e != nil {
		panic(errors.Wrap(e, "Could not get or create journal file in Google drive."))
	}
	content, e := contentOfFile(d.Files, fileID)
	if e != nil {
		panic(errors.Wrap(e, "Could not get file contents"))
	}

	return &Journal{files: d.Files, fileID: fileID, content: content}
}

func (j *Journal) AddEntry(entryDate date.Date, text string) error {
	if j.content == "" {
		j.content = "timestamp\tdate\ttext\n"
	}
	j.content += time.Now().Format("2006-01-02 15:04:05 -0700 MST") + "\t" + entryDate.String() + "\t" + text + "\n"

	_, e := j.files.Update(j.fileID, &drive.File{}).Media(strings.NewReader(j.content)).Do()
	if e != nil {
		return errors.Wrap(e, "Could not upload updated content")
	}

	return nil
}

func (j *Journal) GetEntry(entryDate date.Date) (string, error) {
	for _, line := range strings.Split((j.content), "\n") {
		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			continue
		}
		if parts[1] == entryDate.String() {
			return parts[2], nil
		}
	}
	return "Keinen Eintrag gefunden", nil
}

type Entry struct {
	Timestamp string
	EntryDate string
	EntryText string
}

func (j *Journal) GetEntries(timeRange string) ([]Entry, error) {
	var result []Entry
	for _, line := range strings.Split((j.content), "\n") {
		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			continue
		}
		if strings.HasPrefix(parts[1], timeRange) {
			result = append(result, Entry{parts[0], parts[1], parts[2]})
		}
	}
	return result, nil
}

func getOrCreateJournalFile(files *drive.FilesService, filename string) (fileID string, err error) {
	fileList, e := files.List().Q("name = '" + filename + "' and trashed = false").Do()
	if e != nil {
		return "", errors.Wrap(e, "Could not list files")
	}
	switch len(fileList.Files) {
	case 0:
		file, e := files.Create(&drive.File{Name: filename}).Media(strings.NewReader("timestamp\tdate\ttext\n")).Do()
		if e != nil {
			return "", errors.Wrap(e, "Could not create file to store journal")
		}
		return file.Id, nil
	case 1:
		return fileList.Files[0].Id, nil
	default:
		return "", errors.New("Expected exactly 0 or 1 file in drive")
	}
}

func contentOfFile(files *drive.FilesService, fileID string) (string, error) {
	download, e := files.Get(fileID).Download()
	if e != nil {
		return "", errors.Wrap(e, "Could not download file")
	}
	defer download.Body.Close()

	byteContent, e := ioutil.ReadAll(download.Body)
	if e != nil {
		return "", errors.Wrap(e, "Could not read body")
	}
	return string(byteContent), nil
}
