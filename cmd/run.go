package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/jasonaloi/forg/internal/config"
	"github.com/jasonaloi/forg/internal/organizer"
	"github.com/spf13/cobra"
)

var (
	dryRun        bool
	recursive     bool
	includeHidden bool
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Execute organizing rules and move files",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.Load(cfgFile)
		if err != nil {
			return fmt.Errorf("loading config: %w", err)
		}

		opts := organizer.Options{
			DryRun:        dryRun,
			Verbose:       verbose,
			Recursive:     recursive,
			IncludeHidden: includeHidden,
			ConfigPath:    cfgFile,
		}

		report, err := organizer.Run(cfg, opts, logger)
		if err != nil {
			return fmt.Errorf("running organizer: %w", err)
		}

		printReport(report)
		return nil
	},
}

func init() {
	runCmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would happen without moving files")
	runCmd.Flags().BoolVarP(&recursive, "recursive", "r", false, "scan directories recursively")
	runCmd.Flags().BoolVar(&includeHidden, "include-hidden", false, "include hidden files and directories")
	rootCmd.AddCommand(runCmd)
}

// printReport displays the results of an organizer run.
func printReport(report *organizer.Report) {
	if quiet {
		return
	}

	if report.DryRun {
		fmt.Println("--- Dry Run ---")
		if len(report.Operations) == 0 {
			fmt.Println("No files matched.")
			return
		}
		printTable(report.Operations)
		fmt.Printf("\n%d file(s) would be moved.\n", len(report.Operations))
		return
	}

	fmt.Printf("Moved %d file(s) (%d skipped, %d conflict(s))\n",
		report.Moved, report.Skipped, report.Conflicts)
}

// printTable renders a formatted table of move operations.
func printTable(ops []organizer.MoveOp) {
	fileHeader := "File"
	ruleHeader := "Rule"
	destHeader := "Destination"

	fileWidth := len(fileHeader)
	ruleWidth := len(ruleHeader)
	destWidth := len(destHeader)

	for _, op := range ops {
		sp := shortPath(op.Source)
		if len(sp) > fileWidth {
			fileWidth = len(sp)
		}
		if len(op.RuleName) > ruleWidth {
			ruleWidth = len(op.RuleName)
		}
		dp := shortPath(op.Destination)
		if len(dp) > destWidth {
			destWidth = len(dp)
		}
	}

	format := fmt.Sprintf("  %%-%ds  %%-%ds  %%-%ds\n", fileWidth, ruleWidth, destWidth)
	sep := fmt.Sprintf("  %s  %s  %s\n",
		repeat("\u2500", fileWidth),
		repeat("\u2500", ruleWidth),
		repeat("\u2500", destWidth),
	)

	fmt.Printf(format, fileHeader, ruleHeader, destHeader)
	fmt.Print(sep)
	for _, op := range ops {
		fmt.Printf(format, shortPath(op.Source), op.RuleName, shortPath(op.Destination))
	}
}

// shortPath replaces the user's home directory prefix with ~ for brevity.
func shortPath(path string) string {
	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	rel, err := filepath.Rel(home, abs)
	if err != nil || len(rel) > 1 && rel[:2] == ".." {
		return path
	}
	return filepath.Join("~", rel)
}

// repeat returns a string consisting of s repeated n times.
func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}
