package journalskill

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	. "github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/petergtz/alexa-journal/locale"
	r "github.com/petergtz/alexa-journal/locale/resources"

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

type ErrorReporter interface {
	ReportPanic(e interface{}, requestEnv *alexa.RequestEnvelope)
	ReportError(e error)
}
type JournalSkill struct {
	journalProvider  JournalProvider
	errorInterpreter ErrorInterpreter
	log              *zap.SugaredLogger
	errorReporter    ErrorReporter
	i18nBundle       *i18n.Bundle
	configService    ConfigService
}

type ConfigService interface {
	GetConfig(userID string) Config
	PersistConfig(userID string, config Config)
}

type Config struct {
	BeSuccinct                     bool
	ShouldExplainAboutSuccinctMode bool
}

func NewJournalSkill(journalProvider JournalProvider,
	errorInterpreter ErrorInterpreter,
	log *zap.SugaredLogger,
	errorReporter ErrorReporter,
	i18nBundle *i18n.Bundle,
	configService ConfigService,
) *JournalSkill {
	return &JournalSkill{
		journalProvider:  journalProvider,
		errorInterpreter: errorInterpreter,
		log:              log,
		errorReporter:    errorReporter,
		configService:    configService,
		i18nBundle:       i18nBundle,
	}
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

type SessionAttributes struct {
	Drafts   map[string][]string `json:"drafts"`
	Drafting bool                `json:"drafting"`
}

func (h *JournalSkill) ProcessRequest(requestEnv *alexa.RequestEnvelope) (responseEnv *alexa.ResponseEnvelope) {
	defer func() {
		if e := recover(); e != nil {
			h.errorReporter.ReportPanic(e, requestEnv)
			responseEnv = internalError(i18n.NewLocalizer(h.i18nBundle, requestEnv.Request.Locale))
		}
	}()

	log := h.log.With("request", requestEnv.Request, "session", requestEnv.Session)
	log.Infow("Request started")
	defer log.Infow("Request completed")

	if requestEnv.Session.User.AccessToken == "" {
		return &alexa.ResponseEnvelope{Version: "1.0",
			Response: &alexa.Response{
				OutputSpeech: plainText(i18n.NewLocalizer(h.i18nBundle, requestEnv.Request.Locale).
					MustLocalize(&LocalizeConfig{DefaultMessage: &Message{ID: r.LinkWithGoogleAccount.String()}})),
				Card:             &alexa.Card{Type: "LinkAccount"},
				ShouldSessionEnd: true,
			},
			SessionAttributes: requestEnv.Session.Attributes,
		}
	}
	config := h.configService.GetConfig(requestEnv.Session.User.UserID)
	l := locale.NewLocalizer(
		h.i18nBundle,
		requestEnv.Request.Locale,
		config.BeSuccinct)

	switch requestEnv.Request.Type {

	case "LaunchRequest":
		// cache warming:
		go h.journalProvider.Get(requestEnv.Session.User.AccessToken)

		return &alexa.ResponseEnvelope{Version: "1.0",
			Response:          &alexa.Response{OutputSpeech: plainText(l.Get(r.YourJournalIsNowOpen))},
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
		case "BeSuccinctIntent":
			newConfig := config
			newConfig.BeSuccinct = true
			h.configService.PersistConfig(requestEnv.Session.User.UserID, newConfig)
			return &alexa.ResponseEnvelope{Version: "1.0",
				Response: &alexa.Response{
					OutputSpeech: plainText(l.Get(r.OkayWillBeSuccinct, r.WhatDoYouWantToDoNext)),
					Reprompt:     &alexa.Reprompt{OutputSpeech: plainText(l.Get(r.WhatDoYouWantToDoNext))},
				},
				SessionAttributes: mapStringInterfaceFrom(sessionAttributes),
			}
		case "BeVerboseIntent":
			newConfig := config
			newConfig.BeSuccinct = false
			h.configService.PersistConfig(requestEnv.Session.User.UserID, newConfig)
			return &alexa.ResponseEnvelope{Version: "1.0",
				Response: &alexa.Response{
					OutputSpeech: plainText(l.Get(r.OkayWillBeVerbose, r.WhatDoYouWantToDoNext)),
					Reprompt:     &alexa.Reprompt{OutputSpeech: plainText(l.Get(r.WhatDoYouWantToDoNext))},
				},
				SessionAttributes: mapStringInterfaceFrom(sessionAttributes),
			}
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
					_, _, dateType := DateFrom(intent.Slots["date"].Value)
					if dateType != DayDate {
						return &alexa.ResponseEnvelope{Version: "1.0",
							Response: &alexa.Response{
								Directives: []interface{}{
									alexa.DialogDirective{
										Type:         "Dialog.ElicitSlot",
										SlotToElicit: "date",
									},
								},
								OutputSpeech: plainText(l.Get(r.InvalidDate)),
							},
							SessionAttributes: mapStringInterfaceFrom(sessionAttributes),
						}
					}
					// TODO: could we use intent.Slots["text"].Value == "" instead of !sessionAttributes.Drafting?
					if _, exists := sessionAttributes.Drafts[intent.Slots["date"].Value]; exists && !sessionAttributes.Drafting {
						switch intent.Slots["text"].ConfirmationStatus {
						case "NONE":
							outputSpeech := plainText(l.GetTemplated(r.NewEntryDraftExists, map[string]interface{}{
								"Draft": strings.Join(sessionAttributes.Drafts[intent.Slots["date"].Value], ". "),
							}))
							return &alexa.ResponseEnvelope{Version: "1.0",
								Response: &alexa.Response{
									OutputSpeech: outputSpeech,
									Directives:   []interface{}{alexa.DialogDirective{Type: "Dialog.ConfirmSlot", SlotToConfirm: "text"}},
									Reprompt:     &alexa.Reprompt{OutputSpeech: outputSpeech},
								},
								SessionAttributes: mapStringInterfaceFrom(sessionAttributes),
							}
						case "CONFIRMED":
							break
						case "DENIED":
							delete(sessionAttributes.Drafts, intent.Slots["date"].Value)
						}
					}
					switch strings.ToLower(intent.Slots["text"].Value) {
					case "":
						sessionAttributes.Drafting = true
						dateString := ""
						if intent.Slots["date"].Value != "" {
							dateString = l.GetTemplated(r.ForDate,
								map[string]interface{}{"Date": intent.Slots["date"].Value})
						}

						return &alexa.ResponseEnvelope{Version: "1.0",
							Response: &alexa.Response{

								OutputSpeech: plainText(l.GetTemplated(r.YouCanNowCreateYourEntry, map[string]interface{}{"ForDate": dateString})),
								Directives:   []interface{}{alexa.DialogDirective{Type: "Dialog.ElicitSlot", SlotToElicit: "text"}},
								Reprompt: &alexa.Reprompt{
									OutputSpeech: plainText(l.GetTemplated(r.YouCanNowCreateYourEntry_succinct, map[string]interface{}{"ForDate": dateString})),
								},
							},
							SessionAttributes: mapStringInterfaceFrom(sessionAttributes),
						}
					case "wiederhole", "wiederholen":
						if len(sessionAttributes.Drafts[intent.Slots["date"].Value]) == 0 {
							return &alexa.ResponseEnvelope{Version: "1.0",
								Response: &alexa.Response{
									OutputSpeech: plainText(l.Get(r.YourEntryIsEmptyNoRepeat)),
									Directives:   []interface{}{alexa.DialogDirective{Type: "Dialog.ElicitSlot", SlotToElicit: "text"}},
								},
								SessionAttributes: requestEnv.Session.Attributes,
							}
						}

						return &alexa.ResponseEnvelope{Version: "1.0",
							Response: &alexa.Response{
								OutputSpeech: plainText(l.GetTemplated(r.IRepeat, map[string]interface{}{
									"Text": sessionAttributes.Drafts[intent.Slots["date"].Value][len(sessionAttributes.Drafts[intent.Slots["date"].Value])-1]})),
								Directives: []interface{}{alexa.DialogDirective{Type: "Dialog.ElicitSlot", SlotToElicit: "text"}},
							},
							SessionAttributes: requestEnv.Session.Attributes,
						}
					case "korrigiere", "korrigieren":
						if len(sessionAttributes.Drafts[intent.Slots["date"].Value]) == 0 {
							return &alexa.ResponseEnvelope{Version: "1.0",
								Response: &alexa.Response{
									OutputSpeech: plainText(l.Get(r.YourEntryIsEmptyNoCorrect)),
									Directives:   []interface{}{alexa.DialogDirective{Type: "Dialog.ElicitSlot", SlotToElicit: "text"}},
								},
								SessionAttributes: requestEnv.Session.Attributes,
							}
						}
						sessionAttributes.Drafts[intent.Slots["date"].Value] = sessionAttributes.Drafts[intent.Slots["date"].Value][:len(sessionAttributes.Drafts[intent.Slots["date"].Value])-1]
						return &alexa.ResponseEnvelope{Version: "1.0",
							Response: &alexa.Response{
								OutputSpeech: plainText(l.Get(r.OkayCorrectPart)),
								Directives:   []interface{}{alexa.DialogDirective{Type: "Dialog.ElicitSlot", SlotToElicit: "text"}},
								Reprompt:     &alexa.Reprompt{OutputSpeech: plainText(l.Get(r.CorrectPartReprompt))},
							},
							SessionAttributes: mapStringInterfaceFrom(sessionAttributes),
						}
					case "abbrechen":
						sessionAttributes.Drafting = false
						return &alexa.ResponseEnvelope{Version: "1.0",
							Response: &alexa.Response{
								OutputSpeech: plainText(l.Get(r.NewEntryAborted, r.LongPause) +
									h.succinctModeExplanation(requestEnv.Session.User.UserID, config, l) +
									l.Get(r.LongPause, r.WhatDoYouWantToDoNext)),
							},
							SessionAttributes: mapStringInterfaceFrom(sessionAttributes),
						}
					case "fertig":
						if len(sessionAttributes.Drafts[intent.Slots["date"].Value]) == 0 {
							sessionAttributes.Drafting = false
							return &alexa.ResponseEnvelope{Version: "1.0",
								Response: &alexa.Response{OutputSpeech: plainText(l.Get(r.YourEntryIsEmptyNoSave, r.LongPause) +
									h.succinctModeExplanation(requestEnv.Session.User.UserID, config, l) +
									l.Get(r.LongPause, r.WhatDoYouWantToDoNext))},
								SessionAttributes: mapStringInterfaceFrom(sessionAttributes),
							}
						}
						return &alexa.ResponseEnvelope{Version: "1.0",
							Response: &alexa.Response{
								OutputSpeech: plainText(l.GetTemplated(r.NewEntryConfirmation, map[string]interface{}{
									"Date": intent.Slots["date"].Value,
									"Text": strings.Join(sessionAttributes.Drafts[intent.Slots["date"].Value], ". "),
								})),
								Directives: []interface{}{alexa.DialogDirective{Type: "Dialog.ConfirmIntent", UpdatedIntent: &intent}},
								Reprompt:   &alexa.Reprompt{OutputSpeech: plainText(l.Get(r.NewEntryConfirmationReprompt))},
							},
							SessionAttributes: requestEnv.Session.Attributes,
						}
					default:
						sessionAttributes.Drafts[intent.Slots["date"].Value] = append(sessionAttributes.Drafts[intent.Slots["date"].Value], intent.Slots["text"].Value)

						return &alexa.ResponseEnvelope{Version: "1.0",
							Response: &alexa.Response{
								OutputSpeech: plainText(l.GetTemplated(r.IRepeat, map[string]interface{}{"Text": intent.Slots["text"].Value})),
								Directives:   []interface{}{alexa.DialogDirective{Type: "Dialog.ElicitSlot", SlotToElicit: "text"}},
								Reprompt:     &alexa.Reprompt{OutputSpeech: plainText(l.Get(r.NextPartPleaseReprompt))},
							},
							SessionAttributes: mapStringInterfaceFrom(sessionAttributes),
						}
					}
				case "CONFIRMED":
					date, _, dateType := DateFrom(intent.Slots["date"].Value)
					if dateType != DayDate {
						panic(errors.Errorf("Could not parse string '%v' to day date", intent.Slots["date"].Value))
					}

					journal.AddEntry(date, strings.Join(sessionAttributes.Drafts[intent.Slots["date"].Value], ". "))

					sessionAttributes.Drafting = false
					delete(sessionAttributes.Drafts, intent.Slots["date"].Value)

					return &alexa.ResponseEnvelope{Version: "1.0",
						Response: &alexa.Response{OutputSpeech: plainText(l.Get(r.OkaySaved, r.LongPause) +
							h.succinctModeExplanation(requestEnv.Session.User.UserID, config, l) +
							l.Get(r.LongPause, r.WhatDoYouWantToDoNext))},
						SessionAttributes: mapStringInterfaceFrom(sessionAttributes),
					}

				case "DENIED":
					sessionAttributes.Drafting = false
					return &alexa.ResponseEnvelope{Version: "1.0",
						Response: &alexa.Response{OutputSpeech: plainText(l.Get(r.OkayNotSaved, r.LongPause) +
							h.succinctModeExplanation(requestEnv.Session.User.UserID, config, l) +
							l.Get(r.LongPause, r.WhatDoYouWantToDoNext))},
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
				_, monthDate, dateType := DateFrom(intent.Slots["date"].Value)
				if dateType == MonthDate {
					return listAllEntriesInDate(&journal, monthDate, requestEnv.Session.Attributes, h.errorInterpreter, l)
				}
				return &alexa.ResponseEnvelope{Version: "1.0",
					Response: &alexa.Response{
						OutputSpeech: plainText(l.Get(r.DidNotUnderstandTryAgain)),
						Reprompt:     &alexa.Reprompt{OutputSpeech: plainText(l.Get(r.DidNotUnderstandTryAgain))},
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
				entryDate, monthDate, dateType := DateFrom(intent.Slots["date"].Value)
				if dateType == Invalid {
					return &alexa.ResponseEnvelope{Version: "1.0",
						Response: &alexa.Response{
							OutputSpeech: plainText(fmt.Sprintf(l.Get(r.DidNotUnderstandTryAgain, r.ExampleDateQuery))),
							Reprompt:     &alexa.Reprompt{OutputSpeech: plainText(l.Get(r.DidNotUnderstandTryAgain))},
						},
						SessionAttributes: requestEnv.Session.Attributes,
					}
				}
				if dateType == MonthDate {
					return readAllEntriesInDate(&journal, monthDate, requestEnv.Session.Attributes, h.errorInterpreter, l)
				}

				text, e := journal.GetEntry(entryDate)
				if e != nil {
					return plainTextRespEnv(l.Get(r.CouldNotGetEntry, r.ShortPause)+h.errorInterpreter.Interpret(e),
						requestEnv.Session.Attributes)
				}
				if text != "" {
					return &alexa.ResponseEnvelope{Version: "1.0",
						Response: &alexa.Response{
							OutputSpeech: plainText(l.GetTemplated(r.ReadEntry, map[string]interface{}{
								"WeekDay": l.Weekday(entryDate.Weekday()),
								"Date":    entryDate.String(),
								"Text":    text,
							}) + l.Get(r.LongPause, r.WhatDoYouWantToDoNext)),
						},
						SessionAttributes: requestEnv.Session.Attributes,
					}
				}
				closestEntry, e := journal.GetClosestEntry(entryDate)
				if e != nil {
					return plainTextRespEnv(l.Get(r.CouldNotGetEntry, r.ShortPause)+h.errorInterpreter.Interpret(e), requestEnv.Session.Attributes)
				}
				if closestEntry == (j.Entry{}) {
					return &alexa.ResponseEnvelope{Version: "1.0",
						Response:          &alexa.Response{OutputSpeech: plainText(l.Get(r.JournalIsEmpty, r.LongPause, r.WhatDoYouWantToDoNext, r.ShortPause, r.NewEntryExample))},
						SessionAttributes: requestEnv.Session.Attributes,
					}
				}
				return &alexa.ResponseEnvelope{Version: "1.0",
					Response: &alexa.Response{
						OutputSpeech: plainText(l.GetTemplated(r.EntryForDateNotFound, map[string]interface{}{
							"SearchDate": entryDate.String(),
							"WeekDay":    l.Weekday(entryDate.Weekday()),
							"Date":       closestEntry.EntryDate.String(),
							"Text":       closestEntry.EntryText,
						}) + l.Get(r.LongPause, r.WhatDoYouWantToDoNext)),
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
						OutputSpeech: plainText(l.Get(r.DidNotUnderstandTryAgain, r.ShortPause, r.ExampleRelativeDateQuery)),
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
				return plainTextRespEnv(l.Get(r.CouldNotGetEntry, r.ShortPause)+h.errorInterpreter.Interpret(e),
					requestEnv.Session.Attributes)
			}
			if text != "" {
				return &alexa.ResponseEnvelope{Version: "1.0",
					Response: &alexa.Response{
						OutputSpeech: plainText(l.GetTemplated(r.ReadEntry, map[string]interface{}{
							"WeekDay": l.Weekday(entryDate.Weekday()),
							"Date":    entryDate.String(),
							"Text":    text,
						}) + l.Get(r.LongPause, r.WhatDoYouWantToDoNext)),
					},
					SessionAttributes: requestEnv.Session.Attributes,
				}
			}
			closestEntry, e := journal.GetClosestEntry(entryDate)
			if e != nil {
				return plainTextRespEnv(l.Get(r.CouldNotGetEntry, r.ShortPause)+h.errorInterpreter.Interpret(e),
					requestEnv.Session.Attributes)
			}
			if closestEntry == (j.Entry{}) {
				return &alexa.ResponseEnvelope{Version: "1.0",
					Response: &alexa.Response{OutputSpeech: plainText(l.Get(
						r.JournalIsEmpty, r.LongPause, r.WhatDoYouWantToDoNext, r.ShortPause, r.NewEntryExample))},
					SessionAttributes: requestEnv.Session.Attributes,
				}
			}
			return &alexa.ResponseEnvelope{Version: "1.0",
				Response: &alexa.Response{
					OutputSpeech: plainText(l.GetTemplated(r.EntryForDateNotFound, map[string]interface{}{
						"SearchDate": entryDate.String(),
						"WeekDay":    l.Weekday(closestEntry.EntryDate.Weekday()),
						"Date":       closestEntry.EntryDate.String(),
						"Text":       closestEntry.EntryText,
					}) + l.Get(r.LongPause, r.WhatDoYouWantToDoNext)),
				},
				SessionAttributes: requestEnv.Session.Attributes,
			}
		case "SearchIntent":
			entries, e := journal.SearchFor(intent.Slots["query"].Value)
			if e != nil {
				return plainTextRespEnv(l.Get(r.SearchError, r.ShortPause)+h.errorInterpreter.Interpret(e),
					requestEnv.Session.Attributes)
			}
			if len(entries) == 0 {
				return &alexa.ResponseEnvelope{Version: "1.0",
					Response: &alexa.Response{
						OutputSpeech: plainText(l.GetTemplated(r.SearchNoResultsFound,
							map[string]interface{}{"Query": intent.Slots["query"].Value}) +
							l.Get(r.LongPause, r.WhatDoYouWantToDoNext)),
					},
					SessionAttributes: requestEnv.Session.Attributes,
				}
			}
			text := l.GetTemplated(r.SearchResults, map[string]interface{}{"Query": intent.Slots["query"].Value})
			for _, entry := range entries {
				tuple := l.Weekday(entry.EntryDate.Weekday()) + ", " + entry.EntryDate.String() + ": " + strings.TrimRight(entry.EntryText, ". ") + ". "
				if len(text)+len(tuple)+len(l.Get(r.WhatDoYouWantToDoNext)) > responseTextLimit {
					break
				}
				text += tuple
			}
			return &alexa.ResponseEnvelope{Version: "1.0",
				Response:          &alexa.Response{OutputSpeech: plainText(strings.TrimSpace(text) + l.Get(r.LongPause, r.WhatDoYouWantToDoNext))},
				SessionAttributes: requestEnv.Session.Attributes,
			}
		case "DeleteEntryIntent":
			switch requestEnv.Request.DialogState {
			case "STARTED":
				return pureDelegate(&intent, requestEnv.Session.Attributes)
			case "IN_PROGRESS", "COMPLETED":
				switch intent.ConfirmationStatus {
				case "NONE":
					date, _, dateType := DateFrom(intent.Slots["date"].Value)
					if dateType != DayDate {
						intent.Slots["date"] = alexa.IntentSlot{
							Name:               intent.Slots["date"].Name,
							ConfirmationStatus: intent.Slots["date"].ConfirmationStatus,
							Resolutions:        intent.Slots["date"].Resolutions,
							Value:              "",
						}
						return pureDelegate(&intent, requestEnv.Session.Attributes)
					}
					util.PanicOnError(errors.Wrapf(e, "Could not convert string '%v' to date", intent.Slots["date"].Value))

					entry, e := journal.GetEntry(date)
					if e != nil {
						return plainTextRespEnv(l.Get(r.DeleteEntryCouldNotGetEntry, r.ShortPause)+h.errorInterpreter.Interpret(e),
							requestEnv.Session.Attributes)
					}
					if entry == "" {
						return plainTextRespEnv(l.Get(r.DeleteEntryNotFound), requestEnv.Session.Attributes)
					}
					return &alexa.ResponseEnvelope{Version: "1.0",
						Response: &alexa.Response{
							OutputSpeech: plainText(l.GetTemplated(r.DeleteEntryConfirmation, map[string]interface{}{"Entry": entry})),
							Directives:   []interface{}{alexa.DialogDirective{Type: "Dialog.ConfirmIntent", UpdatedIntent: &intent}},
							Reprompt:     &alexa.Reprompt{OutputSpeech: plainText(l.GetTemplated(r.DeleteEntryConfirmation, map[string]interface{}{"Entry": entry}))},
						},
						SessionAttributes: requestEnv.Session.Attributes,
					}
				case "CONFIRMED":
					date, e := date.AutoParse(intent.Slots["date"].Value)
					util.PanicOnError(errors.Wrapf(e, "Could not convert string '%v' to date", intent.Slots["date"].Value))

					e = journal.DeleteEntry(date)
					if e != nil {
						return plainTextRespEnv(l.Get(r.DeleteEntryError, r.ShortPause)+h.errorInterpreter.Interpret(e),
							requestEnv.Session.Attributes)
					}

					return plainTextRespEnv(l.Get(r.OkayDeleted, r.LongPause, r.WhatDoYouWantToDoNext), requestEnv.Session.Attributes)
				case "DENIED":
					return plainTextRespEnv(l.Get(r.OkayNotDeleted, r.LongPause, r.WhatDoYouWantToDoNext), requestEnv.Session.Attributes)
				default:
					panic(errors.New("Invalid intent.ConfirmationStatus"))
				}
			default:
				panic(errors.New("Invalid requestEnv.Request.DialogState"))
			}

		case "AMAZON.HelpIntent":
			return &alexa.ResponseEnvelope{Version: "1.0",
				Response:          &alexa.Response{OutputSpeech: plainText(l.Get(r.Help))},
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

func (h *JournalSkill) succinctModeExplanation(userID string, config Config, l *locale.Localizer) string {
	if config.ShouldExplainAboutSuccinctMode {
		config.ShouldExplainAboutSuccinctMode = false
		h.configService.PersistConfig(userID, config)
		return l.Get(r.SuccinctModeExplanation)
	}
	return ""
}

func listAllEntriesInDate(journal *j.Journal, dateSlotValue string, sessionAttributes map[string]interface{}, errorInterpreter ErrorInterpreter, l *locale.Localizer) *alexa.ResponseEnvelope {
	entries, e := journal.GetEntries(dateSlotValue[:7])
	if e != nil {
		return plainTextRespEnv(l.Get(r.CouldNotGetEntries, r.ShortPause)+errorInterpreter.Interpret(e),
			sessionAttributes)
	}
	if len(entries) == 0 {
		return &alexa.ResponseEnvelope{Version: "1.0",
			Response: &alexa.Response{
				OutputSpeech: plainText(l.GetTemplated(r.NoEntriesInTimeRangeFound, map[string]interface{}{
					"TimeRange": readableStringFrom(dateSlotValue)}) +
					l.Get(r.LongPause, r.WhatDoYouWantToDoNext)),
			},
			SessionAttributes: sessionAttributes,
		}
	}
	var tuples []string
	for _, entry := range entries {
		tuples = append(tuples, l.Weekday(entry.EntryDate.Weekday())+", "+entry.EntryDate.String()+": "+entry.EntryText)
	}
	// TODO: limit response length
	return &alexa.ResponseEnvelope{Version: "1.0",
		Response: &alexa.Response{
			OutputSpeech: plainText(l.GetTemplated(r.EntriesInTimeRange, map[string]interface{}{
				"Date": readableStringFrom(dateSlotValue), "Entries": strings.Join(tuples, ". "),
			}) + l.Get(r.LongPause, r.WhatDoYouWantToDoNext)),
		},
		SessionAttributes: sessionAttributes,
	}
}

func readAllEntriesInDate(journal *j.Journal, dateSlotValue string, sessionAttributes map[string]interface{}, errorInterpreter ErrorInterpreter, l *locale.Localizer) *alexa.ResponseEnvelope {
	entries, e := journal.GetEntries(dateSlotValue[:7])
	if e != nil {
		return plainTextRespEnv(l.Get(r.CouldNotGetEntries, r.ShortPause)+errorInterpreter.Interpret(e),
			sessionAttributes)
	}
	if len(entries) == 0 {
		return &alexa.ResponseEnvelope{Version: "1.0",
			Response: &alexa.Response{
				OutputSpeech: plainText(l.GetTemplated(r.NoEntriesInTimeRangeFound, map[string]interface{}{
					"TimeRange": readableStringFrom(dateSlotValue)}) +
					l.Get(r.LongPause, r.WhatDoYouWantToDoNext)),
			},
			SessionAttributes: sessionAttributes,
		}
	}
	var tuples []string
	for _, entry := range entries {
		tuples = append(tuples, l.Weekday(entry.EntryDate.Weekday())+", "+entry.EntryDate.String()+": "+entry.EntryText)
	}
	return &alexa.ResponseEnvelope{Version: "1.0",
		Response: &alexa.Response{
			OutputSpeech: plainText(l.GetTemplated(r.EntriesInTimeRange, map[string]interface{}{
				"Date": readableStringFrom(dateSlotValue), "Entries": strings.Join(tuples, ". "),
			}) + l.Get(r.LongPause, r.WhatDoYouWantToDoNext)),
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

func internalError(l *i18n.Localizer) *alexa.ResponseEnvelope {
	return &alexa.ResponseEnvelope{Version: "1.0", Response: &alexa.Response{
		OutputSpeech:     plainText(l.MustLocalize(&LocalizeConfig{MessageID: r.InternalError.String()})),
		ShouldSessionEnd: true,
	}}
}

func mapStringInterfaceFrom(sessionAttributes SessionAttributes) map[string]interface{} {
	sessionAttributesBuf, e := json.Marshal(sessionAttributes)
	util.PanicOnError(errors.Wrap(e, "Could not marshal sessionAttributes"))

	var sessionAttributesMap map[string]interface{}
	e = json.Unmarshal(sessionAttributesBuf, &sessionAttributesMap)
	util.PanicOnError(errors.Wrap(e, "Could not unmarshal sessionAttributesBuf"))

	return sessionAttributesMap
}
