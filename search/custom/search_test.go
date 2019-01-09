package custom_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/petergtz/alexa-journal/search/custom"
)

var _ = Describe("CustomIndex", func() {
	It("finds correct entries", func() {
		file, e := os.Open("../../private/my-journal.tsv")
		Expect(e).NotTo(HaveOccurred())
		defer file.Close()
		b, e := ioutil.ReadAll(file)
		Expect(e).NotTo(HaveOccurred())

		index := custom.NewSearchIndex(nil)

		for _, line := range strings.Split(string(b), "\n") {
			parts := strings.Split(line, "\t")
			if len(parts) != 3 {
				continue
			}
			index.Add(parts[1], parts[2])
		}

		hits := index.Search("Dampfmaschine")
		for _, line := range strings.Split(string(b), "\n") {
			parts := strings.Split(line, "\t")
			if len(parts) != 3 {
				continue
			}
			for _, hit := range hits {
				if parts[1] == hit.Result {
					fmt.Println(hit.Confidence, parts[1], parts[2])
				}
			}
		}
	})
})
