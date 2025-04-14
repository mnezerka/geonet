package cmd

import (
	"encoding/json"
	"fmt"
	"mnezerka/geonet/config"
	"mnezerka/geonet/log"
	"mnezerka/geonet/s2store"
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
var processingSave bool

func addProcessingFlags(cmd *cobra.Command) {

	// rendering
	cmd.PersistentFlags().BoolVar(&processingRender, "render", false, "render final geonet to html page")

	// export
	cmd.PersistentFlags().BoolVarP(&processingExport, "export", "e", false, "export final geonet to geojson")

	// rendering + export
	cmd.PersistentFlags().BoolVar(&config.Cfg.ShowPoints, "points", config.Cfg.ShowPoints, "render individual points")
	cmd.PersistentFlags().BoolVar(&config.Cfg.ShowEdges, "edges", config.Cfg.ShowEdges, "render edges")

	// simplification
	cmd.PersistentFlags().BoolVar(&processingSimplify, "simplify", false, "simplify geonet")
	cmd.PersistentFlags().Int64Var(&config.Cfg.SimplifyMinDistance, "sim-min-dist", config.Cfg.MatchMaxDistance, "minimal distance between points for simplification")

	// save
	cmd.PersistentFlags().BoolVar(&processingSave, "save", false, "save geonet (json format)")
}

func processing(store *s2store.S2Store) {

	if processingSimplify {
		log.Infof("simplifying network")
		store.Simplify()
	}

	if processingExport {
		log.Infof("exporting network")
		gJson := store.ToGeoJson()

		strJson, err := json.MarshalIndent(gJson, "", " ")
		if err != nil {
			log.ExitWithError(err)
		}

		fmt.Print(string(strJson))
	}

	if processingRender {
		render(store)
	}

	if processingSave {
		store.Save()
	}
}

func render(store *s2store.S2Store) {
	log.Debug("------------ rendering geonet to html --------------")

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
