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
	ShortPause
	LongPause
	EndMarker
)
