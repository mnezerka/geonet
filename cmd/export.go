package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/spf13/cobra"
	"mnezerka/geonet/config"
	"mnezerka/geonet/log"
	"mnezerka/geonet/store"
)

var exportCmdSimplify bool

type exportBundle struct {
	GeoJson string
	Meta    store.DbMeta
}

var exportCmd = &cobra.Command{
	Use:   "export",
	Short: "Export geonet in format consumable by 3rd party apps",
	RunE: func(cmd *cobra.Command, args []string) error {

		store := store.NewMongoStore(&config.Cfg)
		defer func() { store.Close() }()

		log.Info("stashing geonet")
		store.StashPush()

		if exportCmdSimplify {
			log.Info("simplyfying geonet")
			store.Simplify()
		}

		var err error

		// 1. ---------------- points, edges -> geojson
		featureCollection, err := store.ToGeoJson()
		if err != nil {
			return err
		}

		// collection -> json
		geoJsonData, err := featureCollection.MarshalJSON()
		if err != nil {
			return err
		}

		// 2. ---------------- meta data (e.g. tracks)
		meta, err := store.GetMeta()
		if err != nil {
			return err
		}

		export := exportBundle{
			GeoJson: string(geoJsonData),
			Meta:    meta,
		}

		sJson, err := json.MarshalIndent(export, "", " ")
		if err != nil {
			return err
		}

		fmt.Print(string(sJson))

		log.Info("unstashing geonet")
		store.StashPop()

		return nil
	},
}

func init() {
	exportCmd.PersistentFlags().BoolVarP(&exportCmdSimplify, "simplify", "s", false, "simplify geonet")
	exportCmd.PersistentFlags().Int64Var(&config.Cfg.SimplifyMinDistance, "sim-min-dist", config.Cfg.MatchMaxDistance, "minimal distance between points for simplification")

	rootCmd.AddCommand(exportCmd)
}
