package acceptance_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/format"
	"github.com/petergtz/alexa-journal/locale/resources"
	"github.com/rickb777/date"
)

type TestInvocation struct {
	utterance string
	response  string
}

var _ = Describe("Skill Acceptance", func() {
	today := date.Today().Format("2006-01-02")
	todayWithWeekDay := resources.Weekdays["DeDe"][date.Today().Weekday()] + ", " + today

	launchInvocation := TestInvocation{
		utterance: "Starte mein Tagebuch",
		response:  "Dein Tagebuch ist nun geöffnet. Was möchtest Du tun?",
	}

	newEntryForToday := []TestInvocation{{
		utterance: "Eintrag erstellen",
		response:  "Für welches Datum soll der Eintrag erstellt werden?",
	}, {
		utterance: "Heute",
		response:  "Du kannst Deinen Eintrag für den " + today + " nun verfassen. Los geht's!",
	}}

	BeforeSuite(func() {
		format.TruncatedDiff = false
	})

	It("can have a long conversation", func() {
		expectDialogToSucceed([]TestInvocation{
			launchInvocation,
			{
				utterance: "Eintrag löschen",
				response:  "Zu welchem Datum?",
			}, {
				utterance: "Heute",
				response:  "Hm. Zu diesem Datum habe ich leider keinen Eintrag gefunden.",
			},
			newEntryForToday[0], newEntryForToday[1],
			{
				utterance: "Das ist ein Test Eintrag",
				response:  "Ich wiederhole: das ist ein test eintrag.\n\nNaechster Teil bitte?",
			}, {
				utterance: "Dritter Teil",
				response:  "Ich wiederhole: dritter teil.\n\nNaechster Teil bitte?",
			}, {
				utterance: "Korrigieren",
				response:  `OK. Bitte verfasse den letzten Teil Deines Eintrags erneut.`,
			}, {
				utterance: "Zweiter Teil",
				response:  "Ich wiederhole: zweiter teil.\n\nNaechster Teil bitte?",
			}, {
				utterance: "Wiederhole",
				response:  "Ich wiederhole: zweiter teil\n\nNaechster Teil bitte?",
			}, {
				utterance: "Abbrechen",
				response:  "Okay. Abgebrochen.\n\n\n\nWas möchtest Du als nächstes in Deinem Tagebuch machen?",
			}, {
				utterance: "Eintrag erstellen",
				response:  "Für welches Datum soll der Eintrag erstellt werden?",
			}, {
				utterance: "Heute",
				response:  "Fuer dieses Datum hast Du bereits einen Eintrag entworfen. Er lautet: das ist ein test eintrag. zweiter teil. Moechtest Du mit diesem Eintrag weiter machen?",
			}, {
				utterance: "ja",
				response:  "Du kannst Deinen Eintrag für den " + today + " nun verfassen. Los geht's!",
			}, {
				utterance: "wiederhole",
				response:  "Ich wiederhole: zweiter teil\n\nNaechster Teil bitte?",
			}, {
				utterance: "fertig",
				response:  "Alles klar. Ich habe folgenden Eintrag für das Datum " + today + ": \"das ist ein test eintrag. zweiter teil\". Soll ich ihn so speichern?",
			}, {
				utterance: "Nein",
				response:  "Okay. Nicht gespeichert.\n\n\n\nWas möchtest Du als nächstes in Deinem Tagebuch machen?",
			}, {
				utterance: "Neuer Eintrag",
				response:  "Für welches Datum soll der Eintrag erstellt werden?",
			}, {
				utterance: "Heute",
				response:  "Fuer dieses Datum hast Du bereits einen Eintrag entworfen. Er lautet: das ist ein test eintrag. zweiter teil. Moechtest Du mit diesem Eintrag weiter machen?",
			}, {
				utterance: "Ja",
				response:  "Du kannst Deinen Eintrag für den " + today + " nun verfassen. Los geht's!",
			}, {
				utterance: "Fertig",
				response:  "Alles klar. Ich habe folgenden Eintrag für das Datum " + today + ": \"das ist ein test eintrag. zweiter teil\". Soll ich ihn so speichern?",
			}, {
				utterance: "Ja",
				response:  "Okay. Gespeichert.\n\n\n\nWas möchtest Du als nächstes in Deinem Tagebuch machen?",
			}, {
				utterance: "Eintrag vorlesen",
				response:  "Von welchem Datum soll ich einen Eintrag vorlesen?",
			}, {
				utterance: "Heute",
				response:  "Hier ist der Eintrag vom " + todayWithWeekDay + ": das ist ein test eintrag. zweiter teil.\n\nWas möchtest Du als nächstes in Deinem Tagebuch machen?",
			}, {
				utterance: "Suche nach test eintrag",
				// TODO: get rid of the unnecessary space.
				response: "Hier sind die Ergebnisse für die Suche \"test eintrag\": " + todayWithWeekDay + ": das ist ein test eintrag. zweiter teil. \n\nWas möchtest Du als nächstes in Deinem Tagebuch machen?",
			}, {
				utterance: "Eintrag löschen",
				response:  "Zu welchem Datum?",
			}, {
				utterance: "Heute",
				response:  "Du moechtest den folgenden Eintrag loeschen: das ist ein test eintrag. zweiter teil. Soll ich ihn wirklich loeschen?",
			}, {
				utterance: "Ja",
				response:  "Okay. Geloescht.\n\nWas möchtest Du als nächstes in Deinem Tagebuch machen?",
			},
		})
	})

	It("says there is nothing to repeat or correct when there's no new entry part yet", func() {
		expectDialogToSucceed([]TestInvocation{
			launchInvocation,
			newEntryForToday[0], newEntryForToday[1],
			{
				utterance: "Wiederholen",
				response:  `Dein Eintrag ist leer. Es gibt nichts zu wiederholen. Bitte verfasse zuerst den ersten Teil Deines Eintrags.`,
			}, {
				utterance: "Korrigieren",
				response:  `Dein Eintrag ist leer. Es gibt nichts zu korrigieren. Bitte verfasse zuerst den ersten Teil Deines Eintrags.`,
			},
		})
	})

	It("says entry is empty when it is empty", func() {
		expectDialogToSucceed([]TestInvocation{
			launchInvocation,
			newEntryForToday[0], newEntryForToday[1],
			{
				utterance: "Fertig",
				response:  "Dein Eintrag ist leer. Es gibt nichts zu speichern.\n\n\n\nWas möchtest Du als nächstes tun?",
			},
		})
	})

	It("says it when it can't find entries in time range", func() {
		expectDialogToSucceed([]TestInvocation{
			launchInvocation,
			{
				utterance: "Was war im mai neunzehn hundert fünf und zwanzig",
				response:  "Keine Einträge für den Zeitraum mai 1925 gefunden.\n\nWas möchtest Du als nächstes in Deinem Tagebuch machen?",
			},
		})
	})

	It("says it when it can't can't find a search", func() {
		expectDialogToSucceed([]TestInvocation{
			launchInvocation,
			{
				utterance: "Suche nach albert einstein relativitätstheorie",
				response:  "Keine Einträge für die Suche \"albert einstein relativitätstheorie\" gefunden.\n\nWas möchtest Du als nächstes in Deinem Tagebuch machen?",
			},
		})
	})

	It("provides help", func() {
		expectDialogToSucceed([]TestInvocation{
			launchInvocation,
			{
				utterance: "Hilfe",
				response:  `Mit diesem Skill kannst Du Tagebucheintraege erstellen oder vorlesen lassen. Sage z.B. "Neuen Eintrag erstellen". Oder "Lies mir den Eintrag von gestern vor". Oder "Was war heute vor 20 Jahren?". Oder "Was war im August 1994?".`,
			},
		})
	})
})

