package journaldrive_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/petergtz/alexa-journal/drive"
	"github.com/rickb777/date"
)

var _ = Describe("Drive", func() {
	var token string

	BeforeSuite(func() {
		token = os.Getenv("GOOGLE_DRIVE_TOKEN")
		if token == "" {
			panic("Please provide GOOGLE_DRIVE_TOKEN")
		}
	})

	It("can add entry", func() {
		journal := journaldrive.NewJournal(token, "test-journal.tsv")
		d := date.Today().Add(-10)
		e := journal.AddEntry(d, "Test")
		Expect(e).NotTo(HaveOccurred())

		text, e := journal.GetEntry(d)
		Expect(e).NotTo(HaveOccurred())
		Expect(text).To(Equal("Test"))
	})
})
