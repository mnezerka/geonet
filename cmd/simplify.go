package cmd

import (
	"mnezerka/geonet/config"
	"mnezerka/geonet/log"
	"mnezerka/geonet/store"

	"github.com/spf13/cobra"
)

var simplifyCmd = &cobra.Command{
	Use:   "simplify",
	Short: "Simplify geonet in directly in database",
	RunE: func(cmd *cobra.Command, args []string) error {
		store := store.NewMongoStore(&config.Cfg)
		defer func() { store.Close() }()
		store.Simplify()
		log.Info("database successfully simplified")

		return nil
	},
}

func init() {
	rootCmd.AddCommand(simplifyCmd)
}
