package journaldrive_test

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/petergtz/alexa-journal/drive"
)

var _ = Describe("Drive", func() {
	var token string

	BeforeSuite(func() {
		token = os.Getenv("GOOGLE_DRIVE_TOKEN")
		if token == "" {
			panic("Please provide GOOGLE_DRIVE_TOKEN")
		}
	})

	It("can get and update content", func() {

		fileService := journaldrive.NewDriveJournalFileService(token, "test-journal.tsv")
		content := fileService.Content()
		fileService.Update(content + "\nanother line here")
		content2 := fileService.Content()
		Expect(content2).To(Equal(content + "\nanother line here"))
	})
})
