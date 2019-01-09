package custom_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDrive(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Custom Search Test Suite")
}
