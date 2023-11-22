package cmd

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

type CommonParams struct {
	SpotsOnly     bool
	Treshold      float64
	ShowHulls     bool
	ShowHeatSpots bool
	ShowGpxLines  bool
	ShowLines     bool
}

func addCommonFlags(params *CommonParams, cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVar(&params.SpotsOnly, "spots-only", false, "store spots only, do not connect by lines")
	cmd.PersistentFlags().BoolVar(&params.ShowHulls, "show-hulls", false, "show hull rectangles")
	cmd.PersistentFlags().BoolVar(&params.ShowHeatSpots, "show-heat-spots", false, "show hulls as spot circles with variable diameter")
	cmd.PersistentFlags().BoolVar(&params.ShowGpxLines, "show-gpx-lines", false, "show original gpx lines")
	cmd.PersistentFlags().BoolVar(&params.ShowLines, "show-lines", false, "show computed lines connecting hulls")
	cmd.PersistentFlags().Float64Var(&params.Treshold, "treshold", 0.0002, "treshold in meters")
}

func getFileHash(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
