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
	fileID string
	log    *zap.SugaredLogger
}

func NewFileService(accessToken string, filename string, log *zap.SugaredLogger) *FileService {
	startTime := time.Now()
	d, e := drive.New(
		oauth2.NewClient(
			context.TODO(),
			oauth2.StaticTokenSource(&oauth2.Token{
				AccessToken: accessToken,
			})))
	util.PanicOnError(errors.Wrap(e, "Could not instantiate drive"))

	log.Debugw("Time taken to create drive client", "time", time.Since(startTime))
	startTime = time.Now()
	fileID, e := getOrCreateFile(d.Files, filename, log)
	util.PanicOnError(e)

	log.Debugw("Time taken to get or create file in drive", "time", time.Since(startTime))
	return &FileService{fileID: fileID, files: d.Files}
}

func (dfs *FileService) Upload(content string) {
	_, e := dfs.files.Update(dfs.fileID, &drive.File{}).Media(strings.NewReader(content)).Do()
	util.PanicOnError(errors.Wrap(e, "Could not upload file contents"))
}

func (dfs *FileService) Download() string {
	download, e := dfs.files.Get(dfs.fileID).Download()
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
