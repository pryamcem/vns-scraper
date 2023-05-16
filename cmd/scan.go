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
	linksTitle = "Переглянути ваші відповіді в цій спробі"
)

var scanCmd = &cobra.Command{
	Use:   "scan",
	Short: "Scan already completed test.",
	Run:   scan,
}

func scan(_ *cobra.Command, args []string) {
	link := args[0]
	config := config.GetConfig(configPath)

	// Init storage.
	storage, err := storage.New("tests.db")
	if err != nil {
		log.Fatalf("Storage initialization error: %v", err)
	}
	defer storage.Close()

	// Create new browser.
	l := launcher.New().Headless(false).Devtools(true)

	defer l.Cleanup()
	url := l.MustLaunch()

	browser := rod.New().ControlURL(url).Trace(true).MustConnect()
	defer browser.Close()

	page := browser.MustPage(link)

	err = scruper.Login(page, config)
	if err != nil {
		log.Fatalln("Can't login: ", err)
	}

	testNum := scruper.MustFindTestNum(page)
	log.Println(testNum)

	err = storage.CreateTableByNum(testNum)
	if err != nil {
		log.Fatalf("Cant't create table schema: %v", err)
	}

	//urls, err := scruper.GetLinksByTitle(page, linksTitle)
	urls := scruper.FindTests(page)
	if err != nil {
		log.Fatalf("Can't get number of urls: %v", err)
	}
	log.Println(urls)

	for _, url := range urls {
		err := page.Navigate(url)
		if err != nil {
			log.Fatalf("Can't navigate to results link: %v", err)
		}
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
