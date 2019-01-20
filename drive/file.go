package drive

import (
	"context"
	"fmt"

	"github.com/petergtz/alexa-journal/util"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
)

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

func fileIDFrom(files *drive.FilesService, filename string, log *zap.SugaredLogger) (fileID string, err error) {
	fileList, e := files.List().Q("name = '" + filename + "' and trashed = false").Do()
	if e != nil {
		return "", errors.Wrap(e, "Could not list files")
	}
	switch len(fileList.Files) {
	case 0:
		return "", nil
	case 1:
		log.Infof("File %v already exists. Using it.", filename)
		return fileList.Files[0].Id, nil
	default:
		var filenames []string
		for _, file := range fileList.Files {
			filenames = append(filenames, fmt.Sprintf("%#v", file.Name))
		}
		return "", NewMultipleFilesFoundError(filename)
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
