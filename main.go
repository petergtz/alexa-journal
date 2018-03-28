package main

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

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
}

const helpText = "Um einen Artikel vorgelesen zu bekommen, " +
	"sage z.B. \"Suche nach Käsekuchen.\" oder \"Was ist Käsekuchen?\". " +
	"Du kannst jederzeit zum Inhaltsverzeichnis springen, indem Du \"Inhaltsverzeichnis\" sagst. " +
	"Oder sage \"Springe zu Abschnitt 3.2\", um direkt zu diesem Abschnitt zu springen."

const quickHelpText = "Suche zunächst nach einem Begriff. " +
	"Sage z.B. \"Suche nach Käsekuchen.\" oder \"Was ist Käsekuchen?\"."

func (h *JournalSkill) ProcessRequest(requestEnv *alexa.RequestEnvelope) *alexa.ResponseEnvelope {
	log.Infow("Request", "request", requestEnv.Request, "SessionAttributes", requestEnv.Session.Attributes)
	switch requestEnv.Request.Type {

	case "LaunchRequest":
		return &alexa.ResponseEnvelope{Version: "1.0",
			Response: &alexa.Response{
				OutputSpeech: plainText("Du befindest Dich jetzt in Deinem Tagebuch. Moechtest Du einen neuen Eintrag erstellen oder vorhandene Eintraege vorlesen?"),
			},
		}

	case "IntentRequest":
		intent := requestEnv.Request.Intent
		switch intent.Name {
		case "NewEntryIntent":

			switch requestEnv.Request.DialogState {
			case "STARTED":
				// intent.Slots["date"] = IntentSlot{
				// 	Name:  "date",
				// 	Value: time.Now().Format("11.01.2018"),
				// }
				return &alexa.ResponseEnvelope{Version: "1.0",
					Response: &alexa.Response{
						Directives: []interface{}{
							alexa.DialogDirective{
								Type:          "Dialog.Delegate",
								UpdatedIntent: &intent,
							},
						},
					},
				}
			case "IN_PROGRESS":
				// if intent.Slots["correct"].Value == "" {
				return &alexa.ResponseEnvelope{Version: "1.0",
					Response: &alexa.Response{
						Directives: []interface{}{
							alexa.DialogDirective{
								Type:          "Dialog.Delegate",
								UpdatedIntent: &intent,
							},
						},
					},
				}
				// }
				// switch intent.Slots["correct"].Value {
				// case "datum":
				// 	delete(intent.Slots, "date")
				// 	delete(intent.Slots, "correct")
				// 	intent.ConfirmationStatus = "NONE"
				// 	return &alexa.ResponseEnvelope{Version: "1.0",
				// 		Response: &Response{
				// 			Directives: []interface{}{
				// 				DialogDirective{
				// 					Type:          "Dialog.Delegate",
				// 					UpdatedIntent: &intent,
				// 				},
				// 			},
				// 		},
				// 	}
				// case "text":
				// 	delete(intent.Slots, "text")
				// 	delete(intent.Slots, "correct")
				// 	intent.ConfirmationStatus = "NONE"
				// 	return &alexa.ResponseEnvelope{Version: "1.0",
				// 		Response: &Response{
				// 			Directives: []interface{}{
				// 				DialogDirective{
				// 					Type:          "Dialog.Delegate",
				// 					UpdatedIntent: &intent,
				// 				},
				// 			},
				// 		},
				// 	}
				// default:
				// 	panic("TODO")
				// }
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
					dateParts := strings.Split(intent.Slots["date"].Value, "-")
					year, _ := strconv.Atoi(dateParts[0])
					month, _ := strconv.Atoi(dateParts[1])
					day, _ := strconv.Atoi(dateParts[2])
					date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.Local)
					// date, e := time.Parse("2018-03-24", intent.Slots["date"].Value)
					// if e != nil {
					// 	log.Errorw("Could not parse date", "date", intent.Slots["date"].Value, "error", e)
					// 	return &alexa.ResponseEnvelope{Version: "1.0",
					// 		Response: &alexa.Response{
					// 			OutputSpeech:     plainText("Es ist ein interner Fehler aufgetreten."),
					// 			ShouldSessionEnd: true,
					// 		},
					// 	}

					// }
					journal, e := h.journalProvider.Get(requestEnv.Session.User.AccessToken)
					if e != nil {
						log.Errorw("Could not get Journal", "token", requestEnv.Session.User.AccessToken, "error", e)
						return &alexa.ResponseEnvelope{Version: "1.0",
							Response: &alexa.Response{
								OutputSpeech:     plainText("Es ist ein interner Fehler aufgetreten."),
								ShouldSessionEnd: true,
							},
						}
					}
					e = journal.AddEntry(date, intent.Slots["text"].Value)
					if e != nil {
						log.Errorw("Could not add entry", "date", date, "text", intent.Slots["text"].Value, "error", e)
						return &alexa.ResponseEnvelope{Version: "1.0",
							Response: &alexa.Response{
								OutputSpeech:     plainText("Es ist ein interner Fehler aufgetreten."),
								ShouldSessionEnd: true,
							},
						}
					}

					return &alexa.ResponseEnvelope{Version: "1.0",
						Response: &alexa.Response{OutputSpeech: plainText("Okay. Gespeichert.")},
					}

				case "DENIED":
					return &alexa.ResponseEnvelope{Version: "1.0",
						Response: &alexa.Response{OutputSpeech: plainText("Okay. Wurde verworfen.")},
					}
					// intent.ConfirmationStatus = "NONE"
					// return &alexa.ResponseEnvelope{Version: "1.0",
					// 	Response: &Response{
					// 		OutputSpeech: plainText("Okay. Was möchtest Du ändern?"),
					// 		Directives: []interface{}{
					// 			DialogDirective{
					// 				Type:          "Dialog.ElicitSlot",
					// 				SlotToElicit:  "correct",
					// 				UpdatedIntent: &intent,
					// 			},
					// 		},
					// 	},
					// }
				default:
					return internalError()
				}
			default:
				return internalError()
			}
		case "AMAZON.HelpIntent":
			return &alexa.ResponseEnvelope{Version: "1.0",
				Response: &alexa.Response{
					OutputSpeech: plainText(helpText),
				},
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

func quickHelp(sessionAttributes map[string]interface{}) *alexa.ResponseEnvelope {
	return &alexa.ResponseEnvelope{Version: "1.0",
		Response:          &alexa.Response{OutputSpeech: plainText(quickHelpText)},
		SessionAttributes: sessionAttributes,
	}
}

func plainText(text string) *alexa.OutputSpeech {
	return &alexa.OutputSpeech{Type: "PlainText", Text: text}
}

func internalError() *alexa.ResponseEnvelope {
	return &alexa.ResponseEnvelope{Version: "1.0",
		Response: &alexa.Response{
			OutputSpeech:     plainText("Es ist ein interner Fehler aufgetreten bei der Benutzung von Wikipedia."),
			ShouldSessionEnd: false,
		},
	}
}
