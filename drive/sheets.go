package drive

import (
	"context"
	"fmt"
	"strings"

	"github.com/petergtz/alexa-journal/util"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/sheets/v4"
)

type SheetBasedTabularData struct {
	Service       *sheets.Service
	Log           *zap.SugaredLogger
	SpreadsheetID string
	sheetTitle    string
}

func NewSheetBasedTabularData(accessToken string, filename string, sheetTitle string, log *zap.SugaredLogger) *SheetBasedTabularData {
	sheetsService := newSheetsService(accessToken)
	return &SheetBasedTabularData{
		Service:       sheetsService,
		Log:           log,
		SpreadsheetID: getOrCreateSpreadsheet(newDriveService(accessToken).Files, sheetsService, filename, sheetTitle, log),
		sheetTitle:    sheetTitle,
	}
}

func newSheetsService(accessToken string) *sheets.Service {
	sh, e := sheets.New(oauth2.NewClient(
		context.TODO(),
		oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: accessToken,
		})))
	if e != nil {
		panic(errors.Wrap(e, "Could not instantiate sheets client"))
	}
	return sh
}

func getOrCreateSpreadsheet(files *drive.FilesService, sh *sheets.Service, filename string, sheetTitle string, log *zap.SugaredLogger) string {
	fileList, e := files.List().Q("name = '" + filename + "' and trashed = false").Do()
	util.PanicOnError(errors.Wrap(e, "Could not list files"))
	switch len(fileList.Files) {
	case 0:
		log.Infof("Spreadsheet %v does not exist. Creating it.", filename)
		ss, e := sh.Spreadsheets.Create(&sheets.Spreadsheet{
			Properties: &sheets.SpreadsheetProperties{Title: filename},
			Sheets: []*sheets.Sheet{&sheets.Sheet{
				Properties: &sheets.SheetProperties{Title: sheetTitle},
			}},
		}).Do()
		util.PanicOnError(errors.Wrap(e, "Could not create spreadsheet"))
		return ss.SpreadsheetId
	case 1:
		log.Infof("Spreadsheet %v already exists. Using it.", filename)
		return fileList.Files[0].Id
	default:
		var filenames []string
		for _, file := range fileList.Files {
			filenames = append(filenames, fmt.Sprintf("%#v", file.Name))
		}
		panic(errors.Errorf("Expected exactly 0 or 1 Spreadsheet in drive. Found: %v", strings.Join(filenames, ", ")))
	}
}

func (td *SheetBasedTabularData) AppendRow(row []string) {
	if len(row) != 3 {
		panic(errors.Errorf("Currently only rows with 3 cells are supported. Given: %v", row))
	}
	interfaceRow := make([]interface{}, len(row))
	for i, cell := range row {
		interfaceRow[i] = cell
	}
	_, e := td.Service.Spreadsheets.Values.Append(td.SpreadsheetID, td.sheetTitle+"!A1:C1", &sheets.ValueRange{
		Values: [][]interface{}{interfaceRow},
	}).ValueInputOption("USER_ENTERED").Do()
	util.PanicOnError(errors.Wrapf(e, "Could not append values %v to spreadhseet", row))
}

func (td *SheetBasedTabularData) Rows() [][]string {
	resp, e := td.Service.Spreadsheets.Values.Get(td.SpreadsheetID, td.sheetTitle).Do()
	util.PanicOnError(errors.Wrapf(e, "Could not get values"))
	if len(resp.Values) == 0 {
		return nil
	}
	result := make([][]string, len(resp.Values))
	for rowNum, row := range resp.Values {
		result[rowNum] = make([]string, len(row))
		for colNum, cell := range row {
			result[rowNum][colNum] = cell.(string)
		}
	}
	return result
}

func (td *SheetBasedTabularData) Empty() bool {
	resp, e := td.Service.Spreadsheets.Values.Get(td.SpreadsheetID, td.sheetTitle).Do()
	util.PanicOnError(errors.Wrapf(e, "Could not get values"))
	return len(resp.Values) == 0
}
