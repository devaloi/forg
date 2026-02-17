// Package organizer orchestrates file scanning, rule matching, and file operations.
package organizer

import (
	"github.com/devaloi/forg/internal/rules"
	"github.com/devaloi/forg/internal/scanner"
)

// MoveOp represents a planned file move operation from a source path to a
// destination directory, triggered by a named rule.
type MoveOp struct {
	Source      string
	Destination string
	RuleName    string
}

// Report summarises the results of executing a plan.
type Report struct {
	Moved      int
	Skipped    int
	Conflicts  int
	Errors     int
	DryRun     bool
	Operations []MoveOp
}

// BuildPlan evaluates every scanned file against the rule engine and returns
// a slice of MoveOp entries for files that match at least one rule.
func BuildPlan(files []scanner.FileInfo, engine *rules.Engine) []MoveOp {
	var ops []MoveOp
	for _, f := range files {
		rule := engine.Match(f)
		if rule != nil {
			ops = append(ops, MoveOp{
				Source:      f.Path,
				Destination: rule.Destination,
				RuleName:    rule.Name,
			})
		}
	}
	return ops
}
