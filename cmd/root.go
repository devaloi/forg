// Package cmd implements the CLI commands for forg.
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgFile string
	verbose bool
	quiet   bool
	version = "0.1.0"
)

var rootCmd = &cobra.Command{
	Use:     "forg",
	Short:   "A smart file organizer CLI",
	Long:    "forg organises files into directories based on YAML rules.\nIt supports dry-run previews, recursive scanning, and undo\nso you can confidently tidy up any folder.",
	Version: version,
}

// Execute runs the root command and returns any error.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", ".forg.yaml", "path to configuration file")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "suppress all non-error output")
}

// logger prints a formatted message to stderr unless quiet mode is enabled.
func logger(format string, args ...interface{}) {
	if !quiet {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}
