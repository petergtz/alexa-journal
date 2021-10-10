package acceptance_test

import (
	. "github.com/onsi/ginkgo"
	"github.com/petergtz/alexa-journal/locale/resources"
	"github.com/rickb777/date"
)

var _ = Describe("Skill Acceptance EN", func() {
	today := date.Today().Format("2006-01-02")
	todayWithWeekDay := resources.Weekdays["EnUs"][date.Today().Weekday()] + ", " + today

	launchInvocation := TestInvocation{
		utterance: "open my journal",
		response:  "Okay, your journal is open. What do you want to do next?",
	}

	newEntryForToday := []TestInvocation{{
		utterance: "new entry",
		response:  "For which date should I create the new entry?",
	}, {
		utterance: "Today",
		response:  "You can draft your entry for " + today + " now; let's go!",
	}}

	It("can have a long conversation", func() {
		expectDialogToSucceed("en-US", []TestInvocation{
			launchInvocation,
			{
				utterance: "Delete entry",
				response:  "For which date?",
			}, {
				utterance: "Today",
				response:  "Um. I couldn't find an entry for this date.",
			},
			newEntryForToday[0], newEntryForToday[1],
			{
				utterance: "This is a test entry",
				response:  "I repeat: this is a test entry.\n\nNext part please?",
			}, {
				utterance: "Third part",
				response:  "I repeat: third part.\n\nNext part please?",
			}, {
				utterance: "Correct",
				response:  `OK. Please draft the last part of your entry again.`,
			}, {
				utterance: "Second part",
				response:  "I repeat: second part.\n\nNext part please?",
			}, {
				utterance: "Repeat",
				response:  "I repeat: second part\n\nNext part please?",
			}, {
				utterance: "Abort",
				response:  "Okay. Aborted.\n\n\n\nWhat do you want to do next in your journal?",
			},
			newEntryForToday[0],
			{
				utterance: "Today",
				response:  "A draft for this date already exists. It is: this is a test entry. second part. Do you want to continue with that?",
			}, {
				utterance: "yes",
				response:  "You can draft your entry for " + today + " now; let's go!",
			}, {
				utterance: "repeat",
				response:  "I repeat: second part\n\nNext part please?",
			}, {
				utterance: "done",
				response:  "Alright. I have the following entry for " + today + ": \"this is a test entry. second part\".  Should I save it like this?",
			}, {
				utterance: "No",
				response:  "Okay. Not saved.\n\n\n\nWhat do you want to do next in your journal?",
			},
			newEntryForToday[0],
			{
				utterance: "today",
				response:  "A draft for this date already exists. It is: this is a test entry. second part. Do you want to continue with that?",
			}, {
				utterance: "yes",
				response:  "You can draft your entry for " + today + " now; let's go!",
			}, {
				utterance: "done",
				response:  "Alright. I have the following entry for " + today + ": \"this is a test entry. second part\".  Should I save it like this?",
			}, {
				utterance: "yes",
				response:  "Okay. Saved.\n\n\n\nnWhat do you want to do next in your journal?",
			}, {
				utterance: "Read an entry",
				response:  "From what date should I read an entry?",
			}, {
				utterance: "Today",
				response:  "Here's the entry from " + todayWithWeekDay + ": this is a test entry. second part.\n\nWhat do you want to do next in your journal?",
			}, {
				utterance: "Search for test entry",
				// TODO: get rid of the unnecessary space.
				response: "Here are the results for the query \"test entry\": " + todayWithWeekDay + ": this is a test entry. second part. \n\nWhat do you want to do next in your journal?",
			}, {
				utterance: "Delete entry",
				response:  "For which date?",
			}, {
				utterance: "today",
				response:  "You'd like to delete the following entry: this is a test entry. second part. Should I really delete it?",
			}, {
				utterance: "yes",
				response:  "Okay. Deleted.\n\nWhat do you want to do next in your journal?",
			},
		})
	})

	It("says there is nothing to repeat or correct when there's no new entry part yet", func() {
		expectDialogToSucceed("en-US", []TestInvocation{
			launchInvocation,
			newEntryForToday[0], newEntryForToday[1],
			{
				utterance: "Repeat",
				response:  `Your entry is empty. There's nothing to repeat. Please draft your first part of your entry.`,
			}, {
				utterance: "Correct",
				response:  `Your entry is empty. There's nothing to correct. Please draft your first part of your entry.`,
			},
		})
	})

	It("says entry is empty when it is empty", func() {
		expectDialogToSucceed("en-US", []TestInvocation{
			launchInvocation,
			newEntryForToday[0], newEntryForToday[1],
			{
				utterance: "Done",
				response:  "Your entry is empty. There is nothing to save.\n\n\n\nWas möchtest Du als nächstes tun?",
			},
		})
	})

	It("says it when it can't find entries in time range", func() {
		expectDialogToSucceed("en-US", []TestInvocation{
			launchInvocation,
			{
				utterance: "What was in May nineteen twenty five",
				response:  "No entries found for time range may 1925.\n\nWhat do you want to do next in your journal?",
			},
		})
	})

	It("says it when it can't can't find a search", func() {
		expectDialogToSucceed("en-US", []TestInvocation{
			launchInvocation,
			{
				utterance: "Search for albert einstein theory of relativity",
				response:  "I couldn't find any entries for the query \"einstein theory of relativity\".\n\nWhat do you want to do next in your journal?",
			},
		})
	})

	It("provides help", func() {
		expectDialogToSucceed("en-US", []TestInvocation{
			launchInvocation,
			{
				utterance: "Help",
				response:  `With this skill, you create journal entries and have them read for you. Say e.g. \"new entry\" . Or \"Read the entry from yesterday\". Or \"What was 20 years ago today?\". Oder \"What was in August 1994?\".`,
			},
		})
	})
})
