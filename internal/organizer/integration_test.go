package organizer_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jasonaloi/forg/internal/config"
	"github.com/jasonaloi/forg/internal/organizer"
)

func noopLogger(string, ...interface{}) {}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func TestIntegration_FullPipeline(t *testing.T) {
	// Redirect HOME so the undo log doesn't touch the real home directory.
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	tmpdir := t.TempDir()

	sourceDir := filepath.Join(tmpdir, "source")
	destImages := filepath.Join(tmpdir, "dest_images")
	destDocs := filepath.Join(tmpdir, "dest_docs")
	destArchives := filepath.Join(tmpdir, "dest_archives")

	for _, d := range []string{sourceDir, destImages, destDocs, destArchives} {
		if err := os.MkdirAll(d, 0o750); err != nil {
			t.Fatalf("creating dir %s: %v", d, err)
		}
	}

	// Create source files with non-empty content.
	sourceFiles := map[string]string{
		"photo.jpg":    "jpeg image data",
		"document.pdf": "pdf document data",
		"data.csv":     "col1,col2,col3\na,b,c",
		"archive.zip":  "PK zip archive data",
		"random.xyz":   "unknown format data",
	}
	for name, content := range sourceFiles {
		if err := os.WriteFile(filepath.Join(sourceDir, name), []byte(content), 0o600); err != nil {
			t.Fatalf("creating source file %s: %v", name, err)
		}
	}

	cfg := &config.Config{
		Source:   sourceDir,
		Conflict: "skip",
		Rules: []config.RuleConfig{
			{
				Name:        "Images",
				Match:       config.MatchConfig{Extensions: []string{".jpg", ".png"}},
				Destination: destImages,
			},
			{
				Name:        "Documents",
				Match:       config.MatchConfig{Extensions: []string{".pdf", ".csv"}},
				Destination: destDocs,
			},
			{
				Name:        "Archives",
				Match:       config.MatchConfig{Extensions: []string{".zip"}},
				Destination: destArchives,
			},
		},
	}

	opts := organizer.Options{
		DryRun:     false,
		Verbose:    false,
		Recursive:  false,
		ConfigPath: "",
	}

	t.Run("DryRun", func(t *testing.T) {
		dryOpts := opts
		dryOpts.DryRun = true

		report, err := organizer.Run(cfg, dryOpts, noopLogger)
		if err != nil {
			t.Fatalf("Run(DryRun=true): %v", err)
		}

		if !report.DryRun {
			t.Error("expected report.DryRun to be true")
		}
		if report.Moved != 4 {
			t.Errorf("expected 4 moved in dry run, got %d", report.Moved)
		}

		// Source files must still exist after a dry run.
		for name := range sourceFiles {
			if !fileExists(filepath.Join(sourceDir, name)) {
				t.Errorf("source file %s should still exist after dry run", name)
			}
		}
	})

	t.Run("RealRun", func(t *testing.T) {
		report, err := organizer.Run(cfg, opts, noopLogger)
		if err != nil {
			t.Fatalf("Run(DryRun=false): %v", err)
		}

		if report.DryRun {
			t.Error("expected report.DryRun to be false")
		}
		if report.Moved != 4 {
			t.Errorf("expected 4 moved, got %d", report.Moved)
		}

		// Matched files should be gone from source.
		movedFiles := []string{"photo.jpg", "document.pdf", "data.csv", "archive.zip"}
		for _, name := range movedFiles {
			if fileExists(filepath.Join(sourceDir, name)) {
				t.Errorf("source file %s should have been moved", name)
			}
		}

		// Unmatched file should remain in source.
		if !fileExists(filepath.Join(sourceDir, "random.xyz")) {
			t.Error("random.xyz should still be in source (no matching rule)")
		}

		// Files should exist in destination directories.
		if !fileExists(filepath.Join(destImages, "photo.jpg")) {
			t.Error("photo.jpg should exist in dest_images")
		}
		if !fileExists(filepath.Join(destDocs, "document.pdf")) {
			t.Error("document.pdf should exist in dest_docs")
		}
		if !fileExists(filepath.Join(destDocs, "data.csv")) {
			t.Error("data.csv should exist in dest_docs")
		}
		if !fileExists(filepath.Join(destArchives, "archive.zip")) {
			t.Error("archive.zip should exist in dest_archives")
		}
	})

	t.Run("Undo", func(t *testing.T) {
		undoLog, err := organizer.ReadUndoLog()
		if err != nil {
			t.Fatalf("ReadUndoLog: %v", err)
		}

		if len(undoLog.Operations) != 4 {
			t.Fatalf("expected 4 undo operations, got %d", len(undoLog.Operations))
		}

		if err := organizer.ExecuteUndo(undoLog, false, noopLogger); err != nil {
			t.Fatalf("ExecuteUndo: %v", err)
		}

		// All matched files should be back in source.
		for name := range sourceFiles {
			if !fileExists(filepath.Join(sourceDir, name)) {
				t.Errorf("file %s should be back in source after undo", name)
			}
		}

		// Destination dirs should no longer contain the files.
		if fileExists(filepath.Join(destImages, "photo.jpg")) {
			t.Error("photo.jpg should no longer be in dest_images after undo")
		}
		if fileExists(filepath.Join(destDocs, "document.pdf")) {
			t.Error("document.pdf should no longer be in dest_docs after undo")
		}
		if fileExists(filepath.Join(destDocs, "data.csv")) {
			t.Error("data.csv should no longer be in dest_docs after undo")
		}
		if fileExists(filepath.Join(destArchives, "archive.zip")) {
			t.Error("archive.zip should no longer be in dest_archives after undo")
		}
	})
}

