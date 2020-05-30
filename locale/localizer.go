package locale

import (
	"github.com/nicksnyder/go-i18n/v2/i18n"
)

type Localizer struct {
	i18n.Localizer
	shouldBeSuccinct bool
}

func NewLocalizer(bundle *i18n.Bundle, lang string, shouldBeSuccinct bool) *Localizer {
	return &Localizer{
		Localizer: *i18n.NewLocalizer(bundle, lang),
	}
}

func (l *Localizer) MustLocalize(lc *i18n.LocalizeConfig) string {
	suffixed := *lc
	if l.shouldBeSuccinct {
		suffixed.MessageID = suffixed.MessageID + "|succinct"
	}
	return l.Localizer.MustLocalize(&suffixed)
}
