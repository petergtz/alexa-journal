package journal

import (
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"

	"github.com/petergtz/alexa-journal/util"
	"github.com/rickb777/date"
)

type Journal struct {
	Data  TabularData
	Index Index
}

type TabularData interface {
	Rows() ([][]string, error)
	AppendRow(row []string) error
	Empty() (bool, error)
	DeleteRow(rowNum int) error
}
type Index interface {
	Add(id string, text string)
	Search(query string) []Rank
}

type Rank struct {
	Result     string
	Confidence float32
}

type Entry struct {
	Timestamp time.Time
	EntryDate date.Date
	EntryText string
}

func entryFromSlice(parts []string) Entry {
	timestamp, e := time.Parse(TimestampFormat, parts[0])
	util.PanicOnError(e)
	if parts[1] == "" {
		panic(errors.New("parts[1] must not be empty"))
	}
	return Entry{timestamp, date.MustAutoParse(parts[1]), parts[2]}
}

const TimestampFormat = "2006-01-02 15:04:05"

func (j *Journal) AddEntry(entryDate date.Date, text string) error {
	empty, e := j.Data.Empty()
	if e != nil {
		return errors.Wrap(e, "Could not add entry")
	}
	if empty {
		e := j.Data.AppendRow([]string{"timestamp", "date", "text"})
		if e != nil {
			return errors.Wrap(e, "Could not add entry")
		}
	}
	e = j.Data.AppendRow([]string{time.Now().Format(TimestampFormat), entryDate.String(), text})
	if e != nil {
		return errors.Wrap(e, "Could not add entry")
	}
	return nil
}

func (j *Journal) GetEntry(entryDate date.Date) (string, error) {
	var entriesFound []Entry
	rows, e := j.Data.Rows()
	if e != nil {
		return "", errors.Wrap(e, "Could not get entry")
	}
	for _, parts := range rows {
		if len(parts) != 3 {
			continue
		}
		if parts[1] == "" {
			continue
		}
		d, e := date.AutoParse(parts[1])
		if e != nil {
			continue
		}
		if d == entryDate {
			entriesFound = append(entriesFound, entryFromSlice(parts))
		}
	}
	sort.SliceStable(entriesFound, ByTimestamp(entriesFound))
	var texts []string
	for _, entry := range entriesFound {
		texts = append(texts, entry.EntryText)
	}
	return strings.Join(texts, ". "), nil
}

func ByTimestamp(entriesFound []Entry) func(i, j int) bool {
	return func(i int, j int) bool { return entriesFound[i].Timestamp.Before(entriesFound[j].Timestamp) }
}

func (j *Journal) DeleteEntry(entryDate date.Date) error {
	rows, e := j.Data.Rows()
	if e != nil {
		return errors.Wrap(e, "Could not get data rows")
	}
	for i, parts := range rows {
		if len(parts) != 3 {
			continue
		}
		if parts[1] == "" {
			continue
		}
		d, e := date.AutoParse(parts[1])
		if e != nil {
			continue
		}
		if d == entryDate {
			e := j.Data.DeleteRow(i)
			if e != nil {
				return errors.Wrapf(e, "Could not delete row %v in data", i)
			}
		}
	}
	return nil
}

func (j *Journal) GetClosestEntry(entryDate date.Date) (Entry, error) {
	var closestPositiveEntry, closestNegativeEntry *Entry

	closestPositiveDiff := -(1 << 30)
	closestNegativeDiff := 1 << 30
	rows, e := j.Data.Rows()
	if e != nil {
		return Entry{}, errors.Wrap(e, "Could not get closest entry")
	}
	for _, parts := range rows {
		if len(parts) != 3 {
			continue
		}
		if parts[1] == "" {
			continue
		}
		d, e := date.AutoParse(parts[1])
		if e != nil {
			continue
		}
		diff := entryDate.Sub(d)

		if diff == 0 {
			return entryFromSlice(parts), nil
		}
		if diff > 0 {
			if int(diff) < closestNegativeDiff {
				closestNegativeDiff = int(diff)
				entry := entryFromSlice(parts)
				closestNegativeEntry = &entry
			}
		}
		if diff < 0 {
			if int(diff) > closestPositiveDiff {
				closestPositiveDiff = int(diff)
				entry := entryFromSlice(parts)
				closestPositiveEntry = &entry
			}
		}
	}
	if closestPositiveEntry != nil {
		return *closestPositiveEntry, nil
	}
	if closestNegativeEntry != nil {
		return *closestNegativeEntry, nil
	}
	return Entry{}, nil
}

func (j *Journal) GetEntries(timeRange string) ([]Entry, error) {
	var result []Entry
	rows, e := j.Data.Rows()
	if e != nil {
		return nil, errors.Wrap(e, "Could not get entries")
	}
	for _, parts := range rows {
		if len(parts) != 3 {
			continue
		}
		if strings.HasPrefix(parts[1], timeRange) {
			result = append(result, entryFromSlice(parts))
		}
	}
	return result, nil
}

func (j *Journal) SearchFor(query string) ([]Entry, error) {
	lookup := make(map[string]string)
	rows, e := j.Data.Rows()
	if e != nil {
		return nil, errors.Wrap(e, "Could not get entries")
	}
	for _, parts := range rows {
		if len(parts) != 3 {
			continue
		}
		if parts[1] != "" {
			if _, e := date.AutoParse(parts[1]); e == nil {
				j.Index.Add(parts[1], parts[2])
				lookup[parts[1]] = parts[2]
			}
		}
	}
	hits := j.Index.Search(query)

	result := make([]Entry, len(hits))
	i := 0
	for _, hit := range hits {
		result[i] = Entry{
			EntryDate: date.MustAutoParse(hit.Result),
			EntryText: lookup[hit.Result],
		}
		i++
	}
	sort.Slice(result, ByEntryDate(result))
	return result, nil
}

func ByEntryDate(entries []Entry) func(i, j int) bool {
	return func(i int, j int) bool { return entries[i].EntryDate.Before(entries[j].EntryDate) }
}
