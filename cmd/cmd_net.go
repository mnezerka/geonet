package cmd

import (
	"fmt"
	"os"

	"mnezerka/geonet/config"
	"mnezerka/geonet/log"
	"mnezerka/geonet/s2store"
	"mnezerka/geonet/tracks"

	"github.com/schollz/progressbar/v3"
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

			limit := len(args)
			if cmdGenLimit >= 0 && cmdGenLimit < limit {
				limit = cmdGenLimit
			}

			var bar *progressbar.ProgressBar
			if !v {
				bar = progressbar.NewOptions(limit,
					progressbar.OptionSetWriter(os.Stderr),
					progressbar.OptionSetDescription(fmt.Sprintf("[%d/%d]", 0, limit)),
					progressbar.OptionSetWidth(40),
					progressbar.OptionShowIts(),
					progressbar.OptionSetItsString("files"),
					progressbar.OptionSetTheme(progressbar.Theme{
						Saucer:        "=",
						SaucerHead:    ">",
						SaucerPadding: " ",
						BarStart:      "[",
						BarEnd:        "]",
					}),
					progressbar.OptionOnCompletion(func() { fmt.Fprint(os.Stderr, "\n") }),
				)
			} else {
				log.Infof("building network")
			}

			for file_ix := 0; file_ix < limit; file_ix++ {
				t := tracks.NewTrack(args[file_ix])

				if v {
					log.Infof("  %d/%d %s (%d points)", file_ix+1, limit, args[file_ix], len(t.Points))
				} else {
					bar.Describe(fmt.Sprintf("[%d/%d]", file_ix+1, limit))
				}

				if cmdGenInterpolate {
					t.InterpolateDistance(config.Cfg.InterpolationDistance)
				}

				err := store.AddGpx(t)
				if err != nil {
					return err
				}

				if bar != nil {
					bar.Add(1)
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
