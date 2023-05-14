package cmd

import (
	"log"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/pryamcem/vns-scraper/config"
	"github.com/pryamcem/vns-scraper/scruper"
	"github.com/pryamcem/vns-scraper/storage"
	"github.com/spf13/cobra"
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
	l := launcher.New().
		Headless(false).
		Devtools(true)

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
	err = storage.CreateTableByNum(testNum)
	if err != nil {
		log.Fatalf("Cant't create table schema: %v", err)
	}
}
