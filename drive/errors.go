package drive

import (
	journalskill "github.com/petergtz/alexa-journal"
	"github.com/pkg/errors"
)

type DriveSheetErrorInterpreter struct {
	ErrorReporter journalskill.ErrorReporter
}

func (interpreter *DriveSheetErrorInterpreter) Interpret(e error) string {
	cause := errors.Cause(e)
	switch {
	case IsCannotCreateFileError(cause):
		return "Ich kann die Datei in Deinem Google Drive nicht anlegen. Bitte stelle sicher, dass Dein Google Drive mir erlaubt, darauf zuzugreifen."
	case IsMultipleFilesFoundError(cause):
		return "Ich habe in Deinem Google Drive mehr als eine Datei mit dem Namen Tagebuch gefunden. Bitte Stelle sicher, dass es nur eine Datei mit diesem Namen gibt."
	case IsSheetNotFoundError(cause):
		return "Ich habe in Deinem Spreadsheet kein Tabellenblatt mit dem Namen Tagebuch gefunden. Bitte stelle sicher, dass dies existiert."
	default:
		interpreter.ErrorReporter.ReportError(errors.Wrap(e, "Could not interpret this error."))
		return "Genauere Details kann ich aktuell leider nicht herausfinden. Ich habe den Entwickler bereits informiert, er wird sich um das Problem kümmern. Bitte versuche es später noch einmal."
	}
}

type CannotCreateFileError struct{ error }

func NewCannotCreateFileError(filename string, cause error) *CannotCreateFileError {
	return &CannotCreateFileError{errors.Errorf("CannotCreateFileError. filename: %v, cause: %v", filename, cause.Error())}
}
func IsCannotCreateFileError(e error) bool {
	_, is := e.(*CannotCreateFileError)
	return is
}

type MultipleFilesFoundError struct{ error }

func NewMultipleFilesFoundError(filename string) *MultipleFilesFoundError {
	return &MultipleFilesFoundError{errors.Errorf("MultipleFilesFoundError. filename: %v", filename)}
}
func IsMultipleFilesFoundError(e error) bool {
	_, is := e.(*MultipleFilesFoundError)
	return is
}

type SheetNotFoundError struct{ error }

func NewSheetNotFoundError(sheetsTitle string) *SheetNotFoundError {
	return &SheetNotFoundError{errors.Errorf("SheetNotFoundError. sheetTitle: %v", sheetsTitle)}
}
func IsSheetNotFoundError(e error) bool {
	_, is := e.(*SheetNotFoundError)
	return is
}