func expectDialogToSucceed(dialog []TestInvocation) {
	actualDialog := run(utterancesFrom(dialog))
	Expect(actualDialog).To(HaveLen(len(dialog)))
	for i := range dialog {
		Expect(actualDialog[i].utterance).To(Equal(dialog[i].utterance))
		Expect(actualDialog[i].response).To(Equal(dialog[i].response))
	}
}

func run(utterances []string) []TestInvocation {
	utterancesTempFile := writeUtterancesToTempFile(utterances)
	defer os.Remove(utterancesTempFile.Name())

	outputFile, e := os.CreateTemp("", "")
	Expect(e).NotTo(HaveOccurred())
	outputFile.Close()
	defer os.Remove(outputFile.Name())
	ask := exec.Command("ask", "dialog", "--stage", "development", "-r", utterancesTempFile.Name(), "--save-skill-io", outputFile.Name())
	ask.Stdout = os.Stdout
	ask.Stderr = os.Stderr
	Expect(ask.Run()).To(Succeed())

	return dialogFrom(outputFile.Name())
}

func utterancesFrom(dialog []TestInvocation) []string {
	var utterances []string
	for _, invocation := range dialog {
		utterances = append(utterances, `"`+invocation.utterance+`"`)
	}
	return utterances
}

func writeUtterancesToTempFile(utterances []string) *os.File {
	file, e := os.CreateTemp("", "")
	Expect(e).NotTo(HaveOccurred())
	defer file.Close()
	fmt.Fprintf(file, `{
		"skillId": "amzn1.ask.skill.ad1669b4-291c-4daa-9fbb-fa32b8ea3078",
		"locale": "de-DE",
		"type": "text",
		"userInput": [
		  %v,
		  ".quit"
		]
	  }
	  `, strings.Join(utterances, ",\n"))
	return file
}

func dialogFrom(outputFilename string) []TestInvocation {
	content, e := ioutil.ReadFile(outputFilename)
	Expect(e).NotTo(HaveOccurred())
	var skillIO SkillIO
	e = json.Unmarshal(content, &skillIO)
	Expect(e).NotTo(HaveOccurred())
	var testInvocations []TestInvocation
	for _, invocation := range skillIO.Invocations {
		Expect(invocation.Response.Body.Status).To(Equal("SUCCESSFUL"), "UNSUCCESSFUL BODY:\n\n%#v\n\nFULL FILE CONTENT:\n\n%v", invocation.Response.Body, string(content))
		if len(invocation.Response.Body.Result.AlexaExecutionInfo.AlexaResponses) != 1 {
			ginkgo.Fail(fmt.Sprintf("Unexpected output in AlexaResponses for %#v", invocation))
		}
		testInvocations = append(testInvocations, TestInvocation{
			utterance: invocation.Request.Utterance,
			response:  invocation.Response.Body.Result.AlexaExecutionInfo.AlexaResponses[0].Content.Caption,
		})
	}
	return testInvocations
}

type SkillIO struct {
	Invocations []Invocation
}

type Invocation struct {
	Request  Request
	Response Response
}

type Request struct {
	Utterance  string
	NewSession bool
}

type Response struct {
	Body Body
}

type Body struct {
	Status string
	Result Result
}

type Result struct {
	AlexaExecutionInfo AlexaExecutionInfo
}

type AlexaExecutionInfo struct {
	AlexaResponses []AlexaResponse
}

type AlexaResponse struct {
	Type    string
	Content Content
}

type Content struct {
	Caption string
}
