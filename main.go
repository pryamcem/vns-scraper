package main

import (
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/go-rod/rod"
)

type QA struct {
	testnum     int
	question    string
	rightanswer string
}

func main() {
	storage, err := New("tests.db")
	err = storage.Init()
	if err != nil {
		log.Fatalln(err)
	}
	defer storage.Close()

	link := flag.String("link", "", "Link to test")
	login := flag.String("login", "", "VNS login")
	password := flag.String("password", "", "VNS password")
	//iter := flag.Int("iter", 1, "Iterations to generate dataset")
	//dir := flag.String("d", ".", "directory with files to parse")
	flag.Parse()

	// Create new browser.
	browser := rod.New().MustConnect()
	defer browser.Close()

	// Login to VNS by lign and password from flags.
	// TODO: Move login to separete function wich return error if login unsucsessfull.
	page := browser.MustPage(*link)
	page.MustElement("#username").MustInput(*login)
	page.MustElement("#password").MustInput(*password)
	page.MustElement("#loginbtn").MustClick()

	//page.MustWaitLoad().MustNavigate(*link)
	for {
		page.MustWaitLoad().MustNavigate(*link)
		button := page.MustWaitLoad().MustElementR("button", "Зробити наступну спробу")
		button.MustClick()
		//answerTest(*link, page)
		err := makeTest(page, *storage)
		if err != nil {
			//log.Fatalln("Test answering error:", err)
		}
		finishTest(page)
		if isSuccessful(page) {
			break
		} else {
			data, err := parseAnswers(page)
			if err != nil {
				log.Fatalln("Can't parse answers:", err)
			}
			for _, d := range data {
				err := storage.PutQA(d)
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

func makeTest(page *rod.Page, s Storage) error {
	page.MustWaitLoad()
	tests := page.MustElements(".formulation.clearfix")

	// Print the inner text of each element
	for _, element := range tests {
		question := element.MustElement(".qtext").MustText()
		fmt.Println("QUESTION", question)
		rightanswer, err := s.GetRightanswer(question)
		fmt.Println("RIGHTANSWER", rightanswer)
		if err != nil {
			return fmt.Errorf("Can't get rightanswer from storage: %w", err)
		}
		answers := element.MustElements(".flex-fill.ml-1")
		radioBoxes := element.MustElements("input[type='radio']")

		radioBoxes[0].MustClick()
		for i, a := range answers {
			fmt.Println("ANSWER", a.MustText())
			if a.MustText() == rightanswer {
				fmt.Println("RIGHT")
				radioBoxes[i].MustClick()
				break
			} else {
				fmt.Println("FALSE", a.MustText(), rightanswer)
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

func isSuccessful(page *rod.Page) bool {
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
