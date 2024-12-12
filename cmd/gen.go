package cmd

import (
	"mnezerka/geonet/config"
	"mnezerka/geonet/log"
	"mnezerka/geonet/store"

	"github.com/mnezerka/gpxcli/gpxutils"
	"github.com/spf13/cobra"
	"github.com/tkrajina/gpxgo/gpx"
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

			gpxFile, err := gpx.ParseFile(args[file_ix])
			if err != nil {
				return err
			}

			points, err := gpxutils.GpxFileToPoints(gpxFile)
			if err != nil {
				return err
			}

			log.Infof("points read from gpx file: %d", len(points))

			if genCmdInterpolate {
				points, err = gpxutils.InterpolateDistance(points, float64(config.Cfg.InterpolationDistance))
				if err != nil {
					return err
				}

				log.Infof("points interpolated: %d", len(points))
			}

			err = store.AddGpx(points, args[file_ix])
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
