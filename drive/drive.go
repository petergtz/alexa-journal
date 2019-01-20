package drive

import (
	"io/ioutil"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"google.golang.org/api/drive/v3"
)

type FileService struct {
	files  *drive.FilesService
	FileID string
	log    *zap.SugaredLogger
}

func NewFileService(accessToken string, filename string, log *zap.SugaredLogger) (*FileService, error) {
	startTime := time.Now()
	driveService := newDriveService(accessToken)
	log.Debugw("Time taken to create drive client", "time", time.Since(startTime))

	startTime = time.Now()
	fileID, e := fileIDFrom(driveService.Files, filename, log)
	if e != nil {
		return nil, e
	}
	if fileID == "" {
		log.Infof("File %v does not exist. Creating it.", filename)
		file, e := driveService.Files.Create(&drive.File{Name: filename}).Media(strings.NewReader("timestamp\tdate\ttext\n")).Do()
		if e != nil {
			return nil, NewCannotCreateFileError(filename, e)
		}
		fileID = file.Id
	}

	log.Debugw("Time taken to get or create file in drive", "time", time.Since(startTime))

	return &FileService{FileID: fileID, files: driveService.Files}, nil
}

func (dfs *FileService) Upload(content string) error {
	_, e := dfs.files.Update(dfs.FileID, &drive.File{}).Media(strings.NewReader(content)).Do()
	if e != nil {
		return errors.Wrap(e, "Could not upload file contents")
	}
	return nil
}

func (dfs *FileService) Download() (string, error) {
	download, e := dfs.files.Get(dfs.FileID).Download()
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
