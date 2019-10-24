package github_test

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sns"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/petergtz/alexa-journal/github"
	"go.uber.org/zap"
)

var _ = Describe("Github", func() {
	FIt("can create an issue", func() {
		token, e := ioutil.ReadFile("../private/github-access-token")
		Expect(e).NotTo(HaveOccurred())
		l, e := zap.NewDevelopment()
		if e != nil {
			panic(e)
		}
		defer l.Sync()
		log := l.Sugar()

		er := NewGithubErrorReporter(
			"petergtz",
			"alexa-journal",
			strings.TrimSpace(string(token)),
			log,
			"logsUrl %v",
			sns.New(session.Must(session.NewSession(&aws.Config{Region: aws.String("eu-west-1")}))),
			"arn:aws:sns:eu-west-1:512841817041:AlexaJournalErrors",
		)

		er.ReportPanic("Testing: Some error occurred")
	})

	It("can publish on SNS topic", func() {
		snsClient := sns.New(session.Must(session.NewSession(&aws.Config{Region: aws.String("eu-west-1")})))
		_, e := snsClient.Publish(&sns.PublishInput{
			TopicArn: aws.String("arn:aws:sns:eu-west-1:512841817041:AlexaJournalErrors"),
			Subject:  aws.String(fmt.Sprintf("AlexaJournal: Internal Server Error (ErrID: %v)", 12345)),
			Message:  aws.String("Error:\n" + "This will become the content of the error message"),
		})
		Expect(e).NotTo(HaveOccurred())
	})
})
