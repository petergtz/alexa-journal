package main

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"

	"go.uber.org/zap"

	wiki "github.com/petergtz/alexa-wikipedia"
	"github.com/petergtz/alexa-wikipedia/mediawiki"
)

var (
	log                   *zap.SugaredLogger
	expectedApplicationID = os.Getenv("APPLICATION_ID")
)

func main() {
	l, e := zap.NewDevelopment()
	if e != nil {
		panic(e)
	}
	defer l.Sync()
	log = l.Sugar()

	handler := &Handler{wiki: &mediawiki.MediaWiki{}}
	http.HandleFunc("/", handler.handle)
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

type Handler struct {
	wiki wiki.Wiki
}

func (h *Handler) handle(w http.ResponseWriter, req *http.Request) {
	requestBody, e := ioutil.ReadAll(req.Body)
	if e != nil {
		log.Errorw("Error while reading request body", "error", e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var alexaRequest RequestEnvelope
	e = json.Unmarshal(requestBody, &alexaRequest)
	if e != nil {
		log.Errorw("Error while unmarshaling request body", "error", e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if alexaRequest.Session == nil {
		log.Errorw("Session is empty", "error", e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	response := h.processRequest(&alexaRequest)
	log.Infow("Response", "response", response)
	output, e := json.Marshal(response)
	if e != nil {
		log.Errorw("Error while marshalling response", "error", e)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.Write(output)
}

const helpText = "Um einen Artikel vorgelesen zu bekommen, " +
	"sage z.B. \"Suche nach Käsekuchen.\" oder \"Was ist Käsekuchen?\". " +
	"Du kannst jederzeit zum Inhaltsverzeichnis springen, indem Du \"Inhaltsverzeichnis\" sagst. " +
	"Oder sage \"Springe zu Abschnitt 3.2\", um direkt zu diesem Abschnitt zu springen."

const quickHelpText = "Suche zunächst nach einem Begriff. " +
	"Sage z.B. \"Suche nach Käsekuchen.\" oder \"Was ist Käsekuchen?\"."

func (h *Handler) processRequest(requestEnv *RequestEnvelope) *ResponseEnvelope {
	if requestEnv.Session.Application.ApplicationID != expectedApplicationID {
		log.Fatalf("ApplicationID does not match: %v", requestEnv.Session.Application.ApplicationID)
		return internalError()
	}

	log.Infow("Request", "request", requestEnv.Request, "SessionAttributes", requestEnv.Session.Attributes)
	switch requestEnv.Request.Type {

	case "LaunchRequest":
		return &ResponseEnvelope{Version: "1.0",
			Response: &Response{
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
				return &ResponseEnvelope{Version: "1.0",
					Response: &Response{
						Directives: []interface{}{
							DialogDirective{
								Type:          "Dialog.Delegate",
								UpdatedIntent: &intent,
							},
						},
					},
				}
			case "IN_PROGRESS":
				// if intent.Slots["correct"].Value == "" {
				return &ResponseEnvelope{Version: "1.0",
					Response: &Response{
						Directives: []interface{}{
							DialogDirective{
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
				// 	return &ResponseEnvelope{Version: "1.0",
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
				// 	return &ResponseEnvelope{Version: "1.0",
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
					return &ResponseEnvelope{Version: "1.0",
						Response: &Response{
							OutputSpeech: plainText("Alles klar. Ich habe folgenden Eintrag für das Datum " + intent.Slots["date"].Value + ": " +
								"\"" + intent.Slots["text"].Value + "\". Soll ich ihn so speichern?"),
							Directives: []interface{}{
								DialogDirective{
									Type:          "Dialog.ConfirmIntent",
									UpdatedIntent: &intent,
								},
							},
						},
					}
				case "CONFIRMED":
					return &ResponseEnvelope{Version: "1.0",
						Response: &Response{OutputSpeech: plainText("Okay. Gespeichert.")},
					}

				case "DENIED":
					return &ResponseEnvelope{Version: "1.0",
						Response: &Response{OutputSpeech: plainText("Okay. Wurde verworfen.")},
					}
					// intent.ConfirmationStatus = "NONE"
					// return &ResponseEnvelope{Version: "1.0",
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
			return &ResponseEnvelope{Version: "1.0",
				Response: &Response{
					OutputSpeech: plainText(helpText),
				},
			}
		case "AMAZON.CancelIntent", "AMAZON.StopIntent":
			return &ResponseEnvelope{Version: "1.0",
				Response: &Response{ShouldSessionEnd: true},
			}
		default:
			return internalError()
		}

	case "SessionEndedRequest":
		return &ResponseEnvelope{Version: "1.0"}

	default:
		return internalError()
	}
}

func lastQuestionIn(session *Session) string {
	if session.Attributes["last_question"] == nil {
		return ""
	}
	return session.Attributes["last_question"].(string)
}

func (h *Handler) pageFromSession(session *Session) (wiki.Page, *ResponseEnvelope) {
	if !wordIn(session) {
		return wiki.Page{}, quickHelp(session.Attributes)
	}

	page, e := h.wiki.GetPage(session.Attributes["word"].(string))
	if e != nil {
		log.Errorw("Could not get Wikipedia page", "error", e)
		return wiki.Page{}, internalError()
	}
	return page, nil
}

func quickHelp(sessionAttributes map[string]interface{}) *ResponseEnvelope {
	return &ResponseEnvelope{Version: "1.0",
		Response:          &Response{OutputSpeech: plainText(quickHelpText)},
		SessionAttributes: sessionAttributes,
	}
}

func wordIn(session *Session) bool {
	return session.Attributes["word"] != nil
}

type RequestEnvelope struct {
	Version string   `json:"version"`
	Session *Session `json:"session"`
	Request *Request `json:"request"`
	// TODO Add Request Context
}

// Session containes the session data from the Alexa request.
type Session struct {
	New        bool                   `json:"new"`
	SessionID  string                 `json:"sessionId"`
	Attributes map[string]interface{} `json:"attributes"`
	User       struct {
		UserID      string `json:"userId"`
		AccessToken string `json:"accessToken"`
	} `json:"user"`
	Application struct {
		ApplicationID string `json:"applicationId"`
	} `json:"application"`
}

// Request contines the data in the request within the main request.
type Request struct {
	Locale      string `json:"locale"`
	Timestamp   string `json:"timestamp"`
	Type        string `json:"type"`
	RequestID   string `json:"requestId"`
	DialogState string `json:"dialogState"`
	Intent      Intent `json:"intent"`
	Name        string `json:"name"`
}

// Intent contains the data about the Alexa Intent requested.
type Intent struct {
	Name               string                `json:"name"`
	ConfirmationStatus string                `json:"confirmationStatus,omitempty"`
	Slots              map[string]IntentSlot `json:"slots"`
}

// IntentSlot contains the data for one Slot
type IntentSlot struct {
	Name               string `json:"name"`
	ConfirmationStatus string `json:"confirmationStatus,omitempty"`
	Value              string `json:"value"`
}

// ResponseEnvelope contains the Response and additional attributes.
type ResponseEnvelope struct {
	Version           string                 `json:"version"`
	SessionAttributes map[string]interface{} `json:"sessionAttributes,omitempty"`
	Response          *Response              `json:"response"`
}

// Response contains the body of the response.
type Response struct {
	OutputSpeech     *OutputSpeech `json:"outputSpeech,omitempty"`
	Card             *Card         `json:"card,omitempty"`
	Reprompt         *Reprompt     `json:"reprompt,omitempty"`
	Directives       []interface{} `json:"directives,omitempty"`
	ShouldSessionEnd bool          `json:"shouldEndSession"`
}

// OutputSpeech contains the data the defines what Alexa should say to the user.
type OutputSpeech struct {
	Type string `json:"type"`
	Text string `json:"text,omitempty"`
	SSML string `json:"ssml,omitempty"`
}

// Card contains the data displayed to the user by the Alexa app.
type Card struct {
	Type    string `json:"type"`
	Title   string `json:"title,omitempty"`
	Content string `json:"content,omitempty"`
	Text    string `json:"text,omitempty"`
	Image   *Image `json:"image,omitempty"`
}

// Image provides URL(s) to the image to display in resposne to the request.
type Image struct {
	SmallImageURL string `json:"smallImageUrl,omitempty"`
	LargeImageURL string `json:"largeImageUrl,omitempty"`
}

// Reprompt contains data about whether Alexa should prompt the user for more data.
type Reprompt struct {
	OutputSpeech *OutputSpeech `json:"outputSpeech,omitempty"`
}

// AudioPlayerDirective contains device level instructions on how to handle the response.
type AudioPlayerDirective struct {
	Type         string     `json:"type"`
	PlayBehavior string     `json:"playBehavior,omitempty"`
	AudioItem    *AudioItem `json:"audioItem,omitempty"`
}

// AudioItem contains an audio Stream definition for playback.
type AudioItem struct {
	Stream Stream `json:"stream,omitempty"`
}

// Stream contains instructions on playing an audio stream.
type Stream struct {
	Token                string `json:"token"`
	URL                  string `json:"url"`
	OffsetInMilliseconds int    `json:"offsetInMilliseconds"`
}

// DialogDirective contains directives for use in Dialog prompts.
type DialogDirective struct {
	Type          string  `json:"type"`
	SlotToElicit  string  `json:"slotToElicit,omitempty"`
	SlotToConfirm string  `json:"slotToConfirm,omitempty"`
	UpdatedIntent *Intent `json:"updatedIntent,omitempty"`
}

func plainText(text string) *OutputSpeech {
	return &OutputSpeech{Type: "PlainText", Text: text}
}

func internalError() *ResponseEnvelope {
	return &ResponseEnvelope{Version: "1.0",
		Response: &Response{
			OutputSpeech:     plainText("Es ist ein interner Fehler aufgetreten bei der Benutzung von Wikipedia."),
			ShouldSessionEnd: false,
		},
	}
}
