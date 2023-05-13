package cmd

import (
	"log"
	"strings"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/pryamcem/vns-scraper/config"
	"github.com/pryamcem/vns-scraper/scruper"
	"github.com/pryamcem/vns-scraper/storage"
	"github.com/spf13/cobra"
)

const (
	configPath = "config.json"
)

var passCmd = &cobra.Command{
	Use:   "pass",
	Short: "Pass the test.",
	Run:   pass,
}

func pass(_ *cobra.Command, args []string) {
	link := args[0]
	config := config.GetConfig(configPath)

	// Init storage.
	storage, err := storage.New("tests.db")
	if err != nil {
		log.Fatalf("Storage initialization error: %v", err)
	}
	defer storage.Close()

	// Create new browser.
	l := launcher.New().
		Headless(false).
		Devtools(true)

	defer l.Cleanup()
	url := l.MustLaunch()

	browser := rod.New().ControlURL(url).Trace(true).MustConnect()
	defer browser.Close()

	//cmd.Execute()

	page := browser.MustPage(link)

	err = scruper.Login(page, config)
	if err != nil {
		log.Fatalln("Can't login: ", err)
	}
	testNum := scruper.MustFindTestNum(page)
	err = storage.CreateTableByNum(testNum)
	if err != nil {
		log.Fatalf("Cant't create table schema: %v", err)
	}

	i := 0
	trustValue := 10
	for {
		//Go to the test link
		page.MustWaitLoad().MustNavigate(link)

		err := scruper.StartNextAttempt(page)
		if err != nil {
			log.Fatalln("Error while startin new attempt:", err)
		}

		isLastPage := false
		for !isLastPage {
			err = scruper.MakeTest(page, testNum, *storage)
			if err != nil {
				log.Fatalln("Error while making test: ", err)
			}
			_, isLastPage = scruper.FinishTest(page)
		}

		if scruper.IsTestSuccessful(page) {
			i++
			if i == trustValue {
				break
			}
		} else {
			i = 0
			data, err := scruper.ParseAnswers(page)
			if err != nil {
				log.Fatalln("Can't parse answers:", err)
			}
			for _, d := range data {
				_, rightanswerFormated, ok := strings.Cut(d.Rightanswer, "Правильна відповідь: ")
				if !ok {
					log.Fatalln("Can't format answer")
				}

				err := storage.Put(testNum, d.Question, rightanswerFormated)
				if err != nil {
					log.Fatalln("Cant insert data to storage:", err)
				}
			}
		}
	}
}
