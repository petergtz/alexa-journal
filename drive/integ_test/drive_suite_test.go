package journaldrive_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDrive(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Drive Integ Test Suite")
}
