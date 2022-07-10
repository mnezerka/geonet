package cmd

import (
	"os"
	"text/template"

	"mnezerka/geonet/store"

	"github.com/apex/log"

	"github.com/ptrv/go-gpx"
	"github.com/spf13/cobra"
)

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate map",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Do Stuff Here
		log.Infof("generating map from %v", args)

		store := store.NewStore()

		var js []string

		for file_ix := 0; file_ix < len(args); file_ix++ {

			gpx, err := gpx.ParseFile(args[file_ix])
			if err != nil {
				log.WithError(err)
				return err
			}

			//fmt.Printf("  length 2D: %f\n", gpx.Length2D())
			//fmt.Printf("  tracks: %d\n", len(gpx.Tracks))

			store.AddGpx(gpx, 0.0005)

			// visualize add polyline
			js = append(js, GpxToPolyline(gpx))

			// visualize hulls
			js = append(js, StoreHullsToPolygons(store))
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
