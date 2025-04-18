package cmd

import (
	"github.com/spf13/cobra"
)

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch date from various sources",
}

func init() {
	rootCmd.AddCommand(fetchCmd)
}
