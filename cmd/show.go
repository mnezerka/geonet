package cmd

import (
	//"fmt"
	"os"
	"text/template"

	"github.com/spf13/cobra"
	"mnezerka/geonet/log"
	"mnezerka/geonet/store"
)

type mapData struct {
	Title   string
	GeoJson string
}

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Visualise geonet",
	RunE: func(cmd *cobra.Command, args []string) error {

		store := store.NewMongoStore()

		defer func() { store.Close() }()

		log.Info("visualisation of geonet")

		collection, err := store.ToGeoJson()
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
			Title:   "GeoNet",
			GeoJson: string(rawJSON),
		}

		err = tmpl.Execute(os.Stdout, data)
		if err != nil {
			return (err)
		}

		//fmt.Printf("%s", string(rawJSON))

		/*

			  tpl_path = 'templates/tpl_map_grid.html'
			print(f'generating html content from template {tpl_path}')
			with open(tpl_path, 'r') as tpl_file:
			    tpl = string.Template(tpl_file.read())

			html_content = tpl.substitute({
			    'title': 'Net',
			    'geojson': geojson_content,
			    'meta': n.get_meta()
			})

			html_path = f'{output}.html'
			print(f'writing html to {html_path}')
			with open(html_path, 'w') as html_file:
			    html_file.write(html_content)

		*/

		return nil
	},
}

func init() {
	rootCmd.AddCommand(showCmd)
}
