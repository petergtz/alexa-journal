package resources

import (
	"fmt"
	"strings"
	"time"
)

var DeDe = []byte(tomlStringFrom(map[StringID]string{
	YourJournalIsNowOpen:              `Dein Tagebuch ist nun geöffnet. Was möchtest Du tun?`,
	NewEntryDraftExists:               `Fuer dieses Datum hast Du bereits einen Eintrag entworfen. Er lautet: {{.Draft}}. Moechtest Du mit diesem Eintrag weiter machen?`,
	YouCanNowCreateYourEntry:          `Du kannst Deinen Eintrag {{.ForDate}} nun verfassen; ich werde jeden Teil kurz bestaetigen, sodass du die moeglichkeit hast ihn zu \"korrigieren\" oder \"anzuhoeren\". Sage \"fertig\", wenn Du fertig bist.`,
	YouCanNowCreateYourEntry_succinct: `Du kannst Deinen Eintrag {{.ForDate}} nun verfassen. Los geht's!`,
	ForDate:                           `für den {{.Date}}`,
	IRepeat:                           `Ich wiederhole: {{.Text}}.\n\nNaechster Teil bitte?`,
	NextPartPlease:                    `Bitte verfasse den nächsten Teil Deines Eintrags.`,
	YourEntryIsEmptyNoRepeat:          `Dein Eintrag ist leer. Es gibt nichts zu wiederholen. Bitte verfasse zuerst den ersten Teil Deines Eintrags.`,
	YourEntryIsEmptyNoCorrect:         `Dein Eintrag ist leer. Es gibt nichts zu korrigieren. Bitte verfasse zuerst den ersten Teil Deines Eintrags.`,
	OkayCorrectPart:                   `OK. Bitte verfasse den letzten Teil Deines Eintrags erneut.`,
	CorrectPart:                       `Bitte verfasse den letzten Teil Deines Eintrags erneut.`,
	NewEntryAborted:                   `Okay. Abgebrochen.`,
	YourEntryIsEmptyNoSave:            `Dein Eintrag ist leer. Es gibt nichts zu speichern.`,
	NewEntryConfirmation:              `Alles klar. Ich habe folgenden Eintrag für das Datum {{.Date}}: \"{{.Text}}\". Soll ich ihn so speichern?`,
	NewEntryConfirmationReprompt:      `Soll ich Deinen Eintrag so speichern?`,
	OkaySaved:                         `Okay. Gespeichert.`,
	OkayNotSaved:                      `Okay. Nicht gespeichert.`,
	SuccinctModeExplanation:           `Übrigens, falls Du keine langen Erklärungen haben möchtest, sage einfach \"Alexa, fasse Dich kurz\".`,
	WhatDoYouWantToDoNext:             `Was möchtest Du als nächstes in Deinem Tagebuch machen?`,
	DidNotUnderstandTryAgain:          `Ich habe Dich leider nicht richtig verstanden. Bitte versuche es noch einmal.`,
	ExampleRelativeDateQuery:          `Sage z.B. was war heute vor einem Jahr?`,
	ExampleDateQuery:                  `Sage z.B. \"was war im Juni 1997?\"`,
	CouldNotGetEntry:                  `Oje. Beim Abrufen des Eintrags ist ein Fehler aufgetreten.`,
	CouldNotGetEntries:                `Oje. Beim Abrufen der Einträge ist ein Fehler aufgetreten.`,
	NoEntriesInTimeRangeFound:         `Keine Einträge für den Zeitraum {{.TimeRange}} gefunden.`,
	EntriesInTimeRange:                `Hier sind die Einträge für den Zeitraum {{.Date}}: {{.Entries}}`,
	ReadEntry:                         `Hier ist der Eintrag vom {{.WeekDay}}, {{.Date}}: {{.Text}}.`,
	JournalIsEmpty:                    `Dein Tagebuch ist noch leer.`,
	NewEntryExample:                   `Sage z.B. neuen Eintrag erstellen.`,
	EntryForDateNotFound:              `Ich habe fuer den {{.SearchDate}} keinen Eintrag gefunden. Der nächste Eintrag ist vom {{.WeekDay}}, {{.Date}}. Er lautet: {{.Text}}.`,
	SearchError:                       `Oje. Beim Suchen nach Eintraegen ist ein Fehler aufgetreten.`,
	SearchNoResultsFound:              `Keine Einträge für die Suche \"{{.Query}}\" gefunden.`,
	SearchResults:                     `Hier sind die Ergebnisse für die Suche \"{{.Query}}\":`,
	DeleteEntryNotFound:               `Hm. Zu diesem Datum habe ich leider keinen Eintrag gefunden.`,
	DeleteEntryCouldNotGetEntry:       `Oje. Beim Aufrufen des zu loeschenden Eintrags ist ein Fehler aufgetreten.`,
	DeleteEntryConfirmation:           `Du moechtest den folgenden Eintrag loeschen: {{.Entry}}. Soll ich ihn wirklich loeschen?`,
	DeleteEntryError:                  `Oje. Beim Loeschen des Eintrags ist ein Fehler aufgetreten.`,
	OkayDeleted:                       `Okay. Geloescht.`,
	OkayNotDeleted:                    `Okay. Nicht geloescht.`,
	LinkWithGoogleAccount:             `Bevor Du Dein Tagebuch öffnen kannst, verbinde bitte zuerst Alexa mit Deinem Google Account in der Alexa App.`,
	OkayWillBeSuccinct:                `Okay. Ich werde mich kurzfassen. Falls ich wieder ausfuehrlicher sein soll, sage Alexa, sei ausfuehrlich.`,
	OkayWillBeVerbose:                 `Okay. Ich werde ausführlich sein. Falls ich mich wieder kurzfassen soll, sage Alexa, fasse Dich kurz.`,
	InternalError:                     `Es ist ein interner Fehler aufgetreten. Ich habe den Entwickler bereits informiert, er wird sich um das Problem kümmern. Bitte versuche es zu einem späteren Zeitpunkt noch einmal.`,
	Help:                              `Mit diesem Skill kannst Du Tagebucheintraege erstellen oder vorlesen lassen. Sage z.B. \"Neuen Eintrag erstellen\". Oder \"Lies mir den Eintrag von gestern vor\". Oder \"Was war heute vor 20 Jahren?\". Oder \"Was war im August 1994?\". Oder \"Suche nach Geburtstag\". Wenn ich mich kurz fassen soll, sage \"Fasse Dich kurz\".`,
	ShortPause:                        ` `,
	LongPause:                         `\n\n`,
	InvalidDate:                       `Das ist ein ungueltiges Datum. Bitte gib einen genauen Tag fuer das Datum an.`,
}))

var Weekdays = map[string]map[time.Weekday]string{
	"DeDe": {
		time.Monday:    "Montag",
		time.Tuesday:   "Dienstag",
		time.Wednesday: "Mittwoch",
		time.Thursday:  "Donnerstag",
		time.Friday:    "Freitag",
		time.Saturday:  "Samstag",
		time.Sunday:    "Sonntag",
	},
}

func tomlStringFrom(stringMap map[StringID]string) string {
	var lines []string
	for i := 0; i < int(EndMarker); i++ {
		lines = append(lines, fmt.Sprintf("%v = \"%v\"", StringID(i).String(), stringMap[StringID(i)]))
	}
	return strings.Join(lines, "\n")
}
