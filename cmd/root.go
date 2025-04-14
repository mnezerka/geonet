package cmd

import (
	"fmt"
	"os"
	"strings"

	"mnezerka/geonet/log"

	"github.com/spf13/cobra"
)

var (
	v         bool
	q         bool
	verbosity log.LevelName
)

var rootCmd = &cobra.Command{
	Use: "geonet utility",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().BoolVarP(&v, "verbose", "v", false, "verbose output (set verbosity level to DEBUG)")
	rootCmd.PersistentFlags().BoolVarP(&q, "quiet", "q", false, "no output at all (set verbosity level to ERROR)")
	rootCmd.PersistentFlags().Var(&verbosity, "verbosity", fmt.Sprintf("verbosity level (%s)", strings.Join(log.GetSupportedLevels(), ",")))
}

func initConfig() {

	flagCount := 0

	if v {
		log.SetVerbosity(log.LOG_LEVEL_DEBUG)
		flagCount++
	}
	if q {
		log.SetVerbosity(log.LOG_LEVEL_ERROR)
		flagCount++
	}
	if len(verbosity) > 0 {
		log.SetVerbosityByName(verbosity)
		flagCount++
	}

	if flagCount > 1 {
		fmt.Fprintln(os.Stderr, "flags -v, -q, and --verbosity are mutually exclusive; only one can be used at a time")
		os.Exit(1)
	}
}
