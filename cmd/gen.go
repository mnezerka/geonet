package cmd

import (
	"os"
	"text/template"

	"mnezerka/geonet/auxiliary"
	"mnezerka/geonet/store"

	"github.com/apex/log"

	"github.com/ptrv/go-gpx"
	"github.com/spf13/cobra"
)

var genFlags CommonParams

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate map",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Do Stuff Here
		auxiliary.Infof("generating map from %v", args)

		store := store.NewStore()
		store.LinesEnabled = !genFlags.SpotsOnly

		var js []string

		for file_ix := 0; file_ix < len(args); file_ix++ {

			// get unique id of the gpx file (computed from file content)
			hash, err := getFileHash(args[file_ix])
			if err != nil {
				return err
			}

			gpx, err := gpx.ParseFile(args[file_ix])
			if err != nil {
				log.WithError(err)
				return err
			}

			store.AddGpx(gpx, genFlags.Treshold, hash)

			// visualize add polyline
			if genFlags.ShowGpxLines {
				js = append(js, GpxToPolyline(gpx, file_ix+1, 1))
			}
		}

		// visualize hulls
		if genFlags.ShowHulls {
			js = append(js, StoreHullsToPolygons(store, 1))
		}

		// visualize lines
		if genFlags.ShowLines {
			js = append(js, StoreLinesToPolylines(store, 1))
		}

		// visualize hulls as circles with variable diameter
		if genFlags.ShowHeatSpots {
			js = append(js, StoreHullsToHeatCircles(store, 1))
		}

		t, err := template.ParseFiles("cmd/tpl_preview.html")
		if err != nil {
			log.WithError(err)
			return err
		}

		f, err := os.Create("preview.html")
		if err != nil {
			log.WithError(err)
			return err
		}

		config := js
		err = t.Execute(f, config)
		if err != nil {
			log.WithError(err)
			return err
		}

		return nil
	},
}

func init() {
	addCommonFlags(&genFlags, genCmd)

	rootCmd.AddCommand(genCmd)
}
