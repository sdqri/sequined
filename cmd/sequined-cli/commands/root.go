package commands

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "sequined",
	Short: "A website mocking tool to test your spider's performance against dynamic and realistic environments",
	Long: "Sequined is a website mocking tool that aims to aid in exploring synchronization problems in crawlers. " +
		"By simulating changes in website structure and data patterns, it helps you measure the freshness and age of web data, " +
		"facilitating the refinement of data refreshing methods and meeting performance requirements.",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
