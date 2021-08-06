package dynamodb_test

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	journalskill "github.com/petergtz/alexa-journal"
	"github.com/petergtz/alexa-journal/dynamodb"
	"github.com/petergtz/go-alexa"
)

var _ = Describe("ConfigService", func() {
	It("works", func() {
		// logger, e := zap.NewDevelopment()
		// Expect(e).NotTo(HaveOccurred())
		// if os.Getenv("ACCESS_KEY_ID") == "" {
		// 	logger.Fatal("env var ACCESS_KEY_ID not provided.")
		// }

		// if os.Getenv("SECRET_ACCESS_KEY") == "" {
		// 	logger.Fatal("env var SECRET_ACCESS_KEY not provided.")
		// }

		cs := dynamodb.CreateConfigService("TestAlexaJournalConfig", "eu-central-1", &StdOutErrorReporter{})

		e := cs.PersistConfig("someUserID", journalskill.Config{BeSuccinct: true, ShouldExplainAboutSuccinctMode: false})
		Expect(e).NotTo(HaveOccurred())

		fmt.Println(cs.GetConfig("someUserID"))
		fmt.Println(cs.GetConfig("someOtherUserID"))
	})
})

type StdOutErrorReporter struct{}

func (*StdOutErrorReporter) ReportError(error)                               {}
func (*StdOutErrorReporter) ReportPanic(interface{}, *alexa.RequestEnvelope) {}
