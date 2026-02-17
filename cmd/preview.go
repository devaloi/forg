package cmd

import (
	"fmt"

	"github.com/jasonaloi/forg/internal/config"
	"github.com/jasonaloi/forg/internal/organizer"
	"github.com/spf13/cobra"
)

var previewCmd = &cobra.Command{
	Use:   "preview",
	Short: "Show what forg would do without moving any files",
	RunE: func(cmd *cobra.Command, _ []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		previewRecursive, _ := cmd.Flags().GetBool("recursive")
		previewHidden, _ := cmd.Flags().GetBool("include-hidden")

		opts := organizer.Options{
			DryRun:        true,
			Verbose:       verbose,
			Recursive:     previewRecursive,
			IncludeHidden: previewHidden,
			ConfigPath:    cfgFile,
		}

		report, err := organizer.Run(cfg, opts, logger)
		if err != nil {
			return fmt.Errorf("running preview: %w", err)
		}

		printReport(report)
		return nil
	},
}

func init() {
	previewCmd.Flags().BoolP("recursive", "r", false, "scan directories recursively")
	previewCmd.Flags().Bool("include-hidden", false, "include hidden files and directories")
	rootCmd.AddCommand(previewCmd)
}
