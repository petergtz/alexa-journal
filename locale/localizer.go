package locale

import (
	"fmt"

	"github.com/nicksnyder/go-i18n/v2/i18n"
	"github.com/petergtz/alexa-journal/util"
)

type Localizer struct {
	i18n.Localizer
	shouldBeSuccinct bool
}

func NewLocalizer(bundle *i18n.Bundle, lang string, shouldBeSuccinct bool) *Localizer {
	return &Localizer{
		Localizer:        *i18n.NewLocalizer(bundle, lang),
		shouldBeSuccinct: shouldBeSuccinct,
	}
}

func (l *Localizer) MustLocalize(lc *i18n.LocalizeConfig) string {
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
