package cmd

import (
	"mnezerka/geonet/config"
	"mnezerka/geonet/log"
	"mnezerka/geonet/s2store"
	"mnezerka/geonet/tracks"

	"github.com/spf13/cobra"
)

var cmdGenInterpolate bool
var cmdGenLimit int
var cmdGenLoadPath string

var cmdNet = &cobra.Command{
	Use:   "net [flags] [gpx files]",
	Short: "Geographic network toolset",
	RunE: func(cmd *cobra.Command, args []string) error {

		store := s2store.NewS2Store(&config.Cfg)

		if len(cmdGenLoadPath) > 0 {
			log.Infof("loading geonet from %s", cmdGenLoadPath)
			store.Load(cmdGenLoadPath)
			log.Infof("loaded")
		}

		if len(args) > 0 {

			log.Infof("generating geonet from %d input files", len(args))

			log.Infof("building network")
			for file_ix := 0; file_ix < len(args); file_ix++ {

				// interrupt loop if number of tracks is limited
				if cmdGenLimit >= 0 {
					if file_ix >= cmdGenLimit {
						break
					}
				}

				t := tracks.NewTrack(args[file_ix])
				log.Infof("  %d/%d %s (%d points)", file_ix+1, len(args), args[file_ix], len(t.Points))

				if cmdGenInterpolate {
					t.InterpolateDistance(config.Cfg.InterpolationDistance)
				}

				err := store.AddGpx(t)
				if err != nil {
					return err
				}
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
	cmdNet.PersistentFlags().BoolVarP(&cmdGenInterpolate, "interpolate", "i", false, "interpolate tracks before adding to geonet")
	cmdNet.PersistentFlags().Int64Var(&config.Cfg.InterpolationDistance, "int-dist", config.Cfg.InterpolationDistance, "distance for interpolation in meters")
	cmdNet.PersistentFlags().Int64Var(&config.Cfg.MatchMaxDistance, "match-max-dist", config.Cfg.MatchMaxDistance, "maximal distance in meters for matching new points against points in geonet")
	cmdNet.PersistentFlags().IntVar(&cmdGenLimit, "limit", -1, "max number of tracks to be processed")

	cmdNet.PersistentFlags().StringVar(&cmdGenLoadPath, "load", "", "load geo network from file before processing")

	addProcessingFlags(cmdNet)

	addExportFlags(cmdNet)

	rootCmd.AddCommand(cmdNet)
}
