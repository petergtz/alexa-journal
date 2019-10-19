package journalskill

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/petergtz/alexa-journal/github"
	"github.com/petergtz/alexa-journal/util"
	"github.com/rickb777/date"

	j "github.com/petergtz/alexa-journal/journal"

	"github.com/pkg/errors"

	"strings"

	alexa "github.com/petergtz/go-alexa"
	"go.uber.org/zap"
)

const responseTextLimit = 8000

type JournalProvider interface {
	Get(accessToken string) (j.Journal, error)
}

type ErrorInterpreter interface {
	Interpret(error) string
}

type JournalSkill struct {
	journalProvider  JournalProvider
	errorInterpreter ErrorInterpreter
	log              *zap.SugaredLogger
	errorReporter    *github.GithubErrorReporter
}

func NewJournalSkill(journalProvider JournalProvider,
	errorInterpreter ErrorInterpreter,
	log *zap.SugaredLogger,
	errorReporter *github.GithubErrorReporter,
) *JournalSkill {
	return &JournalSkill{
		journalProvider:  journalProvider,
		errorInterpreter: errorInterpreter,
		log:              log,
		errorReporter:    errorReporter,
	}
}

const helpText = "Mit diesem Skill kannst Du Tagebucheintraege erstellen oder vorlesen lassen. Sage z.B. \"Neuen Eintrag erstellen\". Oder \"Lies mir den Eintrag von gestern vor\". Oder \"Was war heute vor 20 Jahren?\". Oder \"Was war im August 1994?\"."

var months = map[string]int{
	"januar":    1,
	"februar":   2,
	"maerz":     3,
	"april":     4,
	"mai":       5,
	"juni":      6,
	"juli":      7,
	"august":    8,
	"september": 9,
	"oktober":   10,
	"november":  11,
	"dezember":  12,
}

var monthsReverse = map[int]string{
	1:  "januar",
	2:  "februar",
	3:  "maerz",
	4:  "april",
	5:  "mai",
	6:  "juni",
	7:  "juli",
	8:  "august",
	9:  "september",
	10: "oktober",
	11: "november",
	12: "dezember",
}

var weekdays = map[string]string{
	"Monday":    "Montag",
	"Tuesday":   "Dienstag",
	"Wednesday": "Mittwoch",
	"Thursday":  "Donnerstag",
	"Friday":    "Freitag",
	"Saturday":  "Samstag",
	"Sunday":    "Sonntag",
}

type SessionAttributes struct {
	Drafts   map[string][]string `json:"drafts"`
	Drafting bool                `json:"drafting"`
}