func TestIntegration_EmptySource(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	tmpdir := t.TempDir()
	sourceDir := filepath.Join(tmpdir, "source")
	destDir := filepath.Join(tmpdir, "dest")

	if err := os.MkdirAll(sourceDir, 0o750); err != nil {
		t.Fatalf("creating source dir: %v", err)
	}
	if err := os.MkdirAll(destDir, 0o750); err != nil {
		t.Fatalf("creating dest dir: %v", err)
	}

	cfg := &config.Config{
		Source:   sourceDir,
		Conflict: "skip",
		Rules: []config.RuleConfig{
			{
				Name:        "Images",
				Match:       config.MatchConfig{Extensions: []string{".jpg", ".png"}},
				Destination: destDir,
			},
		},
	}

	report, err := organizer.Run(cfg, organizer.Options{}, noopLogger)
	if err != nil {
		t.Fatalf("Run on empty source: %v", err)
	}

	if report.Moved != 0 {
		t.Errorf("expected 0 moved for empty source, got %d", report.Moved)
	}
}

func TestIntegration_NoMatchingRules(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	tmpdir := t.TempDir()
	sourceDir := filepath.Join(tmpdir, "source")
	destDir := filepath.Join(tmpdir, "dest")

	if err := os.MkdirAll(sourceDir, 0o750); err != nil {
		t.Fatalf("creating source dir: %v", err)
	}
	if err := os.MkdirAll(destDir, 0o750); err != nil {
		t.Fatalf("creating dest dir: %v", err)
	}

	// Create files that won't match any rule.
	for _, name := range []string{"file.abc", "file.def", "file.ghi"} {
		if err := os.WriteFile(filepath.Join(sourceDir, name), []byte("content"), 0o600); err != nil {
			t.Fatalf("creating file %s: %v", name, err)
		}
	}

	cfg := &config.Config{
		Source:   sourceDir,
		Conflict: "skip",
		Rules: []config.RuleConfig{
			{
				Name:        "Images",
				Match:       config.MatchConfig{Extensions: []string{".jpg", ".png"}},
				Destination: destDir,
			},
		},
	}

	report, err := organizer.Run(cfg, organizer.Options{}, noopLogger)
	if err != nil {
		t.Fatalf("Run with no matching rules: %v", err)
	}

	if report.Moved != 0 {
		t.Errorf("expected 0 moved, got %d", report.Moved)
	}

	// All source files should be untouched.
	for _, name := range []string{"file.abc", "file.def", "file.ghi"} {
		if !fileExists(filepath.Join(sourceDir, name)) {
			t.Errorf("file %s should still exist in source", name)
		}
	}
}

func TestIntegration_DestinationAutoCreate(t *testing.T) {
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	tmpdir := t.TempDir()
	sourceDir := filepath.Join(tmpdir, "source")
	// Destination does not exist yet — should be auto-created.
	destDir := filepath.Join(tmpdir, "auto", "created", "dest")

	if err := os.MkdirAll(sourceDir, 0o750); err != nil {
		t.Fatalf("creating source dir: %v", err)
	}

	if err := os.WriteFile(filepath.Join(sourceDir, "report.pdf"), []byte("pdf content"), 0o600); err != nil {
		t.Fatalf("creating source file: %v", err)
	}

	cfg := &config.Config{
		Source:   sourceDir,
		Conflict: "skip",
		Rules: []config.RuleConfig{
			{
				Name:        "Documents",
				Match:       config.MatchConfig{Extensions: []string{".pdf"}},
				Destination: destDir,
			},
		},
	}

	report, err := organizer.Run(cfg, organizer.Options{}, noopLogger)
	if err != nil {
		t.Fatalf("Run with auto-create destination: %v", err)
	}

	if report.Moved != 1 {
		t.Errorf("expected 1 moved, got %d", report.Moved)
	}

	if !fileExists(filepath.Join(destDir, "report.pdf")) {
		t.Error("report.pdf should exist in auto-created destination")
	}

	if fileExists(filepath.Join(sourceDir, "report.pdf")) {
		t.Error("report.pdf should have been moved from source")
	}
}
