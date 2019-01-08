package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/petergtz/alexa-journal/drive"
	j "github.com/petergtz/alexa-journal/journal"
	"github.com/petergtz/alexa-journal/tsv"
	"github.com/petergtz/alexa-journal/util"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"
	"github.com/rickb777/date"

	"github.com/petergtz/go-alexa"
	"go.uber.org/zap"
)

var (
	log *zap.SugaredLogger
)

type TSVDriveFileJournalProvider struct {
	Log *zap.SugaredLogger
}

func (jp *TSVDriveFileJournalProvider) Get(accessToken string) j.Journal {
	return j.Journal{
		Data: &tsv.TextFileBackedTabularData{
			TextFileLoader: drive.NewFileService(accessToken, "my-journal.tsv", jp.Log),
		},
	}
}

func main() {
	l, e := zap.NewDevelopment()
	if e != nil {
		panic(e)
	}
	defer l.Sync()
	log = l.Sugar()

	handler := &alexa.Handler{
		Skill: &JournalSkill{
			log:             log,
			journalProvider: &TSVDriveFileJournalProvider{Log: log},
		},
		Log: log,
		ExpectedApplicationID: os.Getenv("APPLICATION_ID"),
	}
	http.HandleFunc("/", handler.Handle)
	port := os.Getenv("PORT")
	if port == "" { // the port variable lets us distinguish between a local server an done in CF
		log.Debugf("Certificate path: %v", os.Getenv("cert"))
		log.Debugf("Private key path: %v", os.Getenv("key"))
		e = http.ListenAndServeTLS("0.0.0.0:4443", os.Getenv("cert"), os.Getenv("key"), nil)
	} else {
		e = http.ListenAndServe("0.0.0.0:"+port, nil)
	}
	log.Fatal(e)
}