func (h *JournalSkill) ProcessRequest(requestEnv *alexa.RequestEnvelope) (responseEnv *alexa.ResponseEnvelope) {
	defer func() {
		if e := recover(); e != nil {
			h.errorReporter.ReportPanic(e)
			responseEnv = internalError()
		}
	}()

	log := h.log.With("request", requestEnv.Request, "session", requestEnv.Session)
	log.Infow("Request started")
	defer log.Infow("Request completed")

	if requestEnv.Session.User.AccessToken == "" {
		return &alexa.ResponseEnvelope{Version: "1.0",
			Response: &alexa.Response{
				OutputSpeech:     plainText("Bevor Du Dein Tagebuch öffnen kannst, verbinde bitte zuerst Alexa mit Deinem Google Account in der Alexa App."),
				Card:             &alexa.Card{Type: "LinkAccount"},
				ShouldSessionEnd: true,
			},
			SessionAttributes: requestEnv.Session.Attributes,
		}
	}

	switch requestEnv.Request.Type {

	case "LaunchRequest":
		// cache warming:
		go h.journalProvider.Get(requestEnv.Session.User.AccessToken)

		return &alexa.ResponseEnvelope{Version: "1.0",
			Response:          &alexa.Response{OutputSpeech: plainText("Dein Tagebuch ist nun geöffnet. Was möchtest Du tun?")},
			SessionAttributes: requestEnv.Session.Attributes,
		}

	case "IntentRequest":
		journal, e := h.journalProvider.Get(requestEnv.Session.User.AccessToken)
		if e != nil {
			log.Errorw("Error while getting journal via journalProvider", "error", e)
			return plainTextRespEnv(h.errorInterpreter.Interpret(e), requestEnv.Session.Attributes)
		}
		log.Debugw("Journal downloaded")

		var sessionAttributes SessionAttributes
		sessionAttributes.Drafts = make(map[string][]string)
		e = mapstructure.Decode(requestEnv.Session.Attributes, &sessionAttributes)
		util.PanicOnError(errors.Wrap(e, "Could not parse sessionAttributes"))

		intent := requestEnv.Request.Intent
		switch intent.Name {
		case "NewEntryIntent":
			switch requestEnv.Request.DialogState {
			case "STARTED":
				return pureDelegate(&intent, requestEnv.Session.Attributes)
			case "IN_PROGRESS":
				switch intent.ConfirmationStatus {
				case "NONE":
					if intent.Slots["date"].Value == "" {
						return pureDelegate(&intent, requestEnv.Session.Attributes)
					}
					// TODO: could we use intent.Slots["text"].Value == "" instead of !sessionAttributes.Drafting?
					if _, exists := sessionAttributes.Drafts[intent.Slots["date"].Value]; exists && !sessionAttributes.Drafting {
						switch intent.Slots["text"].ConfirmationStatus {
						case "NONE":
							return &alexa.ResponseEnvelope{Version: "1.0",
								Response: &alexa.Response{
									OutputSpeech: plainText("Fuer dieses Datum hast Du bereits einen Eintrag entworfen. " +
										"Er lautet: " + strings.Join(sessionAttributes.Drafts[intent.Slots["date"].Value], ". ") + "." +
										" Moechtest Du mit diesem Eintrag weiter machen?"),
									Directives: []interface{}{alexa.DialogDirective{Type: "Dialog.ConfirmSlot", SlotToConfirm: "text"}},
								},
								SessionAttributes: mapStringInterfaceFrom(sessionAttributes),
							}
						case "CONFIRMED":
							break
						case "DENIED":
							delete(sessionAttributes.Drafts, intent.Slots["date"].Value)
						}
					}
					switch intent.Slots["text"].Value {
					case "":
						sessionAttributes.Drafting = true
						dateString := ""
						if intent.Slots["date"].Value != "" {
							dateString = "für den " + intent.Slots["date"].Value
						}
						return &alexa.ResponseEnvelope{Version: "1.0",
							Response: &alexa.Response{
								OutputSpeech: plainText("Du kannst Deinen eintrag " + dateString + " nun verfassen; ich werde jeden Teil kurz bestaetigen, sodass du die moeglichkeit hast ihn zu \"korrigieren\" oder \"anzuhoeren\". Sage \"fertig\", wenn Du fertig bist."),
								Directives:   []interface{}{alexa.DialogDirective{Type: "Dialog.ElicitSlot", SlotToElicit: "text"}},
							},
							SessionAttributes: mapStringInterfaceFrom(sessionAttributes),
						}
					default:
						sessionAttributes.Drafts[intent.Slots["date"].Value] = append(sessionAttributes.Drafts[intent.Slots["date"].Value], intent.Slots["text"].Value)
						return &alexa.ResponseEnvelope{Version: "1.0",
							Response: &alexa.Response{
								OutputSpeech: plainText("OK, weiter?"),
								Directives:   []interface{}{alexa.DialogDirective{Type: "Dialog.ElicitSlot", SlotToElicit: "text"}},
							},
							SessionAttributes: mapStringInterfaceFrom(sessionAttributes),
						}
					case "wiederhole", "wiederholen":
						if len(sessionAttributes.Drafts[intent.Slots["date"].Value]) == 0 {
							return &alexa.ResponseEnvelope{Version: "1.0",
								Response: &alexa.Response{
									OutputSpeech: plainText("Dein Eintrag ist leer. Es gibt nichts zu wiederholen. Bitte verfasse zuerst den ersten Teil Deines Eintrags."),
									Directives:   []interface{}{alexa.DialogDirective{Type: "Dialog.ElicitSlot", SlotToElicit: "text"}},
								},
								SessionAttributes: requestEnv.Session.Attributes,
							}
						}
						return &alexa.ResponseEnvelope{Version: "1.0",
							Response: &alexa.Response{
								OutputSpeech: plainText("Hier ist der letzte Teil Deines Eintrags: " +
									sessionAttributes.Drafts[intent.Slots["date"].Value][len(sessionAttributes.Drafts[intent.Slots["date"].Value])-1]),
								Directives: []interface{}{alexa.DialogDirective{Type: "Dialog.ElicitSlot", SlotToElicit: "text"}},
							},
							SessionAttributes: requestEnv.Session.Attributes,
						}
					case "korrigiere", "korrigieren":
						if len(sessionAttributes.Drafts[intent.Slots["date"].Value]) == 0 {
							return &alexa.ResponseEnvelope{Version: "1.0",
								Response: &alexa.Response{
									OutputSpeech: plainText("Dein Eintrag ist leer. Es gibt nichts zu korrigieren. Bitte verfasse zuerst den ersten Teil Deines Eintrags."),
									Directives:   []interface{}{alexa.DialogDirective{Type: "Dialog.ElicitSlot", SlotToElicit: "text"}},
								},
								SessionAttributes: requestEnv.Session.Attributes,
							}
						}
						sessionAttributes.Drafts[intent.Slots["date"].Value] = sessionAttributes.Drafts[intent.Slots["date"].Value][:len(sessionAttributes.Drafts[intent.Slots["date"].Value])-1]
						return &alexa.ResponseEnvelope{Version: "1.0",
							Response: &alexa.Response{
								OutputSpeech: plainText("OK. Bitte verfasse den letzten Teil Deines Eintrags erneut."),
								Directives:   []interface{}{alexa.DialogDirective{Type: "Dialog.ElicitSlot", SlotToElicit: "text"}},
							},
							SessionAttributes: mapStringInterfaceFrom(sessionAttributes),
						}
					case "abbrechen":
						sessionAttributes.Drafting = false
						return &alexa.ResponseEnvelope{Version: "1.0",
							Response: &alexa.Response{
								OutputSpeech: plainText("Okay. Abgebrochen.\n\nWas möchtest Du als nächstes in Deinem Tagebuch machen?"),
							},
							SessionAttributes: mapStringInterfaceFrom(sessionAttributes),
						}
					case "fertig":
						if len(sessionAttributes.Drafts[intent.Slots["date"].Value]) == 0 {
							sessionAttributes.Drafting = false
							return &alexa.ResponseEnvelope{Version: "1.0",
								Response:          &alexa.Response{OutputSpeech: plainText("Dein Eintrag ist leer. Es gibt nichts zu speichern. Was möchtest Du als nächstes tun?")},
								SessionAttributes: mapStringInterfaceFrom(sessionAttributes),
							}
						}
						return &alexa.ResponseEnvelope{Version: "1.0",
							Response: &alexa.Response{
								OutputSpeech: plainText("Alles klar. Ich habe folgenden Eintrag für das Datum " + intent.Slots["date"].Value + ": " +
									"\"" + strings.Join(sessionAttributes.Drafts[intent.Slots["date"].Value], ". ") + "\". Soll ich ihn so speichern?"),
								Directives: []interface{}{alexa.DialogDirective{Type: "Dialog.ConfirmIntent", UpdatedIntent: &intent}},
							},
							SessionAttributes: requestEnv.Session.Attributes,
						}
					}
				case "CONFIRMED":
					date, e := date.AutoParse(intent.Slots["date"].Value)
					util.PanicOnError(errors.Wrapf(e, "Could not convert string '%v' to date", intent.Slots["date"].Value))

					journal.AddEntry(date, strings.Join(sessionAttributes.Drafts[intent.Slots["date"].Value], ". "))

					sessionAttributes.Drafting = false
					delete(sessionAttributes.Drafts, intent.Slots["date"].Value)

					return &alexa.ResponseEnvelope{Version: "1.0",
						Response:          &alexa.Response{OutputSpeech: plainText("Okay. Gespeichert.\n\nWas möchtest Du als nächstes in Deinem Tagebuch machen?")},
						SessionAttributes: mapStringInterfaceFrom(sessionAttributes),
					}

				case "DENIED":
					sessionAttributes.Drafting = false
					return &alexa.ResponseEnvelope{Version: "1.0",
						Response:          &alexa.Response{OutputSpeech: plainText("Okay. Nicht gespeichert.\n\nWas möchtest Du als nächstes in Deinem Tagebuch machen?")},
						SessionAttributes: mapStringInterfaceFrom(sessionAttributes),
					}
				default:
					panic(errors.New("Invalid intent.ConfirmationStatus"))
				}
			default:
				panic(errors.New("Invalid requestEnv.Request.DialogState"))
			}
		case "ListAllEntriesInDate":
			switch requestEnv.Request.DialogState {
			case "STARTED", "IN_PROGRESS":
				return pureDelegate(&intent, requestEnv.Session.Attributes)
			case "COMPLETED":
				_, monthDate, dateType := DateFrom(intent.Slots["date"].Value, intent.Slots["year"].Value)
				if dateType == MonthDate {
					return listAllEntriesInDate(&journal, monthDate, requestEnv.Session.Attributes, h.errorInterpreter)
				}
				return &alexa.ResponseEnvelope{Version: "1.0",
					Response: &alexa.Response{
						OutputSpeech: plainText(fmt.Sprintf("Ich habe Dich nicht richtig verstanden. Kannst Du es bitte noch einmal versuchen?")),
					},
					SessionAttributes: requestEnv.Session.Attributes,
				}
			default:
				panic(errors.New("Invalid requestEnv.Request.DialogState"))
			}
		case "ReadAllEntriesInDate", "ReadExistingEntryAbsoluteDateIntent":
			switch requestEnv.Request.DialogState {
			case "STARTED", "IN_PROGRESS":
				return pureDelegate(&intent, requestEnv.Session.Attributes)
			case "COMPLETED":
				entryDate, monthDate, dateType := DateFrom(intent.Slots["date"].Value, intent.Slots["year"].Value)
				if dateType == MonthDate {
					return readAllEntriesInDate(&journal, monthDate, requestEnv.Session.Attributes, h.errorInterpreter)
				}
				if dateType != DayDate {
					return &alexa.ResponseEnvelope{Version: "1.0",
						Response: &alexa.Response{
							OutputSpeech: plainText(fmt.Sprintf("Ich habe Dich nicht richtig verstanden. Bitte versuche es noch einmal. Sage z.B. \"was war im Juni 1997?\"")),
						},
						SessionAttributes: requestEnv.Session.Attributes,
					}
				}

				text, e := journal.GetEntry(entryDate)
				if e != nil {
					return plainTextRespEnv("Oje. Beim Abrufen des Eintrags ist ein Fehler aufgetreten. "+h.errorInterpreter.Interpret(e),
						requestEnv.Session.Attributes)
				}
				if text != "" {
					return &alexa.ResponseEnvelope{Version: "1.0",
						Response: &alexa.Response{
							OutputSpeech: plainText(fmt.Sprintf("Hier ist der Eintrag vom %v, %v: %v.\n\nWas möchtest Du als nächstes in Deinem Tagebuch machen?",
								weekdays[entryDate.Weekday().String()], entryDate, text)),
						},
						SessionAttributes: requestEnv.Session.Attributes,
					}
				}
				closestEntry, e := journal.GetClosestEntry(entryDate)
				if e != nil {
					return plainTextRespEnv("Oje. Beim Abrufen des Eintrags ist ein Fehler aufgetreten. "+h.errorInterpreter.Interpret(e),
						requestEnv.Session.Attributes)
				}
				if closestEntry == (j.Entry{}) {
					return &alexa.ResponseEnvelope{Version: "1.0",
						Response:          &alexa.Response{OutputSpeech: plainText(fmt.Sprintf("Dein Tagebuch ist noch leer.\n\nWas möchtest Du als nächstes in Deinem Tagebuch machen? Sage z.B. neuen Eintrag erstellen."))},
						SessionAttributes: requestEnv.Session.Attributes,
					}
				}
				return &alexa.ResponseEnvelope{Version: "1.0",
					Response: &alexa.Response{
						OutputSpeech: plainText(fmt.Sprintf("Ich habe fuer den %v keinen Eintrag gefunden. "+
							"Der nächste Eintrag ist vom %v, %v. Er lautet: %v.\n\nWas möchtest Du als nächstes in Deinem Tagebuch machen?",
							entryDate, weekdays[closestEntry.EntryDate.Weekday().String()], closestEntry.EntryDate, closestEntry.EntryText)),
					},
					SessionAttributes: requestEnv.Session.Attributes,
				}
			default:
				panic(errors.New("Invalid requestEnv.Request.DialogState"))
			}
		case "ReadExistingEntryRelativeDateIntent":
			today := date.NewAt(time.Now())
			x, e := strconv.Atoi(intent.Slots["number"].Value)
			if e != nil ||
				intent.Slots["unit"].Resolutions.ResolutionsPerAuthority[0].Status["code"] == "ER_SUCCESS_NO_MATCH" {
				return &alexa.ResponseEnvelope{Version: "1.0",
					Response: &alexa.Response{
						OutputSpeech: plainText("Das habe ich leider nicht verstanden. Kannst du es bitte noch einmal versuchen? Sage z.B. was war heute vor einem Jahr?"),
						Directives:   []interface{}{alexa.DialogDirective{Type: "Dialog.ElicitSlot", SlotToElicit: "unit"}},
					},
					SessionAttributes: mapStringInterfaceFrom(sessionAttributes),
				}
			}
			var entryDate date.Date
			switch intent.Slots["unit"].Resolutions.ResolutionsPerAuthority[0].Values[0].Value.ID {
			case "DAYS":
				entryDate = today.AddDate(0, 0, -x)
			case "MONTHS":
				entryDate = today.AddDate(0, -x, 0)
			case "YEARS":
				entryDate = today.AddDate(-x, 0, 0)
			default:
				panic(errors.New("Invalid resolution"))
			}

			text, e := journal.GetEntry(entryDate)
			if e != nil {
				return plainTextRespEnv("Oje. Beim Abrufen des Eintrags ist ein Fehler aufgetreten. "+h.errorInterpreter.Interpret(e),
					requestEnv.Session.Attributes)
			}
			if text != "" {
				return &alexa.ResponseEnvelope{Version: "1.0",
					Response: &alexa.Response{
						OutputSpeech: plainText(fmt.Sprintf("Hier ist der Eintrag vom %v, %v: %v.\n\nWas möchtest Du als nächstes in Deinem Tagebuch machen?",
							weekdays[entryDate.Weekday().String()], entryDate, text)),
					},
					SessionAttributes: requestEnv.Session.Attributes,
				}
			}
			closestEntry, e := journal.GetClosestEntry(entryDate)
			if e != nil {
				return plainTextRespEnv("Oje. Beim Abrufen des Eintrags ist ein Fehler aufgetreten. "+h.errorInterpreter.Interpret(e),
					requestEnv.Session.Attributes)
			}
			if closestEntry == (j.Entry{}) {
				return &alexa.ResponseEnvelope{Version: "1.0",
					Response:          &alexa.Response{OutputSpeech: plainText(fmt.Sprintf("Dein Tagebuch ist noch leer.\n\nWas möchtest Du als nächstes in Deinem Tagebuch machen? Sage z.B. neuen Eintrag erstellen."))},
					SessionAttributes: requestEnv.Session.Attributes,
				}
			}
			return &alexa.ResponseEnvelope{Version: "1.0",
				Response: &alexa.Response{
					OutputSpeech: plainText(fmt.Sprintf("Ich habe fuer den %v keinen Eintrag gefunden. "+
						"Der nächste Eintrag ist vom %v, %v. Er lautet: %v.\n\nWas möchtest Du als nächstes in Deinem Tagebuch machen?",
						entryDate, weekdays[closestEntry.EntryDate.Weekday().String()], closestEntry.EntryDate, closestEntry.EntryText)),
				},
				SessionAttributes: requestEnv.Session.Attributes,
			}
		case "SearchIntent":
			entries, e := journal.SearchFor(intent.Slots["query"].Value)
			if e != nil {
				return plainTextRespEnv("Oje. Beim Suchen nach Eintraegen ist ein Fehler aufgetreten. "+h.errorInterpreter.Interpret(e),
					requestEnv.Session.Attributes)
			}
			if len(entries) == 0 {
				return &alexa.ResponseEnvelope{Version: "1.0",
					Response: &alexa.Response{
						OutputSpeech: plainText(fmt.Sprintf("Keine Einträge für die Suche \"%v\" gefunden.\n\nWas möchtest Du als nächstes in Deinem Tagebuch machen?", intent.Slots["query"].Value)),
					},
					SessionAttributes: requestEnv.Session.Attributes,
				}
			}
			text := fmt.Sprintf("Hier sind die Ergebnisse für die Suche \"%v\": ", intent.Slots["query"].Value)
			for _, entry := range entries {

				tuple := weekdays[entry.EntryDate.Weekday().String()] + ", " + entry.EntryDate.String() + ": " + strings.TrimRight(entry.EntryText, ". ") + ". "
				if len(text)+len(tuple) > responseTextLimit {
					break
				}
				text += tuple
			}
			return &alexa.ResponseEnvelope{Version: "1.0",
				Response:          &alexa.Response{OutputSpeech: plainText(text + "\n\nWas möchtest Du als nächstes in Deinem Tagebuch machen?")},
				SessionAttributes: requestEnv.Session.Attributes,
			}
		case "DeleteEntryIntent":
			switch requestEnv.Request.DialogState {
			case "STARTED":
				return pureDelegate(&intent, requestEnv.Session.Attributes)
			case "IN_PROGRESS":
				switch intent.ConfirmationStatus {
				case "NONE":
					if intent.Slots["date"].Value == "" {
						return pureDelegate(&intent, requestEnv.Session.Attributes)
					}
					date, e := date.AutoParse(intent.Slots["date"].Value)
					util.PanicOnError(errors.Wrapf(e, "Could not convert string '%v' to date", intent.Slots["date"].Value))

					entry, e := journal.GetEntry(date)
					if e != nil {
						return plainTextRespEnv("Oje. Beim Aufrufen des zu loeschenden Eintrags ist ein Fehler aufgetreten. "+h.errorInterpreter.Interpret(e),
							requestEnv.Session.Attributes)
					}
					if entry == "" {
						return plainTextRespEnv("Hm. Zu diesem Datum habe ich leider keinen Eintrag gefunden.", requestEnv.Session.Attributes)
					}
					return &alexa.ResponseEnvelope{Version: "1.0",
						Response: &alexa.Response{
							OutputSpeech: plainText("Du moechtest den folgenden Eintrag loeschen: " + entry + ". Soll ich ihn wirklich loeschen?"),
							Directives:   []interface{}{alexa.DialogDirective{Type: "Dialog.ConfirmIntent", UpdatedIntent: &intent}},
						},
						SessionAttributes: requestEnv.Session.Attributes,
					}
				case "CONFIRMED":
					date, e := date.AutoParse(intent.Slots["date"].Value)
					util.PanicOnError(errors.Wrapf(e, "Could not convert string '%v' to date", intent.Slots["date"].Value))

					e = journal.DeleteEntry(date)
					if e != nil {
						return plainTextRespEnv("Oje. Beim Loeschen des Eintrags ist ein Fehler aufgetreten. "+h.errorInterpreter.Interpret(e),
							requestEnv.Session.Attributes)
					}

					return plainTextRespEnv("Okay. Geloescht.\n\nWas möchtest Du als nächstes in Deinem Tagebuch machen?", requestEnv.Session.Attributes)
				case "DENIED":
					return plainTextRespEnv("Okay. Nicht geloescht.\n\nWas möchtest Du als nächstes in Deinem Tagebuch machen?", requestEnv.Session.Attributes)
				default:
					panic(errors.New("Invalid intent.ConfirmationStatus"))
				}
			default:
				panic(errors.New("Invalid requestEnv.Request.DialogState"))
			}

		case "AMAZON.HelpIntent":
			return &alexa.ResponseEnvelope{Version: "1.0",
				Response:          &alexa.Response{OutputSpeech: plainText(helpText)},
				SessionAttributes: requestEnv.Session.Attributes,
			}
		case "AMAZON.CancelIntent", "AMAZON.StopIntent":
			return &alexa.ResponseEnvelope{Version: "1.0",
				Response:          &alexa.Response{ShouldSessionEnd: true},
				SessionAttributes: requestEnv.Session.Attributes,
			}
		default:
			panic(errors.New("Invalid Intent"))
		}

	case "SessionEndedRequest":
		return &alexa.ResponseEnvelope{Version: "1.0"}

	default:
		panic(errors.New("Invalid Request"))
	}
}

