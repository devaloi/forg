package cmd

import (
	"fmt"
	"os"

	"github.com/devaloi/forg/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate a sample .forg.yaml configuration file",
	RunE: func(_ *cobra.Command, _ []string) error {
		const filename = ".forg.yaml"

		if _, err := os.Stat(filename); err == nil {
			return fmt.Errorf("%s already exists; remove it first or edit it directly", filename)
		}

		if err := os.WriteFile(filename, []byte(config.SampleConfig()), 0o600); err != nil {
			return fmt.Errorf("writing %s: %w", filename, err)
		}

		logger("Created %s — edit it to define your rules, then run 'forg run'.", filename)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
