package drive

import (
	journalskill "github.com/petergtz/alexa-journal"
	r "github.com/petergtz/alexa-journal/locale/resources"
	"github.com/pkg/errors"
)

type DriveSheetErrorInterpreter struct {
	ErrorReporter journalskill.ErrorReporter
}

func (interpreter *DriveSheetErrorInterpreter) Interpret(e error, l journalskill.Localizer) string {
	cause := errors.Cause(e)
	switch {
	case IsCannotCreateFileError(cause):
		return l.Get(r.DriveCannotCreateFileError)
	case IsMultipleFilesFoundError(cause):
		return l.Get(r.DriveMultipleFilesFoundError)
	case IsSheetNotFoundError(cause):
		return l.Get(r.DriveSheetNotFoundError)
	default:
		interpreter.ErrorReporter.ReportError(errors.Wrap(e, "Could not interpret this error."))
		return l.Get(r.DriveUnknownError)
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
