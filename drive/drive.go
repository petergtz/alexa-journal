package journaldrive

import (
	"context"
	"io/ioutil"
	"strings"

	"github.com/petergtz/alexa-journal/journal"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
)

type DriveJournalProvider struct{}

func (jp *DriveJournalProvider) Get(accessToken string) journal.JournalFileService {
	return NewDriveJournalFileService(accessToken, "my-journal.tsv")
}

type DriveJournalFileService struct {
	files  *drive.FilesService
	fileID string
}

func NewDriveJournalFileService(accessToken string, filename string) *DriveJournalFileService {
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
	return &DriveJournalFileService{fileID: fileID, files: d.Files}
}

func (dfs *DriveJournalFileService) Update(content string) error {
	_, e := dfs.files.Update(dfs.fileID, &drive.File{}).Media(strings.NewReader(content)).Do()
	return e
}

func (dfs *DriveJournalFileService) Content() string {
	content, e := contentOfFile(dfs.files, dfs.fileID)
	if e != nil {
		panic(errors.Wrap(e, "Could not get file contents"))
	}
	return content
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
