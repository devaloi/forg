package organizer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FileSystem abstracts file-system operations so that the executor can be
// tested without touching the real file system.
type FileSystem interface {
	Rename(oldpath, newpath string) error
	MkdirAll(path string, perm os.FileMode) error
	Stat(path string) (os.FileInfo, error)
}

// OSFileSystem implements FileSystem using the standard os package.
type OSFileSystem struct{}

// Rename renames (moves) oldpath to newpath.
func (OSFileSystem) Rename(oldpath, newpath string) error { return os.Rename(oldpath, newpath) }

// MkdirAll creates a directory path and all parents that do not yet exist.
func (OSFileSystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

// Stat returns the FileInfo for the named file.
func (OSFileSystem) Stat(path string) (os.FileInfo, error) { return os.Stat(path) }

// Executor moves files according to a plan, handling conflicts and logging.
type Executor struct {
	fs       FileSystem
	conflict string
	verbose  bool
	logger   func(string, ...interface{})
}

// NewExecutor creates an Executor that uses the real OS file system.
func NewExecutor(conflict string, verbose bool, logger func(string, ...interface{})) *Executor {
	return NewExecutorWithFS(OSFileSystem{}, conflict, verbose, logger)
}

// NewExecutorWithFS creates an Executor backed by the provided FileSystem.
// If logger is nil a no-op logger is used.
func NewExecutorWithFS(fs FileSystem, conflict string, verbose bool, logger func(string, ...interface{})) *Executor {
	if logger == nil {
		logger = func(string, ...interface{}) {}
	}
	return &Executor{
		fs:       fs,
		conflict: conflict,
		verbose:  verbose,
		logger:   logger,
	}
}

// Execute runs every operation in plan, moving files to their destinations.
// When dryRun is true no files are moved; the returned report still describes
// what would happen. The returned UndoEntry slice records every successful
// move so it can be reversed later.
func (e *Executor) Execute(plan []MoveOp, dryRun bool) (*Report, []UndoEntry) {
	report := &Report{DryRun: dryRun}
	var undoEntries []UndoEntry

	for _, op := range plan {
		destPath := filepath.Join(op.Destination, filepath.Base(op.Source))

		if dryRun {
			report.Operations = append(report.Operations, op)
			report.Moved++
			if e.verbose {
				e.logger("[dry-run] %s -> %s (rule: %s)", op.Source, destPath, op.RuleName)
			}
			continue
		}

		if err := e.fs.MkdirAll(op.Destination, 0o755); err != nil {
			e.logger("error creating directory %s: %v", op.Destination, err)
			report.Errors++
			continue
		}

		finalDest, hadConflict, err := e.resolveConflict(destPath)
		if err != nil {
			e.logger("error resolving conflict for %s: %v", destPath, err)
			report.Errors++
			continue
		}

		if finalDest == "" {
			// skip strategy
			report.Skipped++
			report.Conflicts++
			if e.verbose {
				e.logger("skipped %s (conflict at %s)", op.Source, destPath)
			}
			continue
		}

		if hadConflict && e.verbose {
			e.logger("conflict resolved for %s -> %s", destPath, finalDest)
		}

		if err := e.fs.Rename(op.Source, finalDest); err != nil {
			e.logger("error moving %s to %s: %v", op.Source, finalDest, err)
			report.Errors++
			continue
		}

		undoEntries = append(undoEntries, UndoEntry{From: op.Source, To: finalDest})
		report.Moved++

		if e.verbose {
			e.logger("moved %s -> %s (rule: %s)", op.Source, finalDest, op.RuleName)
		}
	}

	return report, undoEntries
}

// resolveConflict determines the final destination path when a file already
// exists at destPath. It applies the executor's conflict strategy.
func (e *Executor) resolveConflict(destPath string) (string, bool, error) {
	_, err := e.fs.Stat(destPath)
	if err != nil {
		if os.IsNotExist(err) {
			return destPath, false, nil
		}
		return "", false, fmt.Errorf("stat %q: %w", destPath, err)
	}

	// File exists — apply conflict strategy.
	switch e.conflict {
	case "overwrite":
		return destPath, true, nil
	case "rename":
		newPath, err := e.findUniqueName(destPath)
		if err != nil {
			return "", false, fmt.Errorf("finding unique name for %q: %w", destPath, err)
		}
		return newPath, true, nil
	case "skip":
		return "", true, nil
	default:
		// Default to skip when no strategy is configured.
		return "", true, nil
	}
}

// findUniqueName generates a path like base-1.ext, base-2.ext, … up to 1000.
func (e *Executor) findUniqueName(destPath string) (string, error) {
	dir := filepath.Dir(destPath)
	ext := filepath.Ext(destPath)
	base := strings.TrimSuffix(filepath.Base(destPath), ext)

	for i := 1; i <= 1000; i++ {
		candidate := filepath.Join(dir, fmt.Sprintf("%s-%d%s", base, i, ext))
		_, err := e.fs.Stat(candidate)
		if err != nil {
			if os.IsNotExist(err) {
				return candidate, nil
			}
			return "", fmt.Errorf("stat %q: %w", candidate, err)
		}
	}
	return "", fmt.Errorf("could not find unique name for %q after 1000 attempts", destPath)
}
