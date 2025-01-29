package cmd

import (
	"mnezerka/geonet/config"
	"mnezerka/geonet/log"
	"mnezerka/geonet/store"
	"mnezerka/geonet/tracks"

	"github.com/spf13/cobra"
)

var genCmdInterpolate bool

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate map",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		store := store.NewMongoStore(&config.Cfg)
		defer func() { store.Close() }()

		// clean all data from previous executions
		store.Reset()

		log.Infof("generating geonet from %v", args)

		for file_ix := 0; file_ix < len(args); file_ix++ {

			log.Infof("processing %s", args[file_ix])

			t := tracks.NewTrack(args[file_ix])

			if genCmdInterpolate {
				t.InterpolateDistance(config.Cfg.InterpolationDistance)
			}

			err := store.AddGpx(t)
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	genCmd.PersistentFlags().BoolVarP(&genCmdInterpolate, "interpolate", "i", false, "interpolate tracks before adding to geonet")

	rootCmd.AddCommand(genCmd)
}
