package cmd

import (
	"github.com/ItsJimi/casa/server"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(initCmd)
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Init casa database",
	Long:  "Init casa database.",
	Run: func(cmd *cobra.Command, args []string) {
		server.InitDB()
	},
}
