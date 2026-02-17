package organizer

import (
	"fmt"
	"time"

	"github.com/jasonaloi/forg/internal/config"
	"github.com/jasonaloi/forg/internal/rules"
	"github.com/jasonaloi/forg/internal/scanner"
)

// Options controls the behaviour of a Run invocation.
type Options struct {
	DryRun        bool
	Verbose       bool
	Recursive     bool
	IncludeHidden bool
	ConfigPath    string
}

// Run executes the full organise workflow: scan the source directory, build a
// plan from the configured rules, execute the plan, and optionally write an
// undo log.
func Run(cfg *config.Config, opts Options, logger func(string, ...interface{})) (*Report, error) {
	if logger == nil {
		logger = func(string, ...interface{}) {}
	}

	engine, err := rules.NewEngine(cfg.Rules)
	if err != nil {
		return nil, fmt.Errorf("building rule engine: %w", err)
	}

	sc := scanner.New(scanner.Options{
		Recursive:     opts.Recursive,
		IncludeHidden: opts.IncludeHidden,
	})

	source, err := config.ExpandPath(cfg.Source)
	if err != nil {
		return nil, fmt.Errorf("expanding source path: %w", err)
	}

	files, err := sc.Scan(source)
	if err != nil {
		return nil, fmt.Errorf("scanning source directory: %w", err)
	}

	plan := BuildPlan(files, engine)

	executor := NewExecutor(cfg.Conflict, opts.Verbose, logger)
	report, undoEntries := executor.Execute(plan, opts.DryRun)

	if !opts.DryRun && len(undoEntries) > 0 {
		undoLog := &UndoLog{
			Timestamp:  time.Now(),
			Config:     opts.ConfigPath,
			Operations: undoEntries,
		}
		if err := WriteUndoLog(undoLog); err != nil {
			return report, fmt.Errorf("writing undo log: %w", err)
		}
	}

	return report, nil
}
