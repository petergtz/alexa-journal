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
			fileService := drive.NewFileService(token, "journal-test", log.Sugar())
			defer drive.DeleteFile(token, fileService.FileID)

			content := fileService.Download()
			fileService.Upload(content + "\nanother line here")
			content2 := fileService.Download()

			Expect(content2).To(Equal(content + "\nanother line here"))
		})
	})

	Describe("SheetBasedTabularData", func() {
		It("can append rows and read rows", func() {
			sheetsService := drive.NewSheetBasedTabularData(token, "journal-test", "my-sheet", log.Sugar())
			defer drive.DeleteFile(token, sheetsService.SpreadsheetID)

			sheetsService.AppendRow([]string{"a", "b", "c"})
			sheetsService.AppendRow([]string{"d", "e", "f"})

			Expect(sheetsService.Rows()).To(Equal([][]string{
				[]string{"a", "b", "c"},
				[]string{"d", "e", "f"},
			}))
		})
	})
})
