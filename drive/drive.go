package drive

import (
	"context"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/petergtz/alexa-journal/util"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
)

type FileService struct {
	files  *drive.FilesService
	FileID string
	log    *zap.SugaredLogger
}

func newDriveService(accessToken string) *drive.Service {
	driveService, e := drive.New(
		oauth2.NewClient(
			context.TODO(),
			oauth2.StaticTokenSource(&oauth2.Token{
				AccessToken: accessToken,
			})))
	util.PanicOnError(errors.Wrap(e, "Could not instantiate drive"))
	return driveService
}

func NewFileService(accessToken string, filename string, log *zap.SugaredLogger) *FileService {
	startTime := time.Now()
	driveService := newDriveService(accessToken)

	log.Debugw("Time taken to create drive client", "time", time.Since(startTime))
	startTime = time.Now()
	fileID, e := getOrCreateFile(driveService.Files, filename, log)
	util.PanicOnError(e)

	log.Debugw("Time taken to get or create file in drive", "time", time.Since(startTime))
	return &FileService{FileID: fileID, files: driveService.Files}
}

func (dfs *FileService) Upload(content string) {
	_, e := dfs.files.Update(dfs.FileID, &drive.File{}).Media(strings.NewReader(content)).Do()
	util.PanicOnError(errors.Wrap(e, "Could not upload file contents"))
}

func (dfs *FileService) Download() string {
	download, e := dfs.files.Get(dfs.FileID).Download()
	util.PanicOnError(errors.Wrap(e, "Could not download file"))
	defer download.Body.Close()

	byteContent, e := ioutil.ReadAll(download.Body)
	util.PanicOnError(errors.Wrap(e, "Could not read body"))

	return string(byteContent)
}

func getOrCreateFile(files *drive.FilesService, filename string, log *zap.SugaredLogger) (fileID string, err error) {
	fileList, e := files.List().Q("name = '" + filename + "' and trashed = false").Do()
	if e != nil {
		return "", errors.Wrap(e, "Could not list files")
	}
	switch len(fileList.Files) {
	case 0:
		log.Infof("File %v does not exist. Creating it.", filename)
		file, e := files.Create(&drive.File{Name: filename}).Media(strings.NewReader("timestamp\tdate\ttext\n")).Do()
		if e != nil {
			return "", errors.Wrapf(e, "Could not create file %v", filename)
		}
		return file.Id, nil
	case 1:
		log.Infof("File %v already exists. Using it.", filename)
		return fileList.Files[0].Id, nil
	default:
		var filenames []string
		for _, file := range fileList.Files {
			filenames = append(filenames, fmt.Sprintf("%#v", file.Name))
		}
		return "", errors.Errorf("Expected exactly 0 or 1 file in drive. Found: %v", strings.Join(filenames, ", "))
	}
}

func DeleteFile(accessToken string, fileID string) {
	e := newDriveService(accessToken).Files.Delete(fileID).Do()
	util.PanicOnError(errors.Wrap(e, "Could not delete file"))
}

func MoveToTrash(accessToken string, fileID string) {
	_, e := newDriveService(accessToken).Files.Update(fileID, &drive.File{Trashed: true}).Do()
	util.PanicOnError(errors.Wrap(e, "Could not move file to trash"))
}
