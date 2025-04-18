package cmd

import (
	"mnezerka/geonet/config"
	"mnezerka/geonet/log"
	"mnezerka/geonet/s2store"

	"github.com/spf13/cobra"
)

var processingSimplify bool
var processingSave bool

func addProcessingFlags(cmd *cobra.Command) {

	// simplification
	cmd.PersistentFlags().BoolVar(&processingSimplify, "simplify", false, "simplify geonet")
	cmd.PersistentFlags().Int64Var(&config.Cfg.SimplifyMinDistance, "sim-min-dist", config.Cfg.MatchMaxDistance, "minimal distance between points for simplification")

	// save
	cmd.PersistentFlags().BoolVar(&processingSave, "save", false, "save geonet (json format)")
}

func processing(s2store *s2store.S2Store) {

	if processingSimplify {
		log.Infof("simplifying network")
		s2store.Simplify()
	}

	if processingSave {
		s2store.Save()
	}
}