func listAllEntriesInDate(journal *j.Journal, dateSlotValue string, sessionAttributes map[string]interface{}, errorInterpreter ErrorInterpreter) *alexa.ResponseEnvelope {
	entries, e := journal.GetEntries(dateSlotValue[:7])
	if e != nil {
		return plainTextRespEnv("Oje. Beim Abrufen der Eintraege ist ein Fehler aufgetreten. "+errorInterpreter.Interpret(e),
			sessionAttributes)
	}
	if len(entries) == 0 {
		return &alexa.ResponseEnvelope{Version: "1.0",
			Response: &alexa.Response{
				OutputSpeech: plainText(fmt.Sprintf("Keine Einträge für den Zeitraum " + readableStringFrom(dateSlotValue) + " gefunden.\n\nWas möchtest Du als nächstes in Deinem Tagebuch machen?")),
			},
			SessionAttributes: sessionAttributes,
		}
	}
	var tuples []string
	for _, entry := range entries {
		tuples = append(tuples, weekdays[entry.EntryDate.Weekday().String()]+", "+entry.EntryDate.String()+": "+entry.EntryText)
	}
	return &alexa.ResponseEnvelope{Version: "1.0",
		Response: &alexa.Response{
			OutputSpeech: plainText(fmt.Sprintf("Hier sind die Einträge für den Zeitraum " + readableStringFrom(dateSlotValue) + ": " +
				strings.Join(tuples, ". ") + "\n\nWas möchtest Du als nächstes in Deinem Tagebuch machen?")),
		},
		SessionAttributes: sessionAttributes,
	}
}

