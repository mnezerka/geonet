package cmd

import (
	"encoding/json"
	"fmt"
	"mnezerka/geonet/config"
	"mnezerka/geonet/log"
	"mnezerka/geonet/s2store"
	"mnezerka/geonet/tracks"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var genCmdExport bool
var genCmdInterpolate bool
var genCmdLimit int

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate map",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		//store := store.NewMongoStore(&config.Cfg)

		store := s2store.NewS2Store(&config.Cfg)

		//defer func() { store.Close() }()

		// clean all data from previous executions
		//store.Reset()

		log.Infof("generating geonet from %v", args)

		/*
			store.Log(fmt.Sprintf("generate cfg=(%s) flags=(%s)",
				config.Cfg.ToString(),
				fmt.Sprintf("interpolate: %v, limit: %d", genCmdInterpolate, genCmdLimit),
			))
		*/

		for file_ix := 0; file_ix < len(args); file_ix++ {

			// interrupt loop if number of tracks is limited
			if genCmdLimit >= 0 {
				if file_ix >= genCmdLimit {
					break
				}
			}

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

		if genCmdExport {
			gJson, err := store.ToGeoJson()
			if err != nil {
				return err
			}

			strJson, err := json.MarshalIndent(gJson, "", " ")
			if err != nil {
				return err
			}

			fmt.Print(string(strJson))
		}

		printStoreStat(store.Stat)

		return nil
	},
}

func init() {
	genCmd.PersistentFlags().BoolVarP(&genCmdInterpolate, "interpolate", "i", false, "interpolate tracks before adding to geonet")
	genCmd.PersistentFlags().Int64Var(&config.Cfg.InterpolationDistance, "int-dist", config.Cfg.InterpolationDistance, "distance for interpolation in meters")
	genCmd.PersistentFlags().Int64Var(&config.Cfg.MatchMaxDistance, "match-max-dist", config.Cfg.MatchMaxDistance, "maximal distance in meters for matching new points against points in geonet")
	genCmd.PersistentFlags().IntVar(&genCmdLimit, "limit", -1, "max number of tracks to be processed")

	genCmd.PersistentFlags().BoolVarP(&genCmdExport, "export", "e", false, "export final geonet to geojson")
	genCmd.PersistentFlags().BoolVar(&config.Cfg.ShowPoints, "points", config.Cfg.ShowPoints, "render individual points")
	genCmd.PersistentFlags().BoolVar(&config.Cfg.ShowEdges, "edges", config.Cfg.ShowEdges, "render edges")

	rootCmd.AddCommand(genCmd)
}

func printStoreStat(stat s2store.S3StoreStat) {

	t := table.NewWriter()
	t.SetOutputMirror(os.Stderr)
	t.AppendHeader(table.Row{"Entity", "Processed", "Reused", "Created"})
	t.AppendRows([]table.Row{
		{"tracks", stat.TracksProcessed, "-", "-"},
		{"points", stat.PointsProcessed, stat.PointsReused, stat.PointsCreated},
		{"edges", "-", stat.EdgesReused, stat.EdgesCreated},
	})
	t.Render()

}
