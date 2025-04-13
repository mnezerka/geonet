package cmd

import (
	"mnezerka/geonet/config"

	"github.com/spf13/cobra"
)

var simplifyCmd = &cobra.Command{
	Use:   "simplify",
	Short: "Simplify geonet in directly in database",
	RunE: func(cmd *cobra.Command, args []string) error {
		/*
			store := store.NewMongoStore(&config.Cfg)
			defer func() { store.Close() }()
			store.Simplify()
			log.Info("database successfully simplified")
		*/

		return nil
	},
}

func init() {
	simplifyCmd.PersistentFlags().Int64Var(&config.Cfg.SimplifyMinDistance, "sim-min-dist", config.Cfg.MatchMaxDistance, "minimal distance between points for simplification")
	rootCmd.AddCommand(simplifyCmd)
}
