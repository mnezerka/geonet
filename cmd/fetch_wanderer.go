package cmd

import (
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

		for i := 0; i < len(m); i++ {

			// interrupt loop if number of tracks is limited
			if fetchWandererCmdLimit >= 0 {
				if i >= fetchWandererCmdLimit {
					break
				}
			}

			log.Infof("post %d '%s' (%d tracks)", i, m[i].Title, len(m[i].Tracks))
			for j := 0; j < len(m[i].Tracks); j++ {
				tracks.WandererStoreTrackMeta(
					&m[i],
					&m[i].Tracks[j],
					args[0],
					fetchWandererCmdDir,
					fetchWandererCmdForce)
			}
		}

		return nil
	},
}

func init() {
	fetchWandererCmd.PersistentFlags().StringVar(&fetchWandererCmdDir, "dir", "data", "directory for storing fetched tracks")
	fetchWandererCmd.PersistentFlags().BoolVarP(&fetchWandererCmdForce, "force", "f", false, "download all tracks, no matter if already fetched")
	fetchWandererCmd.PersistentFlags().IntVar(&fetchWandererCmdLimit, "limit", -1, "number of tracks to be read")

	fetchCmd.AddCommand(fetchWandererCmd)
}
