package cmd

import (
	"github.com/spf13/cobra"
)

var loadCmd = &cobra.Command{
	Use:   "load FILEPATH",
	Short: "Load complete geonet from json file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		/*
			s := store.NewMongoStore(&config.Cfg)
			defer func() { s.Close() }()

			log.Infof("loading geonet from %s", args[0])

			// Read the JSON file
			file, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer file.Close()

			var loadedData store.DbContent

			// Decode the JSON file into the importData variable
			decoder := json.NewDecoder(file)
			err = decoder.Decode(&loadedData)
			if err != nil {
				return err
			}

			s.Import(&loadedData)
		*/

		return nil
	},
}

func init() {
	rootCmd.AddCommand(loadCmd)
}
