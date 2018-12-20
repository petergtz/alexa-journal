package main

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rickb777/date"

	"github.com/petergtz/alexa-journal/drive"
	"github.com/petergtz/go-alexa"
	"go.uber.org/zap"
)

var (
	log *zap.SugaredLogger
)

func main() {
	l, e := zap.NewDevelopment()
	if e != nil {
		panic(e)
	}
	defer l.Sync()
	log = l.Sugar()

	handler := &alexa.Handler{
		Skill: &JournalSkill{
			log: log,
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
	journalProvider *journaldrive.JournalProvider
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

func (h *JournalSkill) ProcessRequest(requestEnv *alexa.RequestEnvelope) *alexa.ResponseEnvelope {
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
		}
	}

	switch requestEnv.Request.Type {

	case "LaunchRequest":
		return &alexa.ResponseEnvelope{Version: "1.0",
			Response: &alexa.Response{
				OutputSpeech: plainText("Dein Tagebuch ist nun geöffnet. Was möchtest Du tun?"),
			},
		}

	case "IntentRequest":
		journal := h.journalProvider.Get(requestEnv.Session.User.AccessToken)
		log.Debugw("Journal downloaded")

		intent := requestEnv.Request.Intent
		switch intent.Name {
		case "NewEntryIntent":

			switch requestEnv.Request.DialogState {
			case "STARTED":
				return pureDelegate(&intent)
			case "IN_PROGRESS":
				return pureDelegate(&intent)
			case "COMPLETED":
				switch intent.ConfirmationStatus {
				case "NONE":
					return &alexa.ResponseEnvelope{Version: "1.0",
						Response: &alexa.Response{
							OutputSpeech: plainText("Alles klar. Ich habe folgenden Eintrag für das Datum " + intent.Slots["date"].Value + ": " +
								"\"" + intent.Slots["text"].Value + "\". Soll ich ihn so speichern?"),
							Directives: []interface{}{
								alexa.DialogDirective{
									Type:          "Dialog.ConfirmIntent",
									UpdatedIntent: &intent,
								},
							},
						},
					}
				case "CONFIRMED":
					date, e := date.AutoParse(intent.Slots["date"].Value)
					if e != nil {
						log.Errorw("Could not convert string to date", "date", intent.Slots["date"].Value, e)
						return internalError()
					}
					e = journal.AddEntry(date, intent.Slots["text"].Value)
					if e != nil {
						log.Errorw("Could not add entry", "date", date, "text", intent.Slots["text"].Value, "error", e)
						return internalError()
					}

					return &alexa.ResponseEnvelope{Version: "1.0",
						Response: &alexa.Response{
							OutputSpeech:     plainText("Okay. Gespeichert."),
							ShouldSessionEnd: true,
						},
					}

				case "DENIED":
					return &alexa.ResponseEnvelope{Version: "1.0",
						Response: &alexa.Response{OutputSpeech: plainText("Okay. Wurde verworfen.")},
					}
				default:
					return internalError()
				}
			default:
				return internalError()
			}
		case "ListAllEntriesInDate":
			switch requestEnv.Request.DialogState {
			case "STARTED", "IN_PROGRESS":
				return pureDelegate(&intent)
			case "COMPLETED":
				if matched, e := regexp.MatchString(`\d{4}-\d{2}(-XX)?`, intent.Slots["date"].Value); e == nil && matched {
					entries, e := journal.GetEntries(intent.Slots["date"].Value[:7])
					if e != nil {
						log.Errorw("Could not get entries", "timeRange", intent.Slots["date"].Value[:7], "error", e)
						return internalError()
					}
					if len(entries) == 0 {
						return &alexa.ResponseEnvelope{Version: "1.0",
							Response: &alexa.Response{
								OutputSpeech: plainText(fmt.Sprintf("Keine Einträge für den Zeitraum " + strings.Replace(intent.Slots["date"].Value[:7], "-", "/", -1) + " gefunden.")),
							},
						}
					}
					var dates []string
					for _, entry := range entries {
						dates = append(dates, entry.EntryDate)
					}
					return &alexa.ResponseEnvelope{Version: "1.0",
						Response: &alexa.Response{
							OutputSpeech: plainText(fmt.Sprintf("Folgende Einträge habe ich für den Zeitraum " + strings.Replace(intent.Slots["date"].Value[:7], "-", "/", -1) + " gefunden: " +
								strings.Join(dates, ". "))),
						},
					}
				}
				return &alexa.ResponseEnvelope{Version: "1.0",
					Response: &alexa.Response{
						OutputSpeech: plainText(fmt.Sprintf("Ich habe Dich nicht richtig verstanden.")),
					},
				}
			default:
				return internalError()
			}
		case "ReadAllEntriesInDate":
			switch requestEnv.Request.DialogState {
			case "STARTED", "IN_PROGRESS":
				return pureDelegate(&intent)
			case "COMPLETED":
				if matched, e := regexp.MatchString(`\d{4}-\d{2}(-XX)?`, intent.Slots["date"].Value); e == nil && matched {
					entries, e := journal.GetEntries(intent.Slots["date"].Value[:7])
					if e != nil {
						log.Errorw("Could not get entries", "timeRange", intent.Slots["date"].Value[:7], "error", e)
						return internalError()
					}
					if len(entries) == 0 {
						return &alexa.ResponseEnvelope{Version: "1.0",
							Response: &alexa.Response{
								OutputSpeech: plainText(fmt.Sprintf("Keine Einträge für den Zeitraum " + strings.Replace(intent.Slots["date"].Value[:7], "-", "/", -1) + " gefunden.")),
							},
						}
					}
					var tuples []string
					for _, entry := range entries {
						tuples = append(tuples, entry.EntryDate+": "+entry.EntryText)
					}
					return &alexa.ResponseEnvelope{Version: "1.0",
						Response: &alexa.Response{
							OutputSpeech: plainText(fmt.Sprintf("Hier sind die Einträge für den Zeitraum " + strings.Replace(intent.Slots["date"].Value[:7], "-", "/", -1) + ": " +
								strings.Join(tuples, ". "))),
						},
					}
				}
				return &alexa.ResponseEnvelope{Version: "1.0",
					Response: &alexa.Response{
						OutputSpeech: plainText(fmt.Sprintf("Ich habe Dich nicht richtig verstanden.")),
					},
				}
			default:
				return internalError()
			}

		case "ReadExistingEntryAbsoluteDateIntent":
			switch requestEnv.Request.DialogState {
			case "STARTED", "IN_PROGRESS":
				return pureDelegate(&intent)
			case "COMPLETED":
				entryDate, e := date.AutoParse(intent.Slots["date"].Value)
				if intent.Slots["year"].Value != "" {
					entryDate, e = date.AutoParse(intent.Slots["year"].Value + intent.Slots["date"].Value[4:])
				}
				if e != nil {
					log.Errorw("Could not convert string to date", "date", intent.Slots["date"].Value, e)
					return internalError()
				}

				text, e := journal.GetEntry(entryDate)
				if e != nil {
					log.Errorw("Could not get entry", "date", entryDate, "error", e)
					return internalError()
				}
				if text != "" {
					return &alexa.ResponseEnvelope{Version: "1.0",
						Response: &alexa.Response{
							OutputSpeech: plainText(fmt.Sprintf("Hier ist der Eintrag vom %v.%v.%v: %v.",
								entryDate.Day(), int(entryDate.Month()), entryDate.Year(), text)),
						},
					}
				}
				entry, e := journal.GetClosestEntry(entryDate)
				return &alexa.ResponseEnvelope{Version: "1.0",
					Response: &alexa.Response{
						OutputSpeech: plainText(fmt.Sprintf("Ich habe fuer den %v.%v.%v keinen Eintrag gefunden. "+
							"Der nächste Eintrag ist vom %v. Er lautet: %v.",
							entryDate.Day(), int(entryDate.Month()), entryDate.Year(), entry.EntryDate, entry.EntryText)),
					},
				}
			default:
				return internalError()
			}
		case "ReadExistingEntryRelativeDateIntent":
			switch requestEnv.Request.DialogState {
			case "STARTED", "IN_PROGRESS":
				return pureDelegate(&intent)
			case "COMPLETED":
				today := date.NewAt(time.Now())
				x, e := strconv.Atoi(intent.Slots["number"].Value)
				if e != nil {
					log.Errorw("Could not convert string to date", "date", intent.Slots["date"].Value, e)
					return internalError()
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
					return internalError()
				}

				text, e := journal.GetEntry(entryDate)
				if e != nil {
					log.Errorw("Could not get entry", "date", entryDate, e)
					return internalError()
				}

				if text != "" {
					return &alexa.ResponseEnvelope{Version: "1.0",
						Response: &alexa.Response{
							OutputSpeech: plainText(fmt.Sprintf("Hier ist der Eintrag vom %v.%v.%v: %v.",
								entryDate.Day(), int(entryDate.Month()), entryDate.Year(), text)),
						},
					}
				}
				entry, e := journal.GetClosestEntry(entryDate)
				return &alexa.ResponseEnvelope{Version: "1.0",
					Response: &alexa.Response{
						OutputSpeech: plainText(fmt.Sprintf("Ich habe fuer den %v.%v.%v keinen Eintrag gefunden. "+
							"Der nächste Eintrag ist vom %v. Er lautet: %v.",
							entryDate.Day(), int(entryDate.Month()), entryDate.Year(), entry.EntryDate, entry.EntryText)),
					},
				}
			default:
				return internalError()
			}
		case "AMAZON.HelpIntent":
			return &alexa.ResponseEnvelope{Version: "1.0",
				Response: &alexa.Response{OutputSpeech: plainText(helpText)},
			}
		case "AMAZON.CancelIntent", "AMAZON.StopIntent":
			return &alexa.ResponseEnvelope{Version: "1.0",
				Response: &alexa.Response{ShouldSessionEnd: true},
			}
		default:
			return internalError()
		}

	case "SessionEndedRequest":
		return &alexa.ResponseEnvelope{Version: "1.0"}

	default:
		return internalError()
	}
}

func plainText(text string) *alexa.OutputSpeech {
	return &alexa.OutputSpeech{Type: "PlainText", Text: text}
}

func pureDelegate(intent *alexa.Intent) *alexa.ResponseEnvelope {
	return &alexa.ResponseEnvelope{Version: "1.0",
		Response: &alexa.Response{
			Directives: []interface{}{
				alexa.DialogDirective{
					Type:          "Dialog.Delegate",
					UpdatedIntent: intent,
				},
			},
		},
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
