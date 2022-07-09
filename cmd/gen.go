package cmd

import (
	"fmt"
	"log"
	"os"
	"text/template"

	"github.com/ptrv/go-gpx"
	"github.com/spf13/cobra"
)

var genCmd = &cobra.Command{
	Use:   "gen",
	Short: "Generate map",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Do Stuff Here
		fmt.Println("generating map from %v", args)

		var js []string

		for file_ix := 0; file_ix < len(args); file_ix++ {

			fmt.Printf("Processing file %d: %s\n", file_ix, args[file_ix])

			gpx, err := gpx.ParseFile(args[file_ix])
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed: %s\n", err)
				return err
			}

			fmt.Printf("  length 2D: %f\n", gpx.Length2D())
			fmt.Printf("  tracks: %d\n", len(gpx.Tracks))

			js = append(js, GpxToPolyline(gpx))
		}

		t, err := template.ParseFiles("cmd/tpl_preview.html")
		if err != nil {
			log.Print(err)
			return err
		}

		f, err := os.Create("preview.html")
		if err != nil {
			log.Println("create file: ", err)
			return err
		}

		config := js
		err = t.Execute(f, config)
		if err != nil {
			log.Print("execute: ", err)
			return err
		}

		return nil
	},
}
