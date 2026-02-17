package cmd

import (
	"fmt"

	"github.com/jasonaloi/forg/internal/organizer"
	"github.com/spf13/cobra"
)

var undoCmd = &cobra.Command{
	Use:   "undo",
	Short: "Reverse the most recent forg run",
	RunE: func(_ *cobra.Command, _ []string) error {
		log, err := organizer.ReadUndoLog()
		if err != nil {
			return fmt.Errorf("reading undo log: %w", err)
		}

		logger("Undoing %d operation(s) from %s ...",
			len(log.Operations), log.Timestamp.Format("2006-01-02 15:04:05"))

		if err := organizer.ExecuteUndo(log, verbose, logger); err != nil {
			return fmt.Errorf("executing undo: %w", err)
		}

		if err := organizer.DeleteUndoLog(); err != nil {
			return fmt.Errorf("cleaning up undo log: %w", err)
		}

		logger("Undo complete.")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(undoCmd)
}