func readAllEntriesInDate(journal *j.Journal, dateSlotValue string, sessionAttributes map[string]interface{}, errorInterpreter ErrorInterpreter) *alexa.ResponseEnvelope {
	entries, e := journal.GetEntries(dateSlotValue[:7])
	if e != nil {
		return plainTextRespEnv("Oje. Beim Abrufen der Eintraege ist ein Fehler aufgetreten. "+errorInterpreter.Interpret(e),
			sessionAttributes)
	}
	if len(entries) == 0 {
		return &alexa.ResponseEnvelope{Version: "1.0",
			Response: &alexa.Response{
				OutputSpeech: plainText(fmt.Sprintf("Keine Einträge für den Zeitraum " + readableStringFrom(dateSlotValue) + " gefunden.\n\nWas möchtest Du als nächstes in Deinem Tagebuch machen?")),
			},
			SessionAttributes: sessionAttributes,
		}
	}
	var tuples []string
	for _, entry := range entries {
		tuples = append(tuples, weekdays[entry.EntryDate.Weekday().String()]+", "+entry.EntryDate.String()+": "+entry.EntryText)
	}
	return &alexa.ResponseEnvelope{Version: "1.0",
		Response: &alexa.Response{
			OutputSpeech: plainText(fmt.Sprintf("Hier sind die Einträge für den Zeitraum " + readableStringFrom(dateSlotValue) + ": " +
				strings.Join(tuples, ". ") + "\n\nWas möchtest Du als nächstes in Deinem Tagebuch machen?")),
		},
		SessionAttributes: sessionAttributes,
	}
}

