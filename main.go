package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/go-rod/rod"
)

const (
	configPath = "config.json"
)

type QA struct {
	question    string
	rightanswer string
}

func main() {
	link := flag.String("link", "", "Link to test")
	flag.Parse()

	config := GetConfig(configPath)

	// Create new browser.
	browser := rod.New().MustConnect()
	defer browser.Close()
	page := browser.MustPage(*link)

	err := Login(page, config)
	if err != nil {
		log.Fatalln(err)
	}
	testNum := mustFindTestNum(page)
	storage, err := New("tests.db")
	err = storage.InitByNum(testNum)
	if err != nil {
		log.Fatalf("Login error: %v", err)
	}
	defer storage.Close()
	//storage.ParseToFile(7)

	i := 0
	for {
		page.MustWaitLoad().MustNavigate(*link)
		button := page.MustWaitLoad().MustElementR("button", "Зробити наступну спробу")
		//Timeout(1*time.Second).MustElementR("button", "Спроба тесту").
		//CancelTimeout().
		//Timeout(1*time.Second).MustElementR("button", "Продовжуйте свою спробу")
		button.MustClick()
		err := makeTest(page, testNum, *storage)
		if err != nil {
			log.Fatalln("Test answering error:", err)
		}
		finishTest(page)
		if isTestSuccessful(page) {
			i++
			if i == 10 {
				break
			}
		} else {
			i = 0
			data, err := parseAnswers(page)
			if err != nil {
				log.Fatalln("Can't parse answers:", err)
			}
			for _, d := range data {
				err := storage.PutQA(testNum, d)
				if err != nil {
					log.Fatalln("Cant insert data to storage:", err)
				}
			}
		}
	}
}

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

func finishTest(page *rod.Page) {
	button := page.MustElement("input[type='submit'][value='Завершити спробу...']")
	button.MustClick()
	button = page.MustWaitLoad().MustElementR("button", "Відправити все та завершити")
	button.MustClick()
	modal := page.MustElement(".modal-footer")
	button = modal.MustElementR("button", "Відправити все та завершити")
	button.MustClick()
}

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
