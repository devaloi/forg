package organizer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// UndoEntry records a single file move so it can be reversed.
type UndoEntry struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// UndoLog captures all operations from a single run together with metadata
// needed to reverse them.
type UndoLog struct {
	Timestamp  time.Time   `json:"timestamp"`
	Config     string      `json:"config"`
	Operations []UndoEntry `json:"operations"`
}

// UndoLogPath returns the default path for the undo log file
// (~/.forg/undo.json).
func UndoLogPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("resolving home directory: %w", err)
	}
	return filepath.Join(home, ".forg", "undo.json"), nil
}

// WriteUndoLog serialises log as JSON and writes it to the default undo log
// path, creating the parent directory if necessary.
func WriteUndoLog(log *UndoLog) error {
	path, err := UndoLogPath()
	if err != nil {
		return fmt.Errorf("determining undo log path: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating undo log directory: %w", err)
	}

	data, err := json.MarshalIndent(log, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling undo log: %w", err)
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("writing undo log to %s: %w", path, err)
	}

	return nil
}

// ReadUndoLog reads and deserialises the undo log from the default path.
// If the file does not exist a descriptive error is returned.
func ReadUndoLog() (*UndoLog, error) {
	path, err := UndoLogPath()
	if err != nil {
		return nil, fmt.Errorf("determining undo log path: %w", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no undo log found at %s: run 'forg run' first to create one", path)
		}
		return nil, fmt.Errorf("reading undo log from %s: %w", path, err)
	}

	var log UndoLog
	if err := json.Unmarshal(data, &log); err != nil {
		return nil, fmt.Errorf("parsing undo log from %s: %w", path, err)
	}

	return &log, nil
}

// DeleteUndoLog removes the undo log file.
func DeleteUndoLog() error {
	path, err := UndoLogPath()
	if err != nil {
		return fmt.Errorf("determining undo log path: %w", err)
	}

	if err := os.Remove(path); err != nil {
		return fmt.Errorf("removing undo log at %s: %w", path, err)
	}

	return nil
}

// ExecuteUndo reverses all operations recorded in the undo log, processing
// them in reverse order so that the most recent move is undone first.
func ExecuteUndo(log *UndoLog, verbose bool, logger func(string, ...interface{})) error {
	if logger == nil {
		logger = func(string, ...interface{}) {}
	}

	for i := len(log.Operations) - 1; i >= 0; i-- {
		entry := log.Operations[i]

		dir := filepath.Dir(entry.From)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("creating directory %s for undo: %w", dir, err)
		}

		if err := os.Rename(entry.To, entry.From); err != nil {
			return fmt.Errorf("undoing move %s -> %s: %w", entry.To, entry.From, err)
		}

		if verbose {
			logger("undo: %s -> %s", entry.To, entry.From)
		}
	}

	return nil
}
