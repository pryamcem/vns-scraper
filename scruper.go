package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/go-rod/rod"
)

// login to VNS by login and password from config.
func login(page *rod.Page, data Configuration) error {
	err := page.WaitLoad()
	if err != nil {
		return fmt.Errorf("Can't load page")
	}
	loginEntry, err := page.Timeout(time.Second).Element("#username")
	if err != nil {
		return fmt.Errorf("Can't find login entry: %v", err)
	}
	loginEntry.MustInput(data.Login)
	passwordEntry, err := page.Timeout(time.Second).Element("#password")
	if err != nil {
		return fmt.Errorf("Can't find login entry: %v", err)
	}
	passwordEntry.MustInput(data.Password)
	loginBtn, err := page.Timeout(time.Second).Element("#loginbtn")
	if err != nil {
		return fmt.Errorf("Can't find login entry: %v", err)
	}
	loginBtn.MustClick()
	return nil
}

// mustFindTestNum finds test number to store answers to reqared table in databse.
// It panics if can't find test number in page.
func mustFindTestNum(page *rod.Page) int {
	page.MustWaitLoad()
	// find all <h2> elements on the page and loop through them
	h2Elements := page.MustElements("h2")
	for _, h2 := range h2Elements {
		text := h2.MustText()
		if strings.Contains(text, "Тест") {
			strs := strings.Fields(text)
			num, err := strconv.Atoi(strs[1])
			if err != nil {
				panic(err)
			}
			return num
		}
	}
	panic("Can't find test number")
}

// fundTests return list of links with title='Переглянути ваші відповіді в цій спробі'
func findTests(page *rod.Page) []string {
	var links []string
	elements := page.MustWaitLoad().MustElements("a[title='Переглянути ваші відповіді в цій спробі']")

	for _, link := range elements {
		href := link.MustAttribute("href")
		links = append(links, *href)
	}
	return links
}

func makeTest(page *rod.Page, testNum int, s Storage) error {
	page.MustWaitLoad()
	tests := page.MustElements(".formulation.clearfix")

	// Print the inner text of each element
	for _, element := range tests {
		question := element.MustElement(".qtext").MustText()
		rightanswer, err := s.PickRightanswer(testNum, question)
		if err != nil {
			return fmt.Errorf("Can't get rightanswer from storage: %w", err)
		}
		answers := element.MustElements(".flex-fill.ml-1")
		radioBoxes := element.MustElements("input[type='radio']")

		radioBoxes[0].MustClick()
		for i, a := range answers {
			if a.MustText() == rightanswer {
				radioBoxes[i].MustClick()
				break
			}
		}
	}
	return nil
}

// finishTest finds and clicks all nesesary buttons to complete test.
func finishTest(page *rod.Page) {
	button := page.MustElement("input[type='submit'][value='Завершити спробу...']")
	button.MustClick()
	button = page.MustWaitLoad().MustElementR("button", "Відправити все та завершити")
	button.MustClick()
	modal := page.MustElement(".modal-footer")
	button = modal.MustElementR("button", "Відправити все та завершити")
	button.MustClick()
}

// isTestSuccessful checks rate of correct answers.
// It returns true if test successful (100% of correct answers)
func isTestSuccessful(page *rod.Page) bool {
	cells := page.MustWaitLoad().MustElements(".cell")
	for _, cell := range cells {
		// Get the inner text of the element
		text := cell.MustText()
		fmt.Println(text)

		if strings.Contains(text, "(100%)") {
			return true
		}
	}
	return false
}

// parseAnswers find all open answers of complete test.
func parseAnswers(page *rod.Page) ([]QA, error) {
	qtexts := page.MustWaitLoad().MustElements(".qtext")
	for _, qtext := range qtexts {
		fmt.Println("Question Text:", qtext.MustText())
	}

	// Find all elements with the class "rightanswer"
	rightanswers := page.MustElements(".rightanswer")
	for _, rightanswer := range rightanswers {
		fmt.Println("Right Answer:", rightanswer.MustText())
	}
	var data []QA
	for i := range qtexts {
		data = append(data, QA{
			question:    strings.Join(strings.Fields(qtexts[i].MustText()), " "),
			rightanswer: rightanswers[i].MustText(),
		})
	}
	return data, nil
}
