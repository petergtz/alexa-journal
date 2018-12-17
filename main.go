package main

import (
	"fmt"
	"net/http"
	"os"
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

type Journal interface {
	AddEntry(date time.Time, text string) error
	GetEntry(date time.Time) (string, error)
}

const helpText = "Dieser Hilfetext fehlt leider noch"

// var months = map[string]string{
// 	"january":   "01",
// 	"february":  "02",
// 	"march":     "03",
// 	"april":     "04",
// 	"may":       "05",
// 	"june":      "06",
// 	"july":      "07",
// 	"august":    "08",
// 	"september": "09",
// 	"october":   "10",
// 	"november":  "11",
// 	"december":  "12",
// }
var months = map[string]string{
	"januar":    "01",
	"februar":   "02",
	"maerz":     "03",
	"april":     "04",
	"mai":       "05",
	"juni":      "06",
	"juli":      "07",
	"august":    "08",
	"september": "09",
	"oktober":   "10",
	"november":  "11",
	"dezember":  "12",
}

func (h *JournalSkill) ProcessRequest(requestEnv *alexa.RequestEnvelope) *alexa.ResponseEnvelope {
	log.Infow("Request", "request", requestEnv.Request, "SessionAttributes", requestEnv.Session.Attributes)

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
				OutputSpeech: plainText("Dein Tagebuch ist nun geöffnet. Möchtest Du einen neuen Eintrag erstellen oder vorhandene Einträge vorlesen?"),
			},
		}

	case "IntentRequest":
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
					journal := h.journalProvider.Get(requestEnv.Session.User.AccessToken)
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
		case "ReadExistingEntriesIntent":
			switch requestEnv.Request.DialogState {
			case "STARTED", "IN_PROGRESS":
				return pureDelegate(&intent)
			case "COMPLETED":
				date, e := date.AutoParse(intent.Slots["year"].Value + "/" + months[intent.Slots["month"].Value] + "/" + fmt.Sprintf("%02v", intent.Slots["day"].Value))
				// date, e := date.AutoParse(intent.Slots["date"].Value)
				if e != nil {
					log.Errorw("Could not convert string to date", "date", intent.Slots["date"].Value, e)
					return internalError()
				}

				journal := h.journalProvider.Get(requestEnv.Session.User.AccessToken)
				text, e := journal.GetEntry(date)
				if e != nil {
					log.Errorw("Could not get entry", "date", date, e)
					return internalError()
				}

				return &alexa.ResponseEnvelope{Version: "1.0",
					Response: &alexa.Response{
						OutputSpeech:     plainText("Hier ist der Eintrag vom " + intent.Slots["day"].Value + ". " + intent.Slots["month"].Value + " " + intent.Slots["year"].Value + ": " + text),
						ShouldSessionEnd: true,
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
