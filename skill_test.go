package journalskill_test

import (
	"fmt"
	"runtime/debug"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/petergtz/alexa-journal"
	"github.com/petergtz/alexa-journal/cmd/skill/factory"
	"github.com/petergtz/alexa-journal/drive"
	. "github.com/petergtz/alexa-journal/matchers"
	"github.com/petergtz/go-alexa"
	"github.com/petergtz/pegomock"
	. "github.com/petergtz/pegomock/ginkgo_compatible"
	"go.uber.org/zap"
)

//go:generate pegomock generate --use-experimental-model-gen --package journalskill_test JournalProvider
//go:generate pegomock generate --use-experimental-model-gen --package journalskill_test -m ErrorReporter

var _ = Describe("Skill processes request", func() {

	var (
		skill           *JournalSkill
		journalProvider *MockJournalProvider
		errorReporter   *MockErrorReporter
	)
	BeforeEach(func() {
		loggerConfig := zap.NewDevelopmentConfig()
		logger, e := loggerConfig.Build()
		Expect(e).NotTo(HaveOccurred())

		journalProvider = NewMockJournalProvider()
		errorReporter = NewMockErrorReporter()
		Whenever(func() { errorReporter.ReportPanic(AnyInterface(), AnyPtrToGoAlexaRequestEnvelope()) }).
			Then(func(params []pegomock.Param) pegomock.ReturnValues {
				e := params[0]
				logger.Sugar().Error(fmt.Sprintf("%v\n%s", e, debug.Stack()))
				return nil
			})
		skill = NewJournalSkill(journalProvider,
			&drive.DriveSheetErrorInterpreter{ErrorReporter: errorReporter},
			logger.Sugar(),
			errorReporter,
			factory.CreateI18nBundle(),
			&factory.EmptyConfigService{})
	})

	Context("Session missing from request envelope", func() {
		Context("locale missing", func() {
			It("reports a panic to the error reporter before telling the user there was an internal error in English (default locale)", func() {
				var stackTrace string

				Whenever(func() { errorReporter.ReportPanic(AnyInterface(), AnyPtrToGoAlexaRequestEnvelope()) }).
					Then(func([]pegomock.Param) pegomock.ReturnValues {
						stackTrace = string(debug.Stack())
						return nil
					})

				response := skill.ProcessRequest(&alexa.RequestEnvelope{Request: &alexa.Request{}})

				reportedPanic, _ := errorReporter.VerifyWasCalledOnce().ReportPanic(AnyInterface(), AnyPtrToGoAlexaRequestEnvelope()).GetCapturedArguments()
				Expect(stackTrace).To(ContainSubstring("alexa-journal/skill.go"))
				Expect(reportedPanic).NotTo(BeEmpty())

				Expect(response.Response.OutputSpeech.Text).To(ContainSubstring("internal error"))
			})
		})

		Context("locale present", func() {
			It("reports a panic to the error reporter before informing the user about an internal error", func() {
				respEnv := skill.ProcessRequest(&alexa.RequestEnvelope{
					Request: &alexa.Request{Locale: "de_DE"},
				})

				reportedPanic, _ := errorReporter.VerifyWasCalledOnce().ReportPanic(AnyInterface(), AnyPtrToGoAlexaRequestEnvelope()).GetCapturedArguments()
				Expect(reportedPanic).To(BeEquivalentTo("invalid memory address or nil pointer dereference"))
				Expect(respEnv.Response.OutputSpeech.Text).To(ContainSubstring("interner Fehler"))
			})
		})
	})

	Context("No user access token", func() {
		Context("de", func() {
			It("tells user to link accounts", func() {
				respEnv := skill.ProcessRequest(&alexa.RequestEnvelope{
					Request: &alexa.Request{Locale: "de_DE"},
					Session: &alexa.Session{
						User: struct {
							UserID      string "json:\"userId\""
							AccessToken string "json:\"accessToken\""
						}{ /* empty */ },
					},
				})
				Expect(respEnv.Response.OutputSpeech.Text).To(Equal("Bevor Du Dein Tagebuch öffnen kannst, verbinde bitte zuerst Alexa mit Deinem Google Account in der Alexa App."))
			})
		})
		Context("en", func() {
			It("tells user to link accounts", func() {
				respEnv := skill.ProcessRequest(&alexa.RequestEnvelope{
					Request: &alexa.Request{Locale: "en_US"},
					Session: &alexa.Session{
						User: struct {
							UserID      string "json:\"userId\""
							AccessToken string "json:\"accessToken\""
						}{ /* empty */ },
					},
				})
				Expect(respEnv.Response.OutputSpeech.Text).To(Equal("Before you can open your journal, please link Alexa with your Google account in your Alexa app."))
			})
		})
	})
})
