package drive

import "github.com/pkg/errors"

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
