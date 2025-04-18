package cmd

import (
	"embed"
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

var flagExport bool
var flagExportFormat string

var templatesContent *embed.FS

func addExportFlags(cmd *cobra.Command) {

	// export
	cmd.PersistentFlags().BoolVarP(&flagExport, "export", "e", false, "export content to various formats")
	cmd.PersistentFlags().StringVar(&flagExportFormat, "export-format", "json", "export format (json, geojson, svg, txt)")
	cmd.PersistentFlags().BoolVar(&config.Cfg.ShowPoints, "points", config.Cfg.ShowPoints, "include points in exported content ")
	cmd.PersistentFlags().BoolVar(&config.Cfg.ShowEdges, "edges", config.Cfg.ShowEdges, "include edges in exported content")
	addExportSvgFlags(cmd)
}

func addExportSvgFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().IntVar(&config.Cfg.SvgWidth, "svg-width", config.Cfg.SvgWidth, "width of generated svg in pixels")
	cmd.PersistentFlags().IntVar(&config.Cfg.SvgHeight, "svg-height", config.Cfg.SvgHeight, "height of generated svg in pixels")
	cmd.PersistentFlags().IntVar(&config.Cfg.SvgPadding, "svg-padding", config.Cfg.SvgPadding, "padding of generated svg in pixels")
	cmd.PersistentFlags().BoolVar(&config.Cfg.SvgPointLabels, "svg-point-labels", config.Cfg.SvgPointLabels, "add label (e.g. id) to each point")
	cmd.PersistentFlags().BoolVar(&config.Cfg.SvgEdgeLabels, "svg-edge-labels", config.Cfg.SvgEdgeLabels, "add label (e.g. tracks) to each edge")
}

func export(s2store *s2store.S2Store) {

	if flagExport {
		log.Infof("exporting network")

		switch flagExportFormat {
		case "svg":
			fmt.Print(store.ExportSvg(s2store))
			break
		case "geojson":
			fmt.Print(string(store.ExportGeoJson(s2store)))
			break
		case "txt":
			fmt.Print(s2store.ToTxt())
			break
		case "html":
			render(s2store)
			break
		default:
			fmt.Print(string(store.Export(s2store)))
		}
	}
}

func SetTemplatesContent(content *embed.FS) {
	templatesContent = content
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

	// Load the template file
	htmlContent, err := template.ParseFS(templatesContent, "templates/map.html")
	if err != nil {
		log.ExitWithError(err)
	}

	// load js logic
	jsContent, err := templatesContent.ReadFile("templates/map.js")
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

	err = htmlContent.Execute(os.Stdout, data)
	if err != nil {
		log.ExitWithError(err)
	}
}
