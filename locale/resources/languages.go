package resources

import (
	"fmt"
	"strings"
	"time"
)

var DeDe = []byte(tomlStringFrom(map[StringID]string{
	YourJournalIsNowOpen: `Dein Tagebuch ist nun geöffnet. Was möchtest Du tun?`,
	NewEntryDraftExists:  `Fuer dieses Datum hast Du bereits einen Eintrag entworfen. Er lautet: {{.Draft}}. Moechtest Du mit diesem Eintrag weiter machen?`,

	// Not covered yet:
	YouCanNowCreateYourEntry: `Du kannst Deinen Eintrag {{.ForDate}} nun verfassen; ich werde jeden Teil kurz bestaetigen, sodass du die moeglichkeit hast ihn zu \"korrigieren\" oder \"anzuhoeren\". Sage \"fertig\", wenn Du fertig bist.`,

	YouCanNowCreateYourEntry_succinct: `Du kannst Deinen Eintrag {{.ForDate}} nun verfassen. Los geht's!`,
	ForDate:                           `für den {{.Date}}`,
	IRepeat:                           `Ich wiederhole: {{.Text}}.\n\nNaechster Teil bitte?`,

	// Not covered yet:
	NextPartPleaseReprompt: `Bitte verfasse den nächsten Teil Deines Eintrags.`,

	YourEntryIsEmptyNoRepeat:     `Dein Eintrag ist leer. Es gibt nichts zu wiederholen. Bitte verfasse zuerst den ersten Teil Deines Eintrags.`,
	YourEntryIsEmptyNoCorrect:    `Dein Eintrag ist leer. Es gibt nichts zu korrigieren. Bitte verfasse zuerst den ersten Teil Deines Eintrags.`,
	OkayCorrectPart:              `OK. Bitte verfasse den letzten Teil Deines Eintrags erneut.`,
	CorrectPartReprompt:          `Bitte verfasse den letzten Teil Deines Eintrags erneut.`,
	NewEntryAborted:              `Okay. Abgebrochen.`,
	YourEntryIsEmptyNoSave:       `Dein Eintrag ist leer. Es gibt nichts zu speichern.`,
	NewEntryConfirmation:         `Alles klar. Ich habe folgenden Eintrag für das Datum {{.Date}}: \"{{.Text}}\". Soll ich ihn so speichern?`,
	NewEntryConfirmationReprompt: `Soll ich Deinen Eintrag so speichern?`,
	OkaySaved:                    `Okay. Gespeichert.`,
	OkayNotSaved:                 `Okay. Nicht gespeichert.`,

	// Not covered yet:
	SuccinctModeExplanation: `Übrigens, falls Du keine langen Erklärungen haben möchtest, sage einfach \"Alexa, fasse Dich kurz\".`,

	WhatDoYouWantToDoNext: `Was möchtest Du als nächstes in Deinem Tagebuch machen?`,

	// Not covered yet:
	DidNotUnderstandTryAgain: `Ich habe Dich leider nicht richtig verstanden. Bitte versuche es noch einmal.`,
	ExampleRelativeDateQuery: `Sage z.B. was war heute vor einem Jahr?`,
	ExampleDateQuery:         `Sage z.B. \"was war im Juni 1997?\"`,
	CouldNotGetEntry:         `Oje. Beim Abrufen des Eintrags ist ein Fehler aufgetreten.`,
	CouldNotGetEntries:       `Oje. Beim Abrufen der Einträge ist ein Fehler aufgetreten.`,

	NoEntriesInTimeRangeFound: `Keine Einträge für den Zeitraum {{.TimeRange}} gefunden.`,

	// Not covered yet:
	EntriesInTimeRange: `Hier sind die Einträge für den Zeitraum {{.Date}}: {{.Entries}}`,
	ReadEntry:          `Hier ist der Eintrag vom {{.WeekDay}}, {{.Date}}: {{.Text}}.`,

	// Not covered yet:
	JournalIsEmpty:       `Dein Tagebuch ist noch leer.`,
	NewEntryExample:      `Sage z.B. neuen Eintrag erstellen.`,
	EntryForDateNotFound: `Ich habe fuer den {{.SearchDate}} keinen Eintrag gefunden. Der nächste Eintrag ist vom {{.WeekDay}}, {{.Date}}. Er lautet: {{.Text}}.`,
	SearchError:          `Oje. Beim Suchen nach Eintraegen ist ein Fehler aufgetreten.`,

	SearchNoResultsFound: `Keine Einträge für die Suche \"{{.Query}}\" gefunden.`,
	SearchResults:        `Hier sind die Ergebnisse für die Suche \"{{.Query}}\":`,
	DeleteEntryNotFound:  `Hm. Zu diesem Datum habe ich leider keinen Eintrag gefunden.`,

	// Not covered yet:
	DeleteEntryCouldNotGetEntry: `Oje. Beim Aufrufen des zu loeschenden Eintrags ist ein Fehler aufgetreten.`,

	DeleteEntryConfirmation: `Du moechtest den folgenden Eintrag loeschen: {{.Entry}}. Soll ich ihn wirklich loeschen?`,

	// Not covered yet:
	DeleteEntryError: `Oje. Beim Loeschen des Eintrags ist ein Fehler aufgetreten.`,

	OkayDeleted: `Okay. Geloescht.`,

	// Not covered yet:
	OkayNotDeleted: `Okay. Nicht geloescht.`,

	LinkWithGoogleAccount: `Bevor Du Dein Tagebuch öffnen kannst, verbinde bitte zuerst Alexa mit Deinem Google Account in der Alexa App.`,

	// Not covered yet:
	OkayWillBeSuccinct: `Okay. Ich werde mich kurzfassen. Falls ich wieder ausfuehrlicher sein soll, sage Alexa, sei ausfuehrlich.`,
	OkayWillBeVerbose:  `Okay. Ich werde ausführlich sein. Falls ich mich wieder kurzfassen soll, sage Alexa, fasse Dich kurz.`,
	InternalError:      `Es ist ein interner Fehler aufgetreten. Ich habe den Entwickler bereits informiert, er wird sich um das Problem kümmern. Bitte versuche es zu einem späteren Zeitpunkt noch einmal.`,

	Help: `Mit diesem Skill kannst Du Tagebucheintraege erstellen oder vorlesen lassen. Sage z.B. \"Neuen Eintrag erstellen\". Oder \"Lies mir den Eintrag von gestern vor\". Oder \"Was war heute vor 20 Jahren?\". Oder \"Was war im August 1994?\".`,
	// TODO: add to Help ` Oder \"Suche nach Geburtstag\". Wenn ich mich kurz fassen soll, sage \"Fasse Dich kurz\".`,
	Done:                         "fertig",
	Correct1:                     "korrigiere",
	Correct2:                     "korrigieren",
	Repeat1:                      "wiederhole",
	Repeat2:                      "wiederholen",
	Abort:                        "abbrechen",
	ShortPause:                   ` `,
	LongPause:                    `\n\n`,
	InvalidDate:                  `Das ist ein ungueltiges Datum. Bitte gib einen genauen Tag fuer das Datum an.`,
	DriveCannotCreateFileError:   "Ich kann die Datei in Deinem Google Drive nicht anlegen. Bitte stelle sicher, dass Dein Google Drive mir erlaubt, darauf zuzugreifen.",
	DriveMultipleFilesFoundError: `Ich habe in Deinem Google Drive mehr als eine Datei mit dem Namen Tagebuch gefunden. Bitte Stelle sicher, dass es nur eine Datei mit diesem Namen gibt.`,
	DriveSheetNotFoundError:      "Ich habe in Deinem Spreadsheet kein Tabellenblatt mit dem Namen Tagebuch gefunden. Bitte stelle sicher, dass dies existiert.",
	DriveUnknownError:            "Es gab einen Fehler. Genauere Details kann ich aktuell leider nicht herausfinden. Ich habe den Entwickler bereits informiert, er wird sich um das Problem kümmern. Bitte versuche es später noch einmal.",
	Journal:                      "Tagebuch",
}))

