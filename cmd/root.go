package cmd

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Short: "Simple VNS scruper and test completer.",
}

// TODO: provide better error handling.
func Execute() {
	rootCmd.AddCommand(passCmd)
	rootCmd.AddCommand(saveCmd)
	err := rootCmd.Execute()
	if err != nil {
		panic(err)
	}
}
