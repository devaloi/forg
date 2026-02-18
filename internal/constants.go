// Package internal defines shared constants used across the forg codebase.
package internal

import "os"

const (
	// MaxRenameAttempts is the upper bound on rename-suffix attempts when
	// resolving file-name conflicts (e.g. file-1.txt … file-1000.txt).
	MaxRenameAttempts = 1000

	// DefaultDirPerms is the permission mode used when creating directories.
	DefaultDirPerms os.FileMode = 0o750

	// DefaultConfigFile is the default configuration file name.
	DefaultConfigFile = ".forg.yaml"

	// UndoLogDir is the directory name (under $HOME) that stores undo state.
	UndoLogDir = ".forg"

	// UndoLogFile is the file name used for the JSON undo log.
	UndoLogFile = "undo.json"

	// TimeFormat is the timestamp layout used when displaying undo metadata.
	TimeFormat = "2006-01-02 15:04:05"

	// ConflictSkip leaves the destination file untouched and skips the move.
	ConflictSkip = "skip"

	// ConflictRename appends a numeric suffix to avoid overwriting.
	ConflictRename = "rename"

	// ConflictOverwrite replaces the existing destination file.
	ConflictOverwrite = "overwrite"
)

// ValidConflictStrategy reports whether s is a recognised conflict strategy.
func ValidConflictStrategy(s string) bool {
	switch s {
	case ConflictSkip, ConflictRename, ConflictOverwrite:
		return true
	default:
		return false
	}
}
