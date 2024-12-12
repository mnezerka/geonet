package cmd

import (
	//"fmt"
	"encoding/json"
	"os"
	"text/template"

	"mnezerka/geonet/config"
	"mnezerka/geonet/log"
	"mnezerka/geonet/store"

	"github.com/spf13/cobra"
)

type mapData struct {
	Title          string
	Meta           string
	GeoJson        string
	UseTrackColors bool
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Visualise geonet",
	RunE: func(cmd *cobra.Command, args []string) error {

		store := store.NewMongoStore(&config.Cfg)
		defer func() { store.Close() }()

		log.Info("visualisation of geonet")

		collection, err := store.ToGeoJson()
		if err != nil {
			return err
		}

		meta, err := store.GetMeta()
		if err != nil {
			return err
		}

		// meta - json
		metaJson, err := json.Marshal(meta)
		if err != nil {
			return err
		}

		// collection -> json
		rawJSON, err := collection.MarshalJSON()
		if err != nil {
			return err
		}

		var tmplFile = "templates/map.html"

		// Load the template file
		tmpl, err := template.ParseFiles(tmplFile)
		if err != nil {
			return err
		}

		data := mapData{
			Title:          "GeoNet",
			Meta:           string(metaJson),
			GeoJson:        string(rawJSON),
			UseTrackColors: config.Cfg.ShowTrackColors,
		}

		err = tmpl.Execute(os.Stdout, data)
		if err != nil {
			return (err)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
