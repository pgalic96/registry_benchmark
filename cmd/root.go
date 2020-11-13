package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var writeToCSV bool
var YamlFilename string
var cronJob bool

var rootCmd = &cobra.Command{
	Use:   "benchmarkd",
	Short: "Docker registry benchmark",
	Long:  `Registry benchmark created by pgalic96, intended to measure different metrics related to container registry performance`,
}

// Execute executes
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&YamlFilename, "yaml", "y", "config.yaml", "benchmark config file name")
	rootCmd.PersistentFlags().BoolVarP(&writeToCSV, "csv", "c", false, "write to local csv file")
	rootCmd.PersistentFlags().BoolVarP(&cronJob, "long-running", "l", false, "Flag that determines whether the experiment is long running or not")
}
