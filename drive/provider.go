package drive

import (
	"time"

	"github.com/petergtz/alexa-journal/search/custom"

	"github.com/patrickmn/go-cache"
	j "github.com/petergtz/alexa-journal/journal"

	"go.uber.org/zap"
)

type DriveSheetJournalProvider struct {
	Log   *zap.SugaredLogger
	cache *cache.Cache
}

func NewDriveSheetJournalProvider(log *zap.SugaredLogger) *DriveSheetJournalProvider {
	return &DriveSheetJournalProvider{
		Log:   log,
		cache: cache.New(time.Hour, time.Hour),
	}
}

func (jp *DriveSheetJournalProvider) Get(accessToken string, spreadsheetName string) (j.Journal, error) {
	tabData, exists := jp.cache.Get(accessToken)
	if !exists {
		var e error
		tabData, e = NewSheetBasedTabularData(accessToken, spreadsheetName, spreadsheetName, jp.Log)
		if e != nil {
			return j.Journal{}, e
		}
	}

	jp.cache.SetDefault(accessToken, tabData)

	return j.Journal{
		Data:  tabData.(*SheetBasedTabularData),
		Index: custom.NewSearchIndex(jp.Log),
	}, nil
}
