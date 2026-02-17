package organizer

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jasonaloi/forg/internal/config"
	"github.com/jasonaloi/forg/internal/rules"
	"github.com/jasonaloi/forg/internal/scanner"
)

func createTempFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("creating temp file %s: %v", path, err)
	}
	return path
}

func TestBuildPlan(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	rulesCfg := []config.RuleConfig{
		{
			Name:        "images",
			Match:       config.MatchConfig{Extensions: []string{".png", ".jpg"}},
			Destination: destDir,
		},
	}

	engine, err := rules.NewEngine(rulesCfg)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	files := []scanner.FileInfo{
		{Path: filepath.Join(srcDir, "photo.png"), Name: "photo.png", Extension: ".png", Size: 100},
		{Path: filepath.Join(srcDir, "doc.txt"), Name: "doc.txt", Extension: ".txt", Size: 200},
		{Path: filepath.Join(srcDir, "pic.jpg"), Name: "pic.jpg", Extension: ".jpg", Size: 300},
	}

	plan := BuildPlan(files, engine)

	if len(plan) != 2 {
		t.Fatalf("expected 2 ops, got %d", len(plan))
	}

	t.Run("first op matches photo.png", func(t *testing.T) {
		if plan[0].Source != files[0].Path {
			t.Errorf("expected source %s, got %s", files[0].Path, plan[0].Source)
		}
		if plan[0].Destination != destDir {
			t.Errorf("expected destination %s, got %s", destDir, plan[0].Destination)
		}
		if plan[0].RuleName != "images" {
			t.Errorf("expected rule name %q, got %q", "images", plan[0].RuleName)
		}
	})

	t.Run("second op matches pic.jpg", func(t *testing.T) {
		if plan[1].Source != files[2].Path {
			t.Errorf("expected source %s, got %s", files[2].Path, plan[1].Source)
		}
	})
}

func TestBuildPlan_NoMatches(t *testing.T) {
	destDir := t.TempDir()

	rulesCfg := []config.RuleConfig{
		{
			Name:        "images",
			Match:       config.MatchConfig{Extensions: []string{".png"}},
			Destination: destDir,
		},
	}

	engine, err := rules.NewEngine(rulesCfg)
	if err != nil {
		t.Fatalf("NewEngine: %v", err)
	}

	srcDir := t.TempDir()
	files := []scanner.FileInfo{
		{Path: filepath.Join(srcDir, "readme.md"), Name: "readme.md", Extension: ".md", Size: 50},
		{Path: filepath.Join(srcDir, "main.go"), Name: "main.go", Extension: ".go", Size: 120},
	}

	plan := BuildPlan(files, engine)

	if len(plan) != 0 {
		t.Fatalf("expected 0 ops, got %d", len(plan))
	}
}

func TestExecute_DryRun(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	srcFile := createTempFile(t, srcDir, "file.txt", "hello")

	plan := []MoveOp{
		{Source: srcFile, Destination: destDir, RuleName: "test-rule"},
	}

	exec := NewExecutor("skip", false, nil)
	report, undoEntries := exec.Execute(plan, true)

	t.Run("report flags", func(t *testing.T) {
		if !report.DryRun {
			t.Error("expected DryRun to be true")
		}
		if report.Moved != 1 {
			t.Errorf("expected Moved=1, got %d", report.Moved)
		}
	})

	t.Run("operations populated", func(t *testing.T) {
		if len(report.Operations) != 1 {
			t.Fatalf("expected 1 operation, got %d", len(report.Operations))
		}
		if report.Operations[0].Source != srcFile {
			t.Errorf("expected source %s, got %s", srcFile, report.Operations[0].Source)
		}
	})

	t.Run("source file still exists", func(t *testing.T) {
		if _, err := os.Stat(srcFile); err != nil {
			t.Errorf("source file should still exist: %v", err)
		}
	})

	t.Run("no undo entries for dry run", func(t *testing.T) {
		if len(undoEntries) != 0 {
			t.Errorf("expected 0 undo entries, got %d", len(undoEntries))
		}
	})
}

