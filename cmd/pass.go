package cmd

import (
	"log"
	"time"

	"github.com/go-rod/rod"
	"github.com/pryamcem/VNS-scraper/config"
	"github.com/pryamcem/VNS-scraper/scruper"
	"github.com/pryamcem/VNS-scraper/storage"
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
	_, _, _ = link, config, storage

	// Create new browser.
	browser := rod.New().MustConnect()
	defer browser.Close()

	//cmd.Execute()

	page := browser.MustPage(link)

	err = scruper.Login(page, config)
	if err != nil {
		log.Fatalln(err)
	}
	testNum := scruper.MustFindTestNum(page)
	err = storage.SchemaByNum(testNum)
	if err != nil {
		log.Fatalf("Cant't create table schema: %v", err)
	}

	//storage.ParseToFile(10)
	//return

	i := 0
	trustValue := 10
	for {
		page.MustWaitLoad().MustNavigate(link)
		button, err := page.MustWaitLoad().Timeout(time.Second).ElementR("button", "Спроба тесту")
		if err != nil {
			button, err = page.MustWaitLoad().Timeout(time.Second).ElementR("button", "Зробити наступну спробу")
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
				err := storage.PutQA(testNum, d.Question, d.Rightanswer)
				if err != nil {
					log.Fatalln("Cant insert data to storage:", err)
				}
			}
		}
	}
}
