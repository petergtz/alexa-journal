package drive_test

import (
	"os"

	"go.uber.org/zap"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/petergtz/alexa-journal/drive"
)

var _ = Describe("Drive", func() {
	var (
		token string
		log   *zap.Logger
	)

	BeforeSuite(func() {
		token = os.Getenv("GOOGLE_DRIVE_TOKEN")
		if token == "" {
			panic("Please provide GOOGLE_DRIVE_TOKEN")
		}
		var e error
		log, e = zap.NewDevelopment()
		Expect(e).NotTo(HaveOccurred())
	})

	Describe("FileService", func() {
		It("can download and upload content", func() {
			fileService, e := drive.NewFileService(token, "journal-test", log.Sugar())
			Expect(e).NotTo(HaveOccurred())
			defer drive.DeleteFile(token, fileService.FileID)

			content, e := fileService.Download()
			Expect(e).NotTo(HaveOccurred())
			fileService.Upload(content + "\nanother line here")
			content2, e := fileService.Download()
			Expect(e).NotTo(HaveOccurred())

			Expect(content2).To(Equal(content + "\nanother line here"))
		})
	})

	Describe("SheetBasedTabularData", func() {
		It("can append rows and read rows", func() {
			sheetsService, e := drive.NewSheetBasedTabularData(token, "journal-test", "my-sheet", log.Sugar())
			Expect(e).NotTo(HaveOccurred())
			defer drive.DeleteFile(token, sheetsService.SpreadsheetID)

			e = sheetsService.AppendRow([]string{"a", "b", "c"})
			Expect(e).NotTo(HaveOccurred())
			e = sheetsService.AppendRow([]string{"d", "e", "f"})
			Expect(e).NotTo(HaveOccurred())

			Expect(sheetsService.Rows()).To(Equal([][]string{
				[]string{"a", "b", "c"},
				[]string{"d", "e", "f"},
			}))
		})
	})
})
