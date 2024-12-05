package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tkrajina/gpxgo/gpx"
	"mnezerka/geonet/log"
	"mnezerka/geonet/store"
)

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate map",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		store := store.NewMongoStore()

		defer func() { store.Close() }()

		// clean all data from previous executions
		store.Reset()

		log.Infof("generating map from %v", args)

		for file_ix := 0; file_ix < len(args); file_ix++ {

			log.Infof("processing %s", args[file_ix])

			gpxFile, err := gpx.ParseFile(args[file_ix])
			if err != nil {
				return err
			}

			err = store.AddGpx(gpxFile, args[file_ix])
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
