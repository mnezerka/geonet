package cmd

import (
	"github.com/spf13/cobra"
	"mnezerka/geonet/log"
	"mnezerka/geonet/store"
)

var simplifyCmd = &cobra.Command{
	Use:   "simplify",
	Short: "Simplify geonet in database)",
	RunE: func(cmd *cobra.Command, args []string) error {

		store := store.NewMongoStore()
		store.Simplify()
		log.Info("database successfully simplified")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(simplifyCmd)
}
