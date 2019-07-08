package main_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAlexaJournal(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "AlexaJournal Suite")
}
