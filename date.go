package main

import (
	"regexp"

	"github.com/petergtz/alexa-journal/util"
	"github.com/pkg/errors"
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
	monthDateRegex = regexp.MustCompile(`^\d{4}-\d{2}(-XX)?$`)
	yearDateRegex  = regexp.MustCompile(`^\d{4}(-XX-XX)?$`)
)

func DateFrom(dateString string, yearString string) (dayDate date.Date, monthDate string, dateType DateType) {
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
	entryDate, e := date.AutoParse(dateString)
	if yearString != "" {
		entryDate, e = date.AutoParse(yearString + dateString[4:])
	}
	util.PanicOnError(errors.Wrapf(e, "Could not convert dateString '%v' and yearString '%v' to date", dateString, yearString))
	return entryDate, "", DayDate
}
