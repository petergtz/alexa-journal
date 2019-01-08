package drive_test

import (
	"os"

	"go.uber.org/zap"

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
		l, e := zap.NewDevelopment()
		Expect(e).NotTo(HaveOccurred())
		fileService := drive.NewFileService(token, "journal-test", l.Sugar())
		content := fileService.Download()
		fileService.Upload(content + "\nanother line here")
		content2 := fileService.Download()
		Expect(content2).To(Equal(content + "\nanother line here"))
	})
})
