package cmd

import (
	"encoding/json"
	"fmt"

	"mnezerka/geonet/config"
	"mnezerka/geonet/log"
	"mnezerka/geonet/store"

	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export complete geonet",
	RunE: func(cmd *cobra.Command, args []string) error {

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

		return nil
	},
}

func init() {
	rootCmd.AddCommand(exportCmd)
}
