package journal

import (
	"sort"
	"strings"
	"time"

	"github.com/rickb777/date"
)

type Journal struct {
	Content string
}

func (j *Journal) AddEntry(entryDate date.Date, text string) {
	if j.Content == "" {
		j.Content = "timestamp\tdate\ttext\n"
	}
	j.Content += time.Now().Format("2006-01-02 15:04:05 -0700 MST") + "\t" + entryDate.String() + "\t" + text + "\n"
}

func (j *Journal) GetEntry(entryDate date.Date) string {
	var entriesFound []Entry
	for _, line := range strings.Split((j.Content), "\n") {
		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			continue
		}
		if parts[1] == entryDate.String() {
			entriesFound = append(entriesFound, Entry{parts[0], date.MustAutoParse(parts[1]), parts[2]})
		}
	}
	sort.SliceStable(entriesFound, func(i int, j int) bool {
		iTime, e := time.Parse("2006-01-02 15:04:05 -0700 MST", entriesFound[i].Timestamp)
		if e != nil {
			panic(e)
		}
		jTime, e := time.Parse("2006-01-02 15:04:05 -0700 MST", entriesFound[j].Timestamp)
		if e != nil {
			panic(e)
		}
		return iTime.Before(jTime)
	})
	var texts []string
	for _, entry := range entriesFound {
		texts = append(texts, entry.EntryText)
	}
	return strings.Join(texts, ". ")
}

func (j *Journal) GetClosestEntry(entryDate date.Date) Entry {
	var closestPositiveEntry, closestNegativeEntry *Entry

	closestPositiveDiff := -(1 << 30)
	closestNegativeDiff := 1 << 30
	for _, line := range strings.Split((j.Content), "\n") {
		parts := strings.Split(line, "\t")
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
			return Entry{parts[0], date.MustAutoParse(parts[1]), parts[2]}
		}
		if diff > 0 {
			if int(diff) < closestNegativeDiff {
				closestNegativeDiff = int(diff)
				closestNegativeEntry = &Entry{parts[0], date.MustAutoParse(parts[1]), parts[2]}
			}
		}
		if diff < 0 {
			if int(diff) > closestPositiveDiff {
				closestPositiveDiff = int(diff)
				closestPositiveEntry = &Entry{parts[0], date.MustAutoParse(parts[1]), parts[2]}
			}
		}
	}
	if closestPositiveEntry != nil {
		return *closestPositiveEntry
	}
	if closestNegativeEntry != nil {
		return *closestNegativeEntry
	}
	return Entry{}
}

type Entry struct {
	Timestamp string
	EntryDate date.Date
	EntryText string
}

func (j *Journal) GetEntries(timeRange string) []Entry {
	var result []Entry
	for _, line := range strings.Split((j.Content), "\n") {
		parts := strings.Split(line, "\t")
		if len(parts) != 3 {
			continue
		}
		if strings.HasPrefix(parts[1], timeRange) {
			result = append(result, Entry{parts[0], date.MustAutoParse(parts[1]), parts[2]})
		}
	}
	return result
}
