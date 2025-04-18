package cmd

import (
	"fmt"
	"mnezerka/geonet/config"
	"mnezerka/geonet/tracks"

	"github.com/spf13/cobra"
)

var cmdTracksExport bool
var cmdTracksExportFormat string
var cmdTracksInterpolate bool

var cmdTracks = &cobra.Command{
	Use:   "tracks [FILES]",
	Short: "Process individual tracks, no relation to building geonet",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		set := tracks.NewSet(args)

		if cmdTracksInterpolate {
			set.InterpolateDistance(config.Cfg.InterpolationDistance)
		}

		if cmdTracksExport {
			switch cmdTracksExportFormat {

			case "geojson":
				fmt.Print(string(set.ExportGeoJson()))
				break
			case "svg":
				fmt.Print(set.ExportSvg())
				break
			default:
				return fmt.Errorf("unknown export format: %s", cmdTracksExportFormat)
			}
		}

		return nil
	},
}

func init() {
	// export
	cmdTracks.PersistentFlags().BoolVarP(&cmdTracksExport, "export", "e", false, "export tracks")
	cmdTracks.PersistentFlags().StringVar(&cmdTracksExportFormat, "export-format", "json", "export format (json, geojson, svg)")
	cmdTracks.PersistentFlags().BoolVar(&config.Cfg.ShowPoints, "points", config.Cfg.ShowPoints, "render individual points")

	addExportSvgFlags(cmdTracks)

	// interpolate
	cmdTracks.PersistentFlags().BoolVarP(&cmdTracksInterpolate, "interpolate", "i", false, "interpolate tracks")
	cmdTracks.PersistentFlags().Int64Var(&config.Cfg.InterpolationDistance, "int-dist", config.Cfg.InterpolationDistance, "distance for interpolation in meters")

	rootCmd.AddCommand(cmdTracks)
}