func readableStringFrom(dateLike string) string {
	r := regexp.MustCompile(`(\d{4})-(\d{2})(-XX)?`)
	if matched := r.MatchString(dateLike); matched {
		subMatches := r.FindStringSubmatch(dateLike)
		yearString := subMatches[1]
		monthString := subMatches[2]
		month, e := strconv.Atoi(monthString)
		if e != nil {
			panic(e)
		}
		return monthsReverse[month] + " " + yearString
	}
	return dateLike
}

func plainText(text string) *alexa.OutputSpeech {
	return &alexa.OutputSpeech{Type: "PlainText", Text: text}
}

func plainTextRespEnv(text string, attributes map[string]interface{}) *alexa.ResponseEnvelope {
	return &alexa.ResponseEnvelope{Version: "1.0",
		Response:          &alexa.Response{OutputSpeech: plainText(text)},
		SessionAttributes: attributes,
	}
}

func pureDelegate(intent *alexa.Intent, sessionAttributes map[string]interface{}) *alexa.ResponseEnvelope {
	return &alexa.ResponseEnvelope{Version: "1.0",
		Response: &alexa.Response{
			Directives: []interface{}{
				alexa.DialogDirective{
					Type:          "Dialog.Delegate",
					UpdatedIntent: intent,
				},
			},
		},
		SessionAttributes: sessionAttributes,
	}
}

func internalError() *alexa.ResponseEnvelope {
	return &alexa.ResponseEnvelope{Version: "1.0",
		Response: &alexa.Response{
			OutputSpeech:     plainText("Es ist ein interner Fehler aufgetreten. Ich habe den Entwickler bereits informiert, er wird sich um das Problem kümmern. Bitte versuche es zu einem späteren Zeitpunkt noch einmal."),
			ShouldSessionEnd: true,
		},
	}
}

func mapStringInterfaceFrom(sessionAttributes SessionAttributes) map[string]interface{} {
	sessionAttributesBuf, e := json.Marshal(sessionAttributes)
	util.PanicOnError(errors.Wrap(e, "Could not marshal sessionAttributes"))

	var sessionAttributesMap map[string]interface{}
	e = json.Unmarshal(sessionAttributesBuf, &sessionAttributesMap)
	util.PanicOnError(errors.Wrap(e, "Could not unmarshal sessionAttributesBuf"))

	return sessionAttributesMap
}
