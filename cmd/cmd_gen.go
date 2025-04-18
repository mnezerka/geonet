package cmd

import (
	"mnezerka/geonet/config"
	"mnezerka/geonet/log"
	"mnezerka/geonet/s2store"
	"mnezerka/geonet/tracks"

	"github.com/spf13/cobra"
)

var genCmdInterpolate bool
var genCmdLimit int

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate map",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		store := s2store.NewS2Store(&config.Cfg)

		log.Infof("generating geonet from %d input files", len(args))

		log.Infof("building network")
		for file_ix := 0; file_ix < len(args); file_ix++ {

			// interrupt loop if number of tracks is limited
			if genCmdLimit >= 0 {
				if file_ix >= genCmdLimit {
					break
				}
			}

			t := tracks.NewTrack(args[file_ix])
			log.Infof("  %d/%d %s (%d points)", file_ix+1, len(args), args[file_ix], len(t.Points))

			if genCmdInterpolate {
				t.InterpolateDistance(config.Cfg.InterpolationDistance)
			}

			err := store.AddGpx(t)
			if err != nil {
				return err
			}
		}

		processing(store)

		export(store)

		log.Infof("statistics:")
		store.GetStat().Print()

		return nil
	},
}

func init() {
	genCmd.PersistentFlags().BoolVarP(&genCmdInterpolate, "interpolate", "i", false, "interpolate tracks before adding to geonet")
	genCmd.PersistentFlags().Int64Var(&config.Cfg.InterpolationDistance, "int-dist", config.Cfg.InterpolationDistance, "distance for interpolation in meters")
	genCmd.PersistentFlags().Int64Var(&config.Cfg.MatchMaxDistance, "match-max-dist", config.Cfg.MatchMaxDistance, "maximal distance in meters for matching new points against points in geonet")
	genCmd.PersistentFlags().IntVar(&genCmdLimit, "limit", -1, "max number of tracks to be processed")

	addProcessingFlags(genCmd)

	addExportFlags(genCmd)

	rootCmd.AddCommand(genCmd)
}
