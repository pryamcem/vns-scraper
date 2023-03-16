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
	iter := flag.Int("iter", 1, "Iterations to generate dataset")
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

	for i := 0; i < *iter; i++ {
		button := page.MustWaitLoad().MustElementR("button", "Зробити наступну спробу")
		button.MustClick()
		//answerTest(*link, page)
		err := makeTest(page, *storage)
		if err != nil {
			//log.Fatalln("Test answering error:", err)
		}
		finishTest(page, *link)
	}

	links := findTests(page)
	var dataset []QA
	for _, l := range links {
		p := browser.MustPage(l)
		//p.MustWaitLoad().MustScreenshot(fmt.Sprintf("%d.png", i))
		data, err := parseAnswers(p)
		if err != nil {
		}
		for _, d := range data {
			err := storage.PutQA(d)
			if err != nil {
				log.Fatalln("Cant insert data to storage:", err)
			}
		}
		dataset = append(dataset, data...)
		p.Close()
	}

	setSize := len(dataset)
	dataset = removeDuplicates(dataset)
	for i := range dataset {
		fmt.Printf("%s\n%s\n\n", dataset[i].question, dataset[i].rightanswer)
	}
	fmt.Println("Set size:", setSize, "Unique: ", len(dataset))
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

func finishTest(page *rod.Page, link string) {
	button := page.MustElement("input[type='submit'][value='Завершити спробу...']")
	button.MustClick()
	button = page.MustWaitLoad().MustElementR("button", "Відправити все та завершити")
	button.MustClick()
	modal := page.MustElement(".modal-footer")
	button = modal.MustElementR("button", "Відправити все та завершити")
	button.MustClick()
	page.MustWaitLoad()
	page.MustNavigate(link)
}

//func answerTest(link string, page *rod.Page, s Storage) error {
////button := page.MustWaitLoad().MustElementR("button", "Завершити спробу...")
//button := page.MustElement("input[type='submit'][value='Завершити спробу...']")
//button.MustClick()
//button = page.MustWaitLoad().MustElementR("button", "Відправити все та завершити")
//button.MustClick()
//modal := page.MustElement(".modal-footer")
//button = modal.MustElementR("button", "Відправити все та завершити")
//button.MustClick()
//page.MustWaitLoad()
//page.MustNavigate(link)
//return nil
//}

func removeDuplicates(strSlice []QA) []QA {
	allKeys := make(map[QA]bool)
	list := []QA{}
	for _, item := range strSlice {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}