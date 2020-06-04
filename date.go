package journalskill

import (
	"fmt"
	"regexp"

	"github.com/rickb777/date"
)

type DateType int

const (
	DayDate = iota
	MonthDate
	YearDate
	Invalid
)

var (
	monthDateRegex     = regexp.MustCompile(`^\d{4}-\d{2}(-XX)?$`)
	yearDateRegex      = regexp.MustCompile(`^\d{4}(-XX-XX)?$`)
	xxDayDateRegex     = regexp.MustCompile(`^XX\d{2}-\d{2}-\d{2}$`)
	xxxxXXDayDateRegex = regexp.MustCompile(`^XXXX-XX-\d{2}$`)
)

func DateFrom(dateString string, yearString string) (dayDate date.Date, monthDate string, dateType DateType) {
	if yearString != "" {
		yearString = fmt.Sprintf("%04s", yearString)
	}
	if dateString == "" {
		if yearString != "" {
			return date.Date{}, "", YearDate
		}
		return date.Date{}, "", Invalid
	}
	if monthDateRegex.MatchString(dateString) {
		if yearString != "" {
			return date.Date{}, yearString + dateString[4:7], MonthDate
		}
		return date.Date{}, dateString[:7], MonthDate
	}
	if yearDateRegex.MatchString(dateString) {
		return date.Date{}, "", YearDate
	}
	if yearString == "?" {
		return date.Date{}, "", Invalid
	}
	if xxxxXXDayDateRegex.MatchString(dateString) {
		today := date.Today()
		dateString = fmt.Sprintf("%04d-%02d-%v", today.Year(), today.Month(), dateString[8:])
	}
	if xxDayDateRegex.MatchString(dateString) {
		dateString = "20" + dateString[2:]
	}
	entryDate, e := date.AutoParse(dateString)
	if e != nil {
		return date.Date{}, "", Invalid
	}
	if yearString != "" {
		entryDate, e = date.AutoParse(yearString + dateString[4:])
		if e != nil {
			return date.Date{}, "", Invalid
		}
	}
	return entryDate, "", DayDate
}
