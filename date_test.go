package journalskill_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/petergtz/alexa-journal"
	"github.com/rickb777/date"
)

var _ = Describe("Date", func() {
	It("works", func() {
		dayDate, monthDate, dateType := DateFrom("2019-01-XX")
		Expect(dateType).To(BeEquivalentTo(MonthDate))
		Expect(dayDate.IsZero()).To(BeTrue())
		Expect(monthDate).To(Equal("2019-01"))

		dayDate, monthDate, dateType = DateFrom("2019-01")
		Expect(dateType).To(BeEquivalentTo(MonthDate))
		Expect(dayDate.IsZero()).To(BeTrue())
		Expect(monthDate).To(Equal("2019-01"))

		dayDate, monthDate, dateType = DateFrom("2019-01-01")
		Expect(dateType).To(BeEquivalentTo(DayDate))
		Expect(dayDate).To(Equal(date.New(2019, time.January, 1)))
		Expect(monthDate).To(BeEmpty())

		dayDate, monthDate, dateType = DateFrom("")
		Expect(dateType).To(BeEquivalentTo(Invalid))
		Expect(dayDate.IsZero()).To(BeTrue())
		Expect(monthDate).To(BeEmpty())

		dayDate, monthDate, dateType = DateFrom("2017-XX-XX")
		Expect(dateType).To(BeEquivalentTo(YearDate))
		Expect(dayDate.IsZero()).To(BeTrue())
		Expect(monthDate).To(BeEmpty())

		dayDate, monthDate, dateType = DateFrom("XXXX-XX-27")
		Expect(dateType).To(BeEquivalentTo(DayDate))
		Expect(dayDate.Day()).To(Equal(27))
		Expect(dayDate.Month()).To(Equal(date.Today().Month()))
		Expect(dayDate.Year()).To(Equal(date.Today().Year()))
		Expect(monthDate).To(BeEmpty())

		dayDate, monthDate, dateType = DateFrom("XX19-12-08")
		Expect(dateType).To(BeEquivalentTo(DayDate))
		Expect(dayDate).To(Equal(date.New(2019, time.December, 8)))
		Expect(monthDate).To(BeEmpty())

		dayDate, monthDate, dateType = DateFrom("XXXX-XX-02")
		Expect(dateType).To(BeEquivalentTo(DayDate))
		today := time.Now()
		Expect(dayDate).To(Equal(date.New(today.Year(), today.Month(), 2)))
		Expect(monthDate).To(BeEmpty())
	})

})
