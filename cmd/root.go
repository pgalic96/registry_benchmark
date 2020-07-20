package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

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
