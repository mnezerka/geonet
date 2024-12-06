package cmd

import (
	"github.com/spf13/cobra"
	"mnezerka/geonet/log"
	"mnezerka/geonet/store"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset database - delete all data (points, tracks, edges)",
	RunE: func(cmd *cobra.Command, args []string) error {

		store := store.NewMongoStore()
		store.Reset()
		log.Info("database reset passed successfully")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(resetCmd)
}
