package resources

//go:generate stringer -type=StringID
type StringID int

const (
	YourJournalIsNowOpen StringID = iota
	NewEntryDraftExists
	YouCanNowCreateYourEntry
	YouCanNowCreateYourEntry_succinct
	ForDate
	IRepeat
	NextPartPleaseReprompt
	YourEntryIsEmptyNoRepeat
	YourEntryIsEmptyNoCorrect
	OkayCorrectPart
	CorrectPartReprompt
	NewEntryAborted
	YourEntryIsEmptyNoSave
	NewEntryConfirmation
	NewEntryConfirmationReprompt
	OkaySaved
	OkayNotSaved
	SuccinctModeExplanation
	WhatDoYouWantToDoNext
	DidNotUnderstandTryAgain
	ExampleRelativeDateQuery
	ExampleDateQuery
	CouldNotGetEntry
	CouldNotGetEntries
	NoEntriesInTimeRangeFound
	EntriesInTimeRange
	ReadEntry
	JournalIsEmpty
	NewEntryExample
	EntryForDateNotFound
	SearchError
	SearchNoResultsFound
	SearchResults
	DeleteEntryNotFound
	DeleteEntryCouldNotGetEntry
	DeleteEntryConfirmation
	DeleteEntryError
	OkayDeleted
	OkayNotDeleted
	LinkWithGoogleAccount
	OkayWillBeSuccinct
	OkayWillBeVerbose
	InvalidDate
	InternalError
	Help
	Done
	Correct1
	Correct2
	Repeat1
	Repeat2
	Abort
	ShortPause
	LongPause
	EndMarker
	DriveCannotCreateFileError
	DriveMultipleFilesFoundError
	DriveSheetNotFoundError
	DriveUnknownError
)
