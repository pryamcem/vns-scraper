package cmd

import (
	"log"
	"strings"
	"time"

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

	// Create new browser.
	l := launcher.New().
		Headless(false).
		Devtools(false)

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
	err = storage.CreateSchemaByNum(testNum)
	if err != nil {
		log.Fatalf("Cant't create table schema: %v", err)
	}

	//storage.ParseToFile(10)
	//return

	i := 0
	trustValue := 10
	for {
		page.MustWaitLoad().MustNavigate(link)
		button, err := page.MustWaitLoad().Timeout(time.Second).ElementR("button", "Зробити наступну спробу")
		if err != nil {
			button, err = page.MustWaitLoad().Timeout(time.Second).ElementR("button", "Спроба тесту")
			if err != nil {
				button, err = page.MustWaitLoad().Timeout(time.Second).ElementR("button", "Продовжуйте свою спробу")
			}
		}
		button.MustClick()
		err = scruper.MakeTest(page, testNum, *storage)
		if err != nil {
			log.Fatalln("Test answering error:", err)
		}
		scruper.FinishTest(page)
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
