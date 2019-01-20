package journal_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	j "github.com/petergtz/alexa-journal/journal"
	"github.com/petergtz/alexa-journal/tsv"
	"github.com/rickb777/date"
)

var _ = Describe("Journal", func() {
	var journal j.Journal

	BeforeEach(func() {
		journal = j.Journal{Data: &tsv.StringBasedTabularData{}}
	})

	Describe("date/time formatting and parsing", func() {
		It("works", func() {
			t := time.Now().UTC()
			s := t.Format(j.TimestampFormat)
			Expect(time.Parse(j.TimestampFormat, s)).To(Equal(t.Truncate(time.Second)))
		})
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

			entry, e := journal.GetClosestEntry(date.MustAutoParse("1994-08-01"))
			Expect(e).NotTo(HaveOccurred())
			Expect(entry.EntryDate).To(Equal(date.MustAutoParse("1994-08-04")))
			Expect(entry.EntryText).To(Equal("One"))

			entry, e = journal.GetClosestEntry(date.MustAutoParse("1994-08-18"))
			Expect(e).NotTo(HaveOccurred())
			Expect(entry.EntryDate).To(Equal(date.MustAutoParse("1994-08-20")))
			Expect(entry.EntryText).To(Equal("Two"))

			entry, e = journal.GetClosestEntry(date.MustAutoParse("1994-08-25"))
			Expect(e).NotTo(HaveOccurred())
			Expect(entry.EntryDate).To(Equal(date.MustAutoParse("1994-08-25")))
			Expect(entry.EntryText).To(Equal("Three"))

			entry, e = journal.GetClosestEntry(date.MustAutoParse("1994-08-27"))
			Expect(e).NotTo(HaveOccurred())
			Expect(entry.EntryDate).To(Equal(date.MustAutoParse("1994-08-25")))
			Expect(entry.EntryText).To(Equal("Three"))
		})
	})
})
