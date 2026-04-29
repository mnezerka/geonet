package cmd

import (
	"fmt"
	"mnezerka/geonet/log"
	"mnezerka/geonet/tracks"
	"mnezerka/geonet/utils"
	"net/url"

	"github.com/spf13/cobra"
)

var fetchWandererCmdDir string
var fetchWandererCmdForce bool
var fetchWandererCmdLimit int

var fetchWandererCmd = &cobra.Command{
	Use:   "wanderer url",
	Short: "Fetch tracks from site powered by wanderer hugo template",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {

		log.Infof("limit: %d", fetchWandererCmdLimit)

		// guess if first arg is file path or url
		urlCheck, err := url.Parse(args[0])
		if err != nil {
			return err
		}

		var m []tracks.WandererPost

		if urlCheck.Scheme == "" {
			m, err = tracks.WandererReadMediaIndexFromFileSystem(args[0])
		} else {
			m, err = tracks.WandererReadMediaIndexFromUrl(args[0])

		}
		if err != nil {
			return err
		}

		// ensure directory to be used for fetched tracks exists
		err = utils.EnsureDirExists(fetchWandererCmdDir)
		if err != nil {
			return err
		}

		log.Infof("%d wanderer posts will be processed", len(m))

		downloaded := 0
		total := 0

		for i := 0; i < len(m); i++ {

			// interrupt loop if number of tracks is limited
			if fetchWandererCmdLimit >= 0 {
				if i >= fetchWandererCmdLimit {
					break
				}
			}

			for j := 0; j < len(m[i].Tracks); j++ {
				total++
				fetched, err := tracks.WandererStoreTrackMeta(
					&m[i],
					&m[i].Tracks[j],
					args[0],
					fetchWandererCmdDir,
					fetchWandererCmdForce)
				if err != nil {
					return err
				}
				if fetched {
					downloaded++
					fmt.Printf("downloaded: %s\n", m[i].Tracks[j].Url)
				}
			}
		}

		fmt.Printf("summary: %d downloaded, %d skipped, %d total\n", downloaded, total-downloaded, total)

		return nil
	},
}

func init() {
	fetchWandererCmd.PersistentFlags().StringVar(&fetchWandererCmdDir, "dir", "data", "directory for storing fetched tracks")
	fetchWandererCmd.PersistentFlags().BoolVarP(&fetchWandererCmdForce, "force", "f", false, "download all tracks, no matter if already fetched")
	fetchWandererCmd.PersistentFlags().IntVar(&fetchWandererCmdLimit, "limit", -1, "number of tracks to be read")

	fetchCmd.AddCommand(fetchWandererCmd)
}