func TestExecute_MoveFiles(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	src1 := createTempFile(t, srcDir, "a.txt", "aaa")
	src2 := createTempFile(t, srcDir, "b.txt", "bbb")

	plan := []MoveOp{
		{Source: src1, Destination: destDir, RuleName: "rule-a"},
		{Source: src2, Destination: destDir, RuleName: "rule-b"},
	}

	exec := NewExecutor("skip", false, nil)
	report, undoEntries := exec.Execute(plan, false)

	t.Run("report counts", func(t *testing.T) {
		if report.Moved != 2 {
			t.Errorf("expected Moved=2, got %d", report.Moved)
		}
		if report.DryRun {
			t.Error("expected DryRun to be false")
		}
	})

	t.Run("files moved to destination", func(t *testing.T) {
		for _, name := range []string{"a.txt", "b.txt"} {
			dest := filepath.Join(destDir, name)
			if _, err := os.Stat(dest); err != nil {
				t.Errorf("expected %s to exist at destination: %v", name, err)
			}
		}
	})

	t.Run("source files removed", func(t *testing.T) {
		for _, src := range []string{src1, src2} {
			if _, err := os.Stat(src); !os.IsNotExist(err) {
				t.Errorf("expected source %s to be gone, err=%v", src, err)
			}
		}
	})

	t.Run("undo entries returned", func(t *testing.T) {
		if len(undoEntries) != 2 {
			t.Fatalf("expected 2 undo entries, got %d", len(undoEntries))
		}
		if undoEntries[0].From != src1 {
			t.Errorf("expected undo From=%s, got %s", src1, undoEntries[0].From)
		}
		if undoEntries[0].To != filepath.Join(destDir, "a.txt") {
			t.Errorf("expected undo To=%s, got %s", filepath.Join(destDir, "a.txt"), undoEntries[0].To)
		}
	})
}

func TestExecute_ConflictSkip(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	srcFile := createTempFile(t, srcDir, "conflict.txt", "source content")
	createTempFile(t, destDir, "conflict.txt", "existing content")

	plan := []MoveOp{
		{Source: srcFile, Destination: destDir, RuleName: "skip-rule"},
	}

	exec := NewExecutor("skip", false, nil)
	report, _ := exec.Execute(plan, false)

	t.Run("file not moved", func(t *testing.T) {
		if _, err := os.Stat(srcFile); err != nil {
			t.Errorf("source file should still exist: %v", err)
		}
	})

	t.Run("report reflects skip", func(t *testing.T) {
		if report.Skipped != 1 {
			t.Errorf("expected Skipped=1, got %d", report.Skipped)
		}
		if report.Conflicts != 1 {
			t.Errorf("expected Conflicts=1, got %d", report.Conflicts)
		}
		if report.Moved != 0 {
			t.Errorf("expected Moved=0, got %d", report.Moved)
		}
	})
}

func TestExecute_ConflictRename(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	srcFile := createTempFile(t, srcDir, "file.txt", "new content")
	createTempFile(t, destDir, "file.txt", "existing content")

	plan := []MoveOp{
		{Source: srcFile, Destination: destDir, RuleName: "rename-rule"},
	}

	exec := NewExecutor("rename", false, nil)
	report, undoEntries := exec.Execute(plan, false)

	t.Run("report counts", func(t *testing.T) {
		if report.Moved != 1 {
			t.Errorf("expected Moved=1, got %d", report.Moved)
		}
	})

	t.Run("renamed file exists", func(t *testing.T) {
		renamed := filepath.Join(destDir, "file-1.txt")
		if _, err := os.Stat(renamed); err != nil {
			t.Errorf("expected renamed file %s to exist: %v", renamed, err)
		}
	})

	t.Run("original destination untouched", func(t *testing.T) {
		orig := filepath.Join(destDir, "file.txt")
		data, err := os.ReadFile(orig)
		if err != nil {
			t.Fatalf("reading original: %v", err)
		}
		if string(data) != "existing content" {
			t.Errorf("expected original content preserved, got %q", string(data))
		}
	})

	t.Run("undo entry points to renamed path", func(t *testing.T) {
		if len(undoEntries) != 1 {
			t.Fatalf("expected 1 undo entry, got %d", len(undoEntries))
		}
		expected := filepath.Join(destDir, "file-1.txt")
		if undoEntries[0].To != expected {
			t.Errorf("expected undo To=%s, got %s", expected, undoEntries[0].To)
		}
	})
}

func TestExecute_ConflictOverwrite(t *testing.T) {
	srcDir := t.TempDir()
	destDir := t.TempDir()

	srcFile := createTempFile(t, srcDir, "file.txt", "new content")
	createTempFile(t, destDir, "file.txt", "old content")

	plan := []MoveOp{
		{Source: srcFile, Destination: destDir, RuleName: "overwrite-rule"},
	}

	exec := NewExecutor("overwrite", false, nil)
	report, _ := exec.Execute(plan, false)

	t.Run("report counts", func(t *testing.T) {
		if report.Moved != 1 {
			t.Errorf("expected Moved=1, got %d", report.Moved)
		}
	})

	t.Run("destination has new content", func(t *testing.T) {
		dest := filepath.Join(destDir, "file.txt")
		data, err := os.ReadFile(dest)
		if err != nil {
			t.Fatalf("reading destination: %v", err)
		}
		if string(data) != "new content" {
			t.Errorf("expected overwritten content %q, got %q", "new content", string(data))
		}
	})

	t.Run("source file removed", func(t *testing.T) {
		if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
			t.Errorf("expected source file to be gone, err=%v", err)
		}
	})
}

