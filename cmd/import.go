package cmd

import (
	"encoding/json"
	"mnezerka/geonet/config"
	"mnezerka/geonet/log"
	"mnezerka/geonet/store"
	"os"

	"github.com/spf13/cobra"
)

var importCmd = &cobra.Command{
	Use:   "import FILEPATH",
	Short: "Import complete geonet from json file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		s := store.NewMongoStore(&config.Cfg)
		defer func() { s.Close() }()

		log.Infof("importing geonet from %s", args[0])

		// Read the JSON file
		file, err := os.Open(args[0])
		if err != nil {
			return err
		}
		defer file.Close()

		var importData store.DbContent

		// Decode the JSON file into the importData variable
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&importData)
		if err != nil {
			return err
		}

		s.Import(&importData)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(importCmd)
}
