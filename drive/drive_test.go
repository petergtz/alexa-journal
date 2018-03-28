package journaldrive_test

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/petergtz/alexa-journal/drive"
	"golang.org/x/oauth2"
	"google.golang.org/api/drive/v3"
)

var token = os.Getenv("GOOGLE_DRIVE_TOKEN")

var _ = Describe("Drive", func() {
	It("can create", func() {
		d, e := drive.New(
			oauth2.NewClient(context.TODO(), oauth2.StaticTokenSource(&oauth2.Token{
				AccessToken: token,
			})))
		Expect(e).NotTo(HaveOccurred())
		file, e := d.Files.Create(&drive.File{
			Name: "pego-journal-test-file3",
		}).Media(strings.NewReader("Hello")).Do()
		Expect(e).NotTo(HaveOccurred())
		fmt.Println("File", file)
	})

	FIt("can list and get", func() {
		journal, e := journaldrive.NewJournal(token)
		Expect(e).NotTo(HaveOccurred())
		e = journal.AddEntry(time.Now(), "Test")
		Expect(e).NotTo(HaveOccurred())

	})

})
