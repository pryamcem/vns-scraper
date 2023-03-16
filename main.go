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

	//for i := 0; i < *iter; i++ {
	//button := page.MustWaitLoad().MustElementR("button", "Зробити наступну спробу")
	//button.MustClick()
	//answerTest(*link, page)
	//}

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

func makeTest(page *rod.Page) error {
	questions := page.MustWaitLoad().MustElements(".formulation.clearfix")
	for _, element := range questions {
		//TODO: Find here question text and put it to GetRightanswer()
		// And then use the resoult of GetRightanswer() to find and click right radio box.
		radioBoxes := element.MustElements("input[type='radio']")
		radioBoxes[0].MustClick()
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
	page.MustWaitLoad()
}

func answerTest(link string, page *rod.Page) {
	questions := page.MustWaitLoad().MustElements(".formulation.clearfix")

	// Print the inner text of each element
	for _, element := range questions {
		radioBoxes := element.MustElements("input[type='radio']")
		radioBoxes[0].MustClick()
	}
	//button := page.MustWaitLoad().MustElementR("button", "Завершити спробу...")
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
