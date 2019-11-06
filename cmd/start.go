package cmd

import (
	"github.com/ItsJimi/casa/server"
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
		port := "4353"
		if len(args) > 0 {
			port = args[0]
		}
		server.StartDB()
		go server.Automations()
		server.Start(port)
	},
}
