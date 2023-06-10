package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Short: "Simple VNS scraper and test completer.",
}

// TODO: provide better error handling.
func Execute() {
	rootCmd.AddCommand(passCmd, saveCmd, scanCmd)
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
