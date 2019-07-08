package main_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/petergtz/alexa-journal"
	"github.com/rickb777/date"
)

var _ = Describe("Date", func() {
	It("works", func() {
		dayDate, monthDate, dateType := DateFrom("2019-01-XX", "")
		Expect(dateType).To(BeEquivalentTo(MonthDate))
		Expect(dayDate.IsZero()).To(BeTrue())
		Expect(monthDate).To(Equal("2019-01"))

		dayDate, monthDate, dateType = DateFrom("2019-01", "")
		Expect(dateType).To(BeEquivalentTo(MonthDate))
		Expect(dayDate.IsZero()).To(BeTrue())
		Expect(monthDate).To(Equal("2019-01"))

		dayDate, monthDate, dateType = DateFrom("2019-01", "1997")
		Expect(dateType).To(BeEquivalentTo(MonthDate))
		Expect(dayDate.IsZero()).To(BeTrue())
		Expect(monthDate).To(Equal("1997-01"))

		dayDate, monthDate, dateType = DateFrom("2019-01-01", "1997")
		Expect(dateType).To(BeEquivalentTo(DayDate))
		Expect(dayDate).To(Equal(date.New(1997, time.January, 1)))
		Expect(monthDate).To(BeEmpty())

		dayDate, monthDate, dateType = DateFrom("2019-01-01", "")
		Expect(dateType).To(BeEquivalentTo(DayDate))
		Expect(dayDate).To(Equal(date.New(2019, time.January, 1)))
		Expect(monthDate).To(BeEmpty())
	})

})
