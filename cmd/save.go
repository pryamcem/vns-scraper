package cmd

import (
	"log"
	"strconv"

	"github.com/pryamcem/vns-scraper/storage"
	"github.com/spf13/cobra"
)

var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save answers to file.",
	Run:   save,
}

func save(_ *cobra.Command, args []string) {
	testNum, err := strconv.Atoi(args[0])
	if err != nil {
		log.Fatalln("Not enough or wrong arguments.")
	}

	storage, err := storage.New("tests.db")
	if err != nil {
		log.Fatalf("Storage initialization error: %v", err)
	}
	defer storage.Close()
	storage.ParseToFile(testNum)
}
