package cmd

import (
	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset database - delete all data (points, tracks, edges)",
	RunE: func(cmd *cobra.Command, args []string) error {

		/*
			store := store.NewMongoStore(&config.Cfg)
			defer func() { store.Close() }()

			store.Reset()
			log.Info("database reset passed successfully")
		*/

		return nil
	},
}

func init() {
	rootCmd.AddCommand(resetCmd)
}
