package drive

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"google.golang.org/api/sheets/v4"
)

type SheetBasedTabularData struct {
	Service       *sheets.Service
	Log           *zap.SugaredLogger
	SpreadsheetID string
	sheetTitle    string
}

func NewSheetBasedTabularData(accessToken string, filename string, sheetTitle string, log *zap.SugaredLogger) (*SheetBasedTabularData, error) {
	sheetsService := newSheetsService(accessToken)
	spreadsheetID, e := fileIDFrom(newDriveService(accessToken).Files, filename, log)
	if e != nil {
		return nil, e
	}
	if spreadsheetID == "" {
		log.Infof("Spreadsheet %v does not exist. Creating it.", filename)
		ss, e := sheetsService.Spreadsheets.Create(&sheets.Spreadsheet{
			Properties: &sheets.SpreadsheetProperties{Title: filename},
			Sheets: []*sheets.Sheet{&sheets.Sheet{
				Properties: &sheets.SheetProperties{Title: sheetTitle},
			}},
		}).Do()
		if e != nil {
			return nil, NewCannotCreateFileError(filename, e)
		}
		spreadsheetID = ss.SpreadsheetId
	}
	return &SheetBasedTabularData{
		Service:       sheetsService,
		Log:           log,
		SpreadsheetID: spreadsheetID,
		sheetTitle:    sheetTitle,
	}, nil
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

func (td *SheetBasedTabularData) AppendRow(row []string) error {
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
	if e != nil {
		return errors.Wrapf(e, "Could not append values %v to spreadhseet", row)
	}
	return nil
}

func (td *SheetBasedTabularData) Rows() ([][]string, error) {
	resp, e := td.Service.Spreadsheets.Values.Get(td.SpreadsheetID, td.sheetTitle).Do()
	if e != nil {
		return nil, errors.Wrapf(e, "Could not get values")
	}
	if len(resp.Values) == 0 {
		return nil, nil
	}
	result := make([][]string, len(resp.Values))
	for rowNum, row := range resp.Values {
		result[rowNum] = make([]string, len(row))
		for colNum, cell := range row {
			result[rowNum][colNum] = cell.(string)
		}
	}
	return result, nil
}

func (td *SheetBasedTabularData) Empty() (bool, error) {
	resp, e := td.Service.Spreadsheets.Values.Get(td.SpreadsheetID, td.sheetTitle).Do()
	if e != nil {
		return false, errors.Wrapf(e, "Could not get values")
	}
	return len(resp.Values) == 0, nil
}