var weekdaysEn = map[time.Weekday]string{
	time.Monday:    "Monday",
	time.Tuesday:   "Tuesday",
	time.Wednesday: "Wednesday",
	time.Thursday:  "Thursday",
	time.Friday:    "Friday",
	time.Saturday:  "Saturday",
	time.Sunday:    "Sunday",
}

var Weekdays = map[string]map[time.Weekday]string{
	"de-DE": {
		time.Monday:    "Montag",
		time.Tuesday:   "Dienstag",
		time.Wednesday: "Mittwoch",
		time.Thursday:  "Donnerstag",
		time.Friday:    "Freitag",
		time.Saturday:  "Samstag",
		time.Sunday:    "Sonntag",
	},
	"en-US": weekdaysEn,
	"en-GB": weekdaysEn,
	"en-IN": weekdaysEn,
	"en-CA": weekdaysEn,
	"en-AU": weekdaysEn,
}

var EnUs = []byte(tomlStringFrom(map[StringID]string{
	YourJournalIsNowOpen: `Okay, your journal is open. What do you want to do next?`,
	NewEntryDraftExists:  `A draft for this date already exists. It is: {{.Draft}}. Do you want to continue with that?`,

	// Not covered yet:
	YouCanNowCreateYourEntry: `You can draft your entry {{.ForDate}} now; I'll quickly confirm every part, so you get the chance to correct it, if necessary. Say \"done\" when you're done, say \"abort\" when you'd like to abort.`,

	YouCanNowCreateYourEntry_succinct: `You can draft your entry {{.ForDate}} now; let's go!`,
	ForDate:                           `for {{.Date}}`,
	IRepeat:                           `I repeat: {{.Text}}.\n\nNext part please?`,

	// Not covered yet:
	NextPartPleaseReprompt: `Please draft the next part of your entry please.`,

	YourEntryIsEmptyNoRepeat:     `Your entry is empty. There's nothing to repeat. Please draft your first part of your entry.`,
	YourEntryIsEmptyNoCorrect:    `Your entry is empty. There's nothing to correct. Please draft your first part of your entry.`,
	OkayCorrectPart:              `OK. Please draft the last part of your entry again.`,
	CorrectPartReprompt:          `Please draft the last part of your entry again.`,
	NewEntryAborted:              `Okay. Aborted.`,
	YourEntryIsEmptyNoSave:       `Your entry is empty. There is nothing to save.`,
	NewEntryConfirmation:         `Alright. I have the following entry for {{.Date}}: \"{{.Text}}\".  Should I save it like this?`,
	NewEntryConfirmationReprompt: `Should I save your entry like this?`,
	OkaySaved:                    `Okay. Saved.`,
	OkayNotSaved:                 `Okay. Not saved.`,

	// Not covered yet:
	SuccinctModeExplanation: `By the way, if you don't want verbose explanations, just say \"Alexa, be brief\".`,

	WhatDoYouWantToDoNext: `What do you want to do next in your journal?`,

	// Not covered yet:
	DidNotUnderstandTryAgain: `Sorry, I didn't understand you correctly. Please try again.`,
	ExampleRelativeDateQuery: `Say e.g. what was one year ago?`,
	ExampleDateQuery:         `Say e.g. \"what was in June 1997\"?`,
	CouldNotGetEntry:         `Uh oh, there was an error when I tried to retrieve your entry.`,
	CouldNotGetEntries:       `Uh oh, there was an error when I tried to retrieve your entries.`,

	NoEntriesInTimeRangeFound: `No entries found for time range {{.TimeRange}}.`,

	// Not covered yet:
	EntriesInTimeRange: `Here are the entries for time range {{.Date}}: {{.Entries}}`,
	ReadEntry:          `Here's the entry from {{.WeekDay}}, {{.Date}}: {{.Text}}.`,

	// Not covered yet:
	JournalIsEmpty:       `Your journal is still empty.`,
	NewEntryExample:      `Say e.g. \"draft new entry\".`,
	EntryForDateNotFound: `I couldn't find an entry for {{.SearchDate}}. The next entry is from {{.WeekDay}}, {{.Date}}. It is: {{.Text}}.`,
	SearchError:          `Uh oh, there was an error when I tried to search for entries.`,

	SearchNoResultsFound: `I couldn't find any entries for the query \"{{.Query}}\".`,
	SearchResults:        `Here are the results for the query \"{{.Query}}\":`,
	DeleteEntryNotFound:  `Um. I couldn't find an entry for this date.`,

	// Not covered yet:
	DeleteEntryCouldNotGetEntry: `Uh oh, there was an error when I tried to access the entry you'd like to delete.`,

	DeleteEntryConfirmation: `You'd like to delete the following entry: {{.Entry}}. Should I really delete it?`,

	// Not covered yet:
	DeleteEntryError: `Uh oh, there was an error when I tried to delete the entry.`,

	OkayDeleted: `Okay. Deleted.`,

	// Not covered yet:
	OkayNotDeleted: `Okay. Not deleted.`,

	LinkWithGoogleAccount: `Before you can open your journal, please link Alexa with your Google account in your Alexa app.`,

	// Not covered yet:
	OkayWillBeSuccinct: `Okay. I'll be brief. In case I should be verbose again, say \"Alexa, be verbose\".`,
	OkayWillBeVerbose:  `Okay. I'll be verbose. In case I should be brief again, say \"Alexa, be brief\".`,
	InternalError:      `There was an internal error. I have already informed the engineer who will take care of the problem. Please try again later.`,

	Help: `With this skill, you create journal entries and have them read for you. Say e.g. \"new entry\" . Or \"Read the entry from yesterday\". Or \"What was 20 years ago today?\". Oder \"What was in August 1994?\".`,
	// TODO: add to Help ` Oder \"Suche nach Geburtstag\". Wenn ich mich kurz fassen soll, sage \"Fasse Dich kurz\".`,
	Done:                         "done",
	Correct1:                     "correct",
	Correct2:                     "revise",
	Repeat1:                      "repeat",
	Repeat2:                      "repeat",
	Abort:                        "abort",
	ShortPause:                   ` `,
	LongPause:                    `\n\n`,
	InvalidDate:                  `That's an invalid date. Please provide a specific day for this date.`,
	DriveCannotCreateFileError:   `I cannot create the file in your Google Drive. Please make sure that your Google Drive allows me to access it.`,
	DriveMultipleFilesFoundError: `I found more than one file with the name Journal in your Google Drive. Please make sure that there is only one file with this name.`,
	DriveSheetNotFoundError:      `I couldn't find a sheet with the name Journal in your spreadsheet. Please make sure this sheet exists.`,
	DriveUnknownError:            `There was an error. Unfortunately, I can't find out more details at the moment. I have already informed the engineer who will take care of the problem. Please try again later.`,
	Journal:                      `Journal`,
}))

func tomlStringFrom(stringMap map[StringID]string) string {
	var lines []string
	for i := 0; i < int(EndMarker); i++ {
		lines = append(lines, fmt.Sprintf("%v = \"%v\"", StringID(i).String(), stringMap[StringID(i)]))
	}
	return strings.Join(lines, "\n")
}
