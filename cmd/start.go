package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start casa server",
	Long:  "Start casa server.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Start")
	},
}
