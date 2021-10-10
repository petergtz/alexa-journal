package acceptance_test

import (
	"testing"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/petergtz/pegomock"
)

func TestAlexaJournal(t *testing.T) {

	gomega.RegisterFailHandler(ginkgo.Fail)
	pegomock.RegisterMockFailHandler(ginkgo.Fail)

	BeforeSuite(func() {
		format.TruncatedDiff = false
	})

	ginkgo.RunSpecs(t, "Acceptance Suite")
}
