package cmd

import (
	"github.com/mnezerka/gpxcli/gpxutils"
	"github.com/spf13/cobra"
	"github.com/tkrajina/gpxgo/gpx"
	"mnezerka/geonet/log"
	"mnezerka/geonet/store"
)

const INTERPOLATION_DISTANCE = 30 // in meters

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate map",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		store := store.NewMongoStore()

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

			log.Debugf("points raw: %d", len(points))
			pointsInterpolated, err := gpxutils.InterpolateDistance(points, INTERPOLATION_DISTANCE)
			if err != nil {
				return err
			}

			log.Debugf("points interpolated: %d", len(pointsInterpolated))

			err = store.AddGpx(pointsInterpolated, args[file_ix])
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(genCmd)
}
