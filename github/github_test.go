package github_test

import (
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/petergtz/alexa-journal/github"
	"go.uber.org/zap"
)

var _ = Describe("Github", func() {
	It("can create an issue", func() {
		token, e := ioutil.ReadFile("../private/github-access-token")
		Expect(e).NotTo(HaveOccurred())
		l, e := zap.NewDevelopment()
		if e != nil {
			panic(e)
		}
		defer l.Sync()
		log := l.Sugar()

		er := NewGithubErrorReporter("petergtz", "alexa-journal", strings.TrimSpace(string(token)), log)

		er.ReportError("Testing: Some error occurred", errors.New("Some test error"))
	})
})
