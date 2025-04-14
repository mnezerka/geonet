package cmd

import (
	"mnezerka/geonet/config"
	"mnezerka/geonet/log"
	"mnezerka/geonet/s2store"

	"github.com/spf13/cobra"
)

var loadCmd = &cobra.Command{
	Use:   "load FILEPATH",
	Short: "Load complete geonet from json file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {

		store := s2store.NewS2Store(&config.Cfg)

		log.Infof("loading geonet from %s", args[0])

		store.Load(args[0])

		processing(store)

		log.Infof("statistics:")
		store.GetStat().Print()
	},
}

func init() {
	addProcessingFlags(loadCmd)
	rootCmd.AddCommand(loadCmd)
}
