package locale

import (
	"fmt"
	"strings"
	"time"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/petergtz/alexa-journal/locale/resources"
	"github.com/petergtz/alexa-journal/util"
)

type Localizer struct {
	i18n.Localizer
	lang             string
	shouldBeSuccinct bool
}

func NewLocalizer(bundle *i18n.Bundle, lang string, shouldBeSuccinct bool) *Localizer {
	return &Localizer{
		Localizer:        *i18n.NewLocalizer(bundle, lang),
		lang:             lang,
		shouldBeSuccinct: shouldBeSuccinct,
	}
}
func (l *Localizer) Get(ids ...resources.StringID) string {
	var result []string
	for _, id := range ids {
		result = append(result, l.mustLocalize(&i18n.LocalizeConfig{
			DefaultMessage: &i18n.Message{ID: id.String()},
		}))
	}
	return strings.Join(result, "\n")
}

func (l *Localizer) GetTemplated(id resources.StringID, templateData interface{}) string {
	return l.mustLocalize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID: id.String(),
		},
		TemplateData: templateData,
	})
}

func (l *Localizer) Weekday(weekday time.Weekday) string {
	return resources.Weekdays[l.lang][weekday]
}

func (l *Localizer) mustLocalize(lc *i18n.LocalizeConfig) string {
	if l.shouldBeSuccinct {
		suffixed := *lc
		dm := *suffixed.DefaultMessage
		suffixed.DefaultMessage = &dm
		suffixed.DefaultMessage.ID = suffixed.DefaultMessage.ID + "_succinct"

		message, e := l.Localizer.Localize(&suffixed)

		if _, isNotFound := e.(*i18n.MessageNotFoundErr); isNotFound {
			fmt.Println("Succinct message not found")
			return l.Localizer.MustLocalize(lc)
		}
		util.PanicOnError(e)
		return message
	}
	return l.Localizer.MustLocalize(lc)
}
