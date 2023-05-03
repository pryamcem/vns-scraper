package main

import (
	"flag"
	"log"
	"time"

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

	err := login(page, config)
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
	//storage.ParseToFile(9)
	//return

	i := 0
	trustValue := 5
	for {
		page.MustWaitLoad().MustNavigate(*link)
		//button, err := page.MustWaitLoad().Timeout(time.Second).ElementR("button", "Спроба тесту")
		//if err != nil {
		button, err := page.MustWaitLoad().Timeout(time.Second).ElementR("button", "Зробити наступну спробу")

		//}
		//Timeout(1*time.Second).MustElementR("button", "Спроба тесту").
		//CancelTimeout().
		//Timeout(1*time.Second).MustElementR("button", "Продовжуйте свою спробу")
		button.MustClick()
		err = makeTest(page, testNum, *storage)
		if err != nil {
			log.Fatalln("Test answering error:", err)
		}
		finishTest(page)
		if isTestSuccessful(page) {
			i++
			if i == trustValue {
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
