package cmd

import (
	"github.com/spf13/cobra"
)

type TrackMeta struct {
	Title string `json:"title"`
	Url   string `json:"url"`
}

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch date from various sources",
}

func init() {
	rootCmd.AddCommand(fetchCmd)
}
