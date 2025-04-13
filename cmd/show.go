package cmd

import (
	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show",
	Short: "Show information of geonet",
	RunE: func(cmd *cobra.Command, args []string) error {

		/*
			store := store.NewMongoStore(&config.Cfg)
			defer func() { store.Close() }()

			meta, err := store.GetMeta()
			if err != nil {
				return err
			}

			logLines, err := store.GetLog()
			if err != nil {
				return err
			}

			fmt.Printf("tracks total: %d\n", len(meta.Tracks))

			fmt.Println("----------------- tracks --------------")

			w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ':', 0)

			for i := 0; i < len(meta.Tracks); i++ {
				t := meta.Tracks[i]
				fmt.Fprintf(w, "%d\t%s\n", t.Id, t.Meta.TrackTitle)
			}
			w.Flush()

			fmt.Println("----------------- log -----------------")
			for i := 0; i < len(logLines); i++ {
				fmt.Println(logLines[i].Msg)
			}

		*/

		return nil
	},
}

func init() {

	rootCmd.AddCommand(showCmd)
}
