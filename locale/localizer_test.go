package locale_test

import (
	"github.com/BurntSushi/toml"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	. "github.com/nicksnyder/go-i18n/v2/i18n"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/petergtz/alexa-journal/locale"
	"golang.org/x/text/language"
)

var _ = Describe("Localizer", func() {

	var i18nBundle *i18n.Bundle
	var l *locale.Localizer

	BeforeEach(func() {
		i18nBundle = i18n.NewBundle(language.English)
		i18nBundle.RegisterUnmarshalFunc("toml", toml.Unmarshal)
		i18nBundle.MustParseMessageFileBytes(locale.DeDe, "active.de.toml")
	})

	Context("Should be succinct", func() {
		BeforeEach(func() { l = locale.NewLocalizer(i18nBundle, "de-DE", true) })

		Context("Succinct message exists", func() {
			It("gives succinct message", func() {
				m := l.MustLocalize(&LocalizeConfig{
					DefaultMessage: &Message{ID: "YouCanNowCreateYourEntry"},
					TemplateData:   map[string]interface{}{"ForDate": "für den X"},
				})
				Expect(m).To(Equal("Du kannst Deinen Eintrag für den X nun verfassen. Los geht's!"))
			})
		})

		Context("Succinct message does not exist", func() {
			It("gives the verbose message", func() {
				l := locale.NewLocalizer(i18nBundle, "de-DE", true)
				m := l.MustLocalize(&LocalizeConfig{DefaultMessage: &Message{ID: "YourJournalIsNowOpen"}})
				Expect(m).To(Equal("Dein Tagebuch ist nun geöffnet. Was möchtest Du tun?"))
			})
		})
	})

	Context("Should be verbose", func() {
		BeforeEach(func() { l = locale.NewLocalizer(i18nBundle, "de-DE", false) })

		Context("Succinct message exists", func() {
			It("gives verbose message", func() {
				m := l.MustLocalize(&LocalizeConfig{
					DefaultMessage: &Message{ID: "YouCanNowCreateYourEntry"},
					TemplateData:   map[string]interface{}{"ForDate": "für den X"},
				})
				Expect(m).To(Equal("Du kannst Deinen Eintrag für den X nun verfassen; ich werde jeden Teil kurz bestaetigen, sodass du die moeglichkeit hast ihn zu \"korrigieren\" oder \"anzuhoeren\". Sage \"fertig\", wenn Du fertig bist."))
			})
		})

		Context("Succinct message does not exist", func() {
			It("gives the verbose message", func() {
				m := l.MustLocalize(&LocalizeConfig{DefaultMessage: &Message{ID: "YourJournalIsNowOpen"}})
				Expect(m).To(Equal("Dein Tagebuch ist nun geöffnet. Was möchtest Du tun?"))
			})
		})
	})
})