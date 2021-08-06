package locale_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDynamodb(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "locale Suite")
}
