package cmd

import (
	"encoding/json"
	"fmt"
	"mnezerka/geonet/config"
	"mnezerka/geonet/log"
	"mnezerka/geonet/s2store"
	"mnezerka/geonet/store"
	"os"
	"text/template"

	"github.com/spf13/cobra"
)

type mapData struct {
	Title          string
	Meta           string
	GeoJson        string
	UseTrackColors bool
	Js             string
}

var processingRender bool
var processingSimplify bool
var processingExport bool
var processingExportFormat string
var processingSave bool

func addProcessingFlags(cmd *cobra.Command) {

	// rendering
	cmd.PersistentFlags().BoolVar(&processingRender, "render", false, "render final geonet to html page")

	// export
	cmd.PersistentFlags().BoolVarP(&processingExport, "export", "e", false, "export final geonet to geojson")
	cmd.PersistentFlags().StringVar(&processingExportFormat, "export-format", "json", "export format (json, geojson, svg)")

	// rendering + export
	cmd.PersistentFlags().BoolVar(&config.Cfg.ShowPoints, "points", config.Cfg.ShowPoints, "render individual points")
	cmd.PersistentFlags().BoolVar(&config.Cfg.ShowEdges, "edges", config.Cfg.ShowEdges, "render edges")

	// simplification
	cmd.PersistentFlags().BoolVar(&processingSimplify, "simplify", false, "simplify geonet")
	cmd.PersistentFlags().Int64Var(&config.Cfg.SimplifyMinDistance, "sim-min-dist", config.Cfg.MatchMaxDistance, "minimal distance between points for simplification")

	// save
	cmd.PersistentFlags().BoolVar(&processingSave, "save", false, "save geonet (json format)")
}

func processing(s2store *s2store.S2Store) {

	if processingSimplify {
		log.Infof("simplifying network")
		s2store.Simplify()
	}

	if processingExport {
		log.Infof("exporting network")

		switch processingExportFormat {
		case "svg":
			fmt.Print(store.ExportSvg(s2store))
			break
		case "geojson":
			fmt.Print(string(store.ExportGeoJson(s2store)))
			break
		default:
			fmt.Print(string(store.Export(s2store)))
		}
	}

	if processingRender {
		render(s2store)
	}

	if processingSave {
		s2store.Save()
	}
}

func render(store *s2store.S2Store) {
	log.Debug("------------ rendering geonet to html --------------")

	collection := store.ToGeoJson(nil)

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
