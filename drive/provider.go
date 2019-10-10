package drive

import (
	"github.com/petergtz/alexa-journal/search/custom"

	j "github.com/petergtz/alexa-journal/journal"

	"go.uber.org/zap"
)

type DriveSheetJournalProvider struct{ Log *zap.SugaredLogger }

func (jp *DriveSheetJournalProvider) Get(accessToken string) (j.Journal, error) {
	tabData, e := NewSheetBasedTabularData(accessToken, "Tagebuch", "Tagebuch", jp.Log)
	if e != nil {
		return j.Journal{}, e
	}
	return j.Journal{
		Data:  tabData,
		Index: custom.NewSearchIndex(jp.Log),
	}, nil
}
