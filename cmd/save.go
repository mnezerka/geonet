package cmd

import (
	"github.com/spf13/cobra"
)

var saveCmd = &cobra.Command{
	Use:   "save",
	Short: "Save geonet to filesystem (complete image of database)",
	RunE: func(cmd *cobra.Command, args []string) error {

		/*
			store := store.NewMongoStore(&config.Cfg)
			defer func() { store.Close() }()

			log.Info("exporting geonet")

			export := store.Export()

			// meta - json
			sJson, err := json.MarshalIndent(export, "", " ")
			if err != nil {
				return err
			}

			fmt.Print(string(sJson))
		*/

		return nil
	},
}

func init() {
	rootCmd.AddCommand(saveCmd)
}