func TestExecute_CreatesDestDir(t *testing.T) {
	srcDir := t.TempDir()
	destDir := filepath.Join(t.TempDir(), "nested", "dest")

	srcFile := createTempFile(t, srcDir, "moveme.txt", "data")

	plan := []MoveOp{
		{Source: srcFile, Destination: destDir, RuleName: "mkdir-rule"},
	}

	exec := NewExecutor("skip", false, nil)
	report, _ := exec.Execute(plan, false)

	t.Run("destination dir created", func(t *testing.T) {
		info, err := os.Stat(destDir)
		if err != nil {
			t.Fatalf("expected dest dir to exist: %v", err)
		}
		if !info.IsDir() {
			t.Error("expected dest path to be a directory")
		}
	})

	t.Run("file moved", func(t *testing.T) {
		if report.Moved != 1 {
			t.Errorf("expected Moved=1, got %d", report.Moved)
		}
		dest := filepath.Join(destDir, "moveme.txt")
		if _, err := os.Stat(dest); err != nil {
			t.Errorf("expected file at destination: %v", err)
		}
	})
}

func TestExecuteUndo(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()

	origPath := filepath.Join(dirA, "undome.txt")
	movedPath := filepath.Join(dirB, "undome.txt")

	if err := os.WriteFile(movedPath, []byte("undo data"), 0o644); err != nil {
		t.Fatalf("writing moved file: %v", err)
	}

	undoLog := &UndoLog{
		Timestamp: time.Now(),
		Config:    "test",
		Operations: []UndoEntry{
			{From: origPath, To: movedPath},
		},
	}

	if err := ExecuteUndo(undoLog, false, nil); err != nil {
		t.Fatalf("ExecuteUndo: %v", err)
	}

	t.Run("file restored to original location", func(t *testing.T) {
		data, err := os.ReadFile(origPath)
		if err != nil {
			t.Fatalf("expected file at original location: %v", err)
		}
		if string(data) != "undo data" {
			t.Errorf("expected content %q, got %q", "undo data", string(data))
		}
	})

	t.Run("file removed from moved location", func(t *testing.T) {
		if _, err := os.Stat(movedPath); !os.IsNotExist(err) {
			t.Errorf("expected moved file to be gone, err=%v", err)
		}
	})
}

func TestExecuteUndo_ReverseOrder(t *testing.T) {
	dirA := t.TempDir()
	dirB := t.TempDir()
	dirC := t.TempDir()

	movedPath1 := filepath.Join(dirB, "file1.txt")
	movedPath2 := filepath.Join(dirC, "file2.txt")
	origPath1 := filepath.Join(dirA, "file1.txt")
	origPath2 := filepath.Join(dirA, "file2.txt")

	if err := os.WriteFile(movedPath1, []byte("one"), 0o644); err != nil {
		t.Fatalf("writing moved file1: %v", err)
	}
	if err := os.WriteFile(movedPath2, []byte("two"), 0o644); err != nil {
		t.Fatalf("writing moved file2: %v", err)
	}

	var order []string
	logger := func(format string, args ...interface{}) {
		order = append(order, args[0].(string))
	}

	undoLog := &UndoLog{
		Timestamp: time.Now(),
		Config:    "test",
		Operations: []UndoEntry{
			{From: origPath1, To: movedPath1},
			{From: origPath2, To: movedPath2},
		},
	}

	if err := ExecuteUndo(undoLog, true, logger); err != nil {
		t.Fatalf("ExecuteUndo: %v", err)
	}

	t.Run("both files restored", func(t *testing.T) {
		if _, err := os.Stat(origPath1); err != nil {
			t.Errorf("file1 should be restored: %v", err)
		}
		if _, err := os.Stat(origPath2); err != nil {
			t.Errorf("file2 should be restored: %v", err)
		}
	})

	t.Run("reverse order", func(t *testing.T) {
		if len(order) != 2 {
			t.Fatalf("expected 2 log entries, got %d", len(order))
		}
		if order[0] != movedPath2 {
			t.Errorf("expected first undo for %s, got %s", movedPath2, order[0])
		}
		if order[1] != movedPath1 {
			t.Errorf("expected second undo for %s, got %s", movedPath1, order[1])
		}
	})
}
