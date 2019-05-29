package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "casa",
	Short: "Casa is a domotic software",
	Long:  "Casa is a domotic software built to connect all of smart things in one place.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Type `casa help` to view commands available")
	},
}

// Execute cli init
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
