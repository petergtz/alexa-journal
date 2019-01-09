package custom

import (
	"sort"
	"strings"
	"unicode"

	"go.uber.org/zap"

	"github.com/petergtz/alexa-journal/journal"
	"github.com/pkg/math"
)

type SearchIndex struct {
	Index map[string][]string
	Log   *zap.SugaredLogger
}

func NewSearchIndex(log *zap.SugaredLogger) *SearchIndex {
	return &SearchIndex{
		Index: make(map[string][]string),
		Log:   log,
	}
}

func (si *SearchIndex) Add(id string, text string) {
	for _, word := range wordsIn(text) {
		word = strings.ToLower(word)
		si.Index[word] = append(si.Index[word], id)
	}
}

func (si *SearchIndex) Search(query string) []journal.Rank {
	wordResults := make(map[string]map[string]float32)
	for _, word := range wordsIn(query) {
		word = strings.ToLower(word)
		wordResults[word] = make(map[string]float32)
		closestWords := closestMatches(word, keysAsSlice(si.Index), 0.75)

		si.Log.Debugw("Closest matches", "word", word, "closest-matches", closestWords)
		for _, closestWord := range closestWords {
			for _, id := range si.Index[strings.ToLower(closestWord.Result)] {
				wordResults[word][id] = closestWord.Confidence
			}
		}
	}
	si.Log.Debugw("word results", "wordResults", wordResults)

	result := make(map[string]float32)
	for _, wordResult := range wordResults {
		for id, confidence := range wordResult {
			result[id] += confidence / float32(len(wordResults))
		}
	}
	resultSlice := rankSliceFrom(result)
	sort.Slice(resultSlice, func(i int, j int) bool { return resultSlice[i].Confidence > resultSlice[j].Confidence })
	return topRanks(resultSlice, 0.75)
}

func wordsIn(text string) []string {
	return strings.FieldsFunc(text, func(c rune) bool { return !unicode.IsLetter(c) && !unicode.IsNumber(c) })
}

func closestMatches(word string, targets []string, cutOffConfidence float32) []journal.Rank {
	var ranks []journal.Rank
	for _, target := range targets {
		ranks = append(ranks, journal.Rank{target, 1.0 - float32(math.Min(len(word), LevenshteinDistance(word, target)))/float32(len(word))})
	}
	sort.Slice(ranks, func(i int, j int) bool { return ranks[i].Confidence > ranks[j].Confidence })
	return topRanks(ranks, cutOffConfidence)
}

func keysAsSlice(m map[string][]string) []string {
	words := make([]string, len(m))
	u := 0
	for word := range m {
		words[u] = strings.ToLower(word)
		u++
	}
	return words
}

func rankSliceFrom(m map[string]float32) []journal.Rank {
	resultSlice := make([]journal.Rank, len(m))
	u := 0
	for id, confidence := range m {
		resultSlice[u] = journal.Rank{id, confidence}
		u++
	}
	return resultSlice
}

func topRanks(ranks []journal.Rank, cutOffConfidence float32) []journal.Rank {
	i := 0
	for _, rank := range ranks {
		if rank.Confidence < cutOffConfidence {
			break
		}
		i++
	}
	return ranks[:i]
}

func LevenshteinDistance(s, t string) int {
	r1, r2 := []rune(s), []rune(t)
	column := make([]int, len(r1)+1)

	for y := 1; y <= len(r1); y++ {
		column[y] = y
	}

	for x := 1; x <= len(r2); x++ {
		column[0] = x

		for y, lastDiag := 1, x-1; y <= len(r1); y++ {
			oldDiag := column[y]
			cost := 0
			if r1[y-1] != r2[x-1] {
				cost = 1
			}
			column[y] = min(column[y]+1, column[y-1]+1, lastDiag+cost)
			lastDiag = oldDiag
		}
	}

	return column[len(r1)]
}

func min(a, b, c int) int {
	if a < b && a < c {
		return a
	} else if b < c {
		return b
	}
	return c
}