type JournalSkill struct {
	journalProvider j.JournalProvider
	log             *zap.SugaredLogger
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
			h.log.Errorw("Internal Server Error.", "error", fmt.Sprintf("%+v", e))
			responseEnv = internalError()
		}
	}()

	log := h.log.With("request", requestEnv.Request, "SessionAttributes", requestEnv.Session.Attributes)
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
		return &alexa.ResponseEnvelope{Version: "1.0",
			Response:          &alexa.Response{OutputSpeech: plainText("Dein Tagebuch ist nun geöffnet. Was möchtest Du tun?")},
			SessionAttributes: requestEnv.Session.Attributes,
		}

	case "IntentRequest":
		journal := h.journalProvider.Get(requestEnv.Session.User.AccessToken)
		// journal := j.Journal{Content: file.Content()}
		log.Debugw("Journal downloaded")

		var sessionAttributes SessionAttributes
		sessionAttributes.Drafts = make(map[string][]string)
		e := mapstructure.Decode(requestEnv.Session.Attributes, &sessionAttributes)
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
						return &alexa.ResponseEnvelope{Version: "1.0",
							Response: &alexa.Response{
								OutputSpeech: plainText("Du kannst Deinen eintrag nun verfassen; ich werde jeden Teil kurz bestaetigen, sodass du die moeglichkeit hast ihn zu \"korrigieren\" oder \"anzuhoeren\". Sage \"fertig\", wenn Du fertig bist."),
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
					case "fertig":
						if len(sessionAttributes.Drafts[intent.Slots["date"].Value]) == 0 {
							sessionAttributes.Drafting = false
							return &alexa.ResponseEnvelope{Version: "1.0",
								Response:          &alexa.Response{OutputSpeech: plainText("Dein Eintrag ist leer. Es gibt nichts zu speichern.")},
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
						Response:          &alexa.Response{OutputSpeech: plainText("Okay. Gespeichert.")},
						SessionAttributes: mapStringInterfaceFrom(sessionAttributes),
					}

				case "DENIED":
					sessionAttributes.Drafting = false
					return &alexa.ResponseEnvelope{Version: "1.0",
						Response:          &alexa.Response{OutputSpeech: plainText("Okay. Nicht gespeichert.")},
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
				if matched, e := regexp.MatchString(`\d{4}-\d{2}(-XX)?`, intent.Slots["date"].Value); e == nil && matched {
					entries := journal.GetEntries(intent.Slots["date"].Value[:7])
					if len(entries) == 0 {
						return &alexa.ResponseEnvelope{Version: "1.0",
							Response: &alexa.Response{
								OutputSpeech: plainText(fmt.Sprintf("Keine Einträge für den Zeitraum " + readableStringFrom(intent.Slots["date"].Value) + " gefunden.")),
							},
							SessionAttributes: requestEnv.Session.Attributes,
						}
					}
					var dates []string
					for _, entry := range entries {
						dates = append(dates, entry.EntryDate.String())
					}
					return &alexa.ResponseEnvelope{Version: "1.0",
						Response: &alexa.Response{
							OutputSpeech: plainText(fmt.Sprintf("Folgende Einträge habe ich für den Zeitraum " + readableStringFrom(intent.Slots["date"].Value) + " gefunden: " +
								strings.Join(dates, ". "))),
						},
						SessionAttributes: requestEnv.Session.Attributes,
					}
				}
				return &alexa.ResponseEnvelope{Version: "1.0",
					Response: &alexa.Response{
						OutputSpeech: plainText(fmt.Sprintf("Ich habe Dich nicht richtig verstanden.")),
					},
					SessionAttributes: requestEnv.Session.Attributes,
				}
			default:
				panic(errors.New("Invalid requestEnv.Request.DialogState"))
			}
		case "ReadAllEntriesInDate":
			switch requestEnv.Request.DialogState {
			case "STARTED", "IN_PROGRESS":
				return pureDelegate(&intent, requestEnv.Session.Attributes)
			case "COMPLETED":
				if matched, e := regexp.MatchString(`\d{4}-\d{2}(-XX)?`, intent.Slots["date"].Value); e == nil && matched {
					entries := journal.GetEntries(intent.Slots["date"].Value[:7])
					if len(entries) == 0 {
						return &alexa.ResponseEnvelope{Version: "1.0",
							Response: &alexa.Response{
								OutputSpeech: plainText(fmt.Sprintf("Keine Einträge für den Zeitraum " + readableStringFrom(intent.Slots["date"].Value) + " gefunden.")),
							},
							SessionAttributes: requestEnv.Session.Attributes,
						}
					}
					var tuples []string
					for _, entry := range entries {
						tuples = append(tuples, weekdays[entry.EntryDate.Weekday().String()]+", "+entry.EntryDate.String()+": "+entry.EntryText)
					}
					return &alexa.ResponseEnvelope{Version: "1.0",
						Response: &alexa.Response{
							OutputSpeech: plainText(fmt.Sprintf("Hier sind die Einträge für den Zeitraum " + readableStringFrom(intent.Slots["date"].Value) + ": " +
								strings.Join(tuples, ". "))),
						},
						SessionAttributes: requestEnv.Session.Attributes,
					}
				}
				return &alexa.ResponseEnvelope{Version: "1.0",
					Response: &alexa.Response{
						OutputSpeech: plainText(fmt.Sprintf("Ich habe Dich nicht richtig verstanden.")),
					},
					SessionAttributes: requestEnv.Session.Attributes,
				}
			default:
				panic(errors.New("Invalid requestEnv.Request.DialogState"))
			}

		case "ReadExistingEntryAbsoluteDateIntent":
			switch requestEnv.Request.DialogState {
			case "STARTED", "IN_PROGRESS":
				return pureDelegate(&intent, requestEnv.Session.Attributes)
			case "COMPLETED":
				entryDate, e := date.AutoParse(intent.Slots["date"].Value)
				if intent.Slots["year"].Value != "" {
					entryDate, e = date.AutoParse(intent.Slots["year"].Value + intent.Slots["date"].Value[4:])
				}
				util.PanicOnError(errors.Wrapf(e, "Could not convert string '%v' to date", intent.Slots["date"].Value))

				text := journal.GetEntry(entryDate)
				if text != "" {
					return &alexa.ResponseEnvelope{Version: "1.0",
						Response: &alexa.Response{
							OutputSpeech: plainText(fmt.Sprintf("Hier ist der Eintrag vom %v, %v: %v.",
								weekdays[entryDate.Weekday().String()], entryDate, text)),
						},
						SessionAttributes: requestEnv.Session.Attributes,
					}
				}
				closestEntry := journal.GetClosestEntry(entryDate)
				if closestEntry == (j.Entry{}) {
					return &alexa.ResponseEnvelope{Version: "1.0",
						Response:          &alexa.Response{OutputSpeech: plainText(fmt.Sprintf("Dein Tagebuch ist noch leer."))},
						SessionAttributes: requestEnv.Session.Attributes,
					}
				}
				return &alexa.ResponseEnvelope{Version: "1.0",
					Response: &alexa.Response{
						OutputSpeech: plainText(fmt.Sprintf("Ich habe fuer den %v keinen Eintrag gefunden. "+
							"Der nächste Eintrag ist vom %v, %v. Er lautet: %v.",
							entryDate, weekdays[closestEntry.EntryDate.Weekday().String()], closestEntry.EntryDate, closestEntry.EntryText)),
					},
					SessionAttributes: requestEnv.Session.Attributes,
				}
			default:
				panic(errors.New("Invalid requestEnv.Request.DialogState"))
			}
		case "ReadExistingEntryRelativeDateIntent":
			switch requestEnv.Request.DialogState {
			case "STARTED", "IN_PROGRESS":
				return pureDelegate(&intent, requestEnv.Session.Attributes)
			case "COMPLETED":
				today := date.NewAt(time.Now())
				x, e := strconv.Atoi(intent.Slots["number"].Value)
				util.PanicOnError(errors.Wrapf(e, "Could not convert string '%v' to date", intent.Slots["date"].Value))
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

				text := journal.GetEntry(entryDate)
				if text != "" {
					return &alexa.ResponseEnvelope{Version: "1.0",
						Response: &alexa.Response{
							OutputSpeech: plainText(fmt.Sprintf("Hier ist der Eintrag vom %v, %v: %v.",
								weekdays[entryDate.Weekday().String()], entryDate, text)),
						},
						SessionAttributes: requestEnv.Session.Attributes,
					}
				}
				closestEntry := journal.GetClosestEntry(entryDate)
				return &alexa.ResponseEnvelope{Version: "1.0",
					Response: &alexa.Response{
						OutputSpeech: plainText(fmt.Sprintf("Ich habe fuer den %v keinen Eintrag gefunden. "+
							"Der nächste Eintrag ist vom %v, %v. Er lautet: %v.",
							entryDate, weekdays[closestEntry.EntryDate.Weekday().String()], closestEntry.EntryDate, closestEntry.EntryText)),
					},
					SessionAttributes: requestEnv.Session.Attributes,
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
	return nil // dummy return to satisfy compiler
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
			OutputSpeech:     plainText("Es ist ein interner Fehler aufgetreten."),
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
