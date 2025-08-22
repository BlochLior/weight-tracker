package tracker

import (
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "weight-tracker",
	Short: "weight-tracker is a tool to track your weight",
	Long: `weight-tracker is a tool to track your weight.
It can be used to track your weight over time and to see how much you have lost or gained.
The tool can also output graphs of your weight over time.`,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(graphCmd)
}
