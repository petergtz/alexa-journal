package journal_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	j "github.com/petergtz/alexa-journal/journal"
	"github.com/rickb777/date"
)

var _ = Describe("Journal", func() {
	var journal j.Journal

	BeforeEach(func() {
		journal = j.Journal{}
	})

	Describe("GetEntry", func() {
		It("can find entry", func() {
			journal.AddEntry(date.MustAutoParse("1994-08-20"), "Example text")

			Expect(journal.GetEntry(date.MustAutoParse("1994-08-20"))).To(Equal("Example text"))
		})

		It("concats multiple entries with same date", func() {
			journal.AddEntry(date.MustAutoParse("1994-08-20"), "one")
			journal.AddEntry(date.MustAutoParse("1994-08-20"), "two")
			journal.AddEntry(date.MustAutoParse("1994-08-20"), "three")

			Expect(journal.GetEntry(date.MustAutoParse("1994-08-20"))).To(Equal("one. two. three"))
		})

	})

	Describe("GetClosestEntry", func() {
		It("can find entry", func() {
			journal.AddEntry(date.MustAutoParse("1994-08-04"), "One")
			journal.AddEntry(date.MustAutoParse("1994-08-20"), "Two")
			journal.AddEntry(date.MustAutoParse("1994-08-25"), "Three")

			entry := journal.GetClosestEntry(date.MustAutoParse("1994-08-01"))
			Expect(entry.EntryDate).To(Equal(date.MustAutoParse("1994-08-04")))
			Expect(entry.EntryText).To(Equal("One"))

			entry = journal.GetClosestEntry(date.MustAutoParse("1994-08-18"))
			Expect(entry.EntryDate).To(Equal(date.MustAutoParse("1994-08-20")))
			Expect(entry.EntryText).To(Equal("Two"))

			entry = journal.GetClosestEntry(date.MustAutoParse("1994-08-25"))
			Expect(entry.EntryDate).To(Equal(date.MustAutoParse("1994-08-25")))
			Expect(entry.EntryText).To(Equal("Three"))

			entry = journal.GetClosestEntry(date.MustAutoParse("1994-08-27"))
			Expect(entry.EntryDate).To(Equal(date.MustAutoParse("1994-08-25")))
			Expect(entry.EntryText).To(Equal("Three"))
		})
	})
})