package cmd

import (
	"encoding/json"
	"fmt"
	"mnezerka/geonet/config"
	"mnezerka/geonet/log"
	"mnezerka/geonet/s2store"
	"mnezerka/geonet/store"
	"mnezerka/geonet/tracks"
	"os"
	"text/template"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
)

var genCmdSimplify bool
var genCmdRender bool
var genCmdExport bool
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

		if genCmdSimplify {
			log.Infof("simplifying network")
			store.Simplify()
		}

		if genCmdExport {
			log.Infof("exporting network")
			gJson := store.ToGeoJson()

			strJson, err := json.MarshalIndent(gJson, "", " ")
			if err != nil {
				return err
			}

			fmt.Print(string(strJson))
		}

		if genCmdRender {
			log.Infof("rendering network")
			render(store)
		}

		log.Infof("statistics:")
		printStoreStat(store.Stat)

		return nil
	},
}

func init() {
	genCmd.PersistentFlags().BoolVarP(&genCmdInterpolate, "interpolate", "i", false, "interpolate tracks before adding to geonet")
	genCmd.PersistentFlags().Int64Var(&config.Cfg.InterpolationDistance, "int-dist", config.Cfg.InterpolationDistance, "distance for interpolation in meters")
	genCmd.PersistentFlags().Int64Var(&config.Cfg.MatchMaxDistance, "match-max-dist", config.Cfg.MatchMaxDistance, "maximal distance in meters for matching new points against points in geonet")
	genCmd.PersistentFlags().IntVar(&genCmdLimit, "limit", -1, "max number of tracks to be processed")

	genCmd.PersistentFlags().BoolVar(&genCmdRender, "render", false, "render final geonet to html page")
	genCmd.PersistentFlags().BoolVarP(&genCmdExport, "export", "e", false, "export final geonet to geojson")
	genCmd.PersistentFlags().BoolVar(&config.Cfg.ShowPoints, "points", config.Cfg.ShowPoints, "render individual points")
	genCmd.PersistentFlags().BoolVar(&config.Cfg.ShowEdges, "edges", config.Cfg.ShowEdges, "render edges")

	genCmd.PersistentFlags().BoolVar(&genCmdSimplify, "simplify", false, "simplify geonet")
	genCmd.PersistentFlags().Int64Var(&config.Cfg.SimplifyMinDistance, "sim-min-dist", config.Cfg.MatchMaxDistance, "minimal distance between points for simplification")

	rootCmd.AddCommand(genCmd)
}

func printStoreStat(stat store.Stat) {

	t := table.NewWriter()
	t.SetOutputMirror(os.Stderr)
	t.AppendHeader(table.Row{"Entity", "Processed", "Reused", "Created", "Simplified", "Final", "Rendered"})
	t.AppendRows([]table.Row{
		{"tracks", stat.TracksProcessed, "-", "-", "-", stat.TracksProcessed, "-"},
		{"points", stat.PointsProcessed, stat.PointsReused, stat.PointsCreated, stat.PointsSimplified, stat.PointsFinal, stat.PointsRendered},
		{"edges", "-", stat.EdgesReused, stat.EdgesCreated, stat.EdgesSimplified, stat.EdgesFinal, stat.EdgesRendered},
		{"segments", stat.SegmentsProcessed, "-", "-", stat.SegmentsSimplified, "-", stat.SegmentsRendered},
	})
	t.Render()
}

func render(store *s2store.S2Store) {
	log.Debug("------------ visualisation of geonet --------------")

	collection := store.ToGeoJson()

	meta := store.GetMeta()

	// meta - json
	metaJson, err := json.Marshal(meta)
	if err != nil {
		log.ExitWithError(err)
	}

	// collection -> json
	rawJSON, err := collection.MarshalJSON()
	if err != nil {
		log.ExitWithError(err)
	}

	var tmplFile = "templates/map.html"

	// Load the template file
	tmpl, err := template.ParseFiles(tmplFile)
	if err != nil {
		log.ExitWithError(err)
	}

	// load js logic
	jsContent, err := os.ReadFile("templates/map.js")
	if err != nil {
		log.ExitWithError(err)
	}

	data := mapData{
		Title:          "GeoNet",
		Meta:           string(metaJson),
		GeoJson:        string(rawJSON),
		UseTrackColors: config.Cfg.ShowTrackColors,
		Js:             string(jsContent),
	}

	err = tmpl.Execute(os.Stdout, data)
	if err != nil {
		log.ExitWithError(err)
	}
}
