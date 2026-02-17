package scanner

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

// createFile is a test helper that writes content to the given path,
// creating parent directories as needed.
func createFile(t *testing.T, path string, content string) {
	t.Helper()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("creating directory %q: %v", dir, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("writing file %q: %v", path, err)
	}
}

func TestScan_BasicFiles(t *testing.T) {
	dir := t.TempDir()

	createFile(t, filepath.Join(dir, "test.txt"), "hello")
	createFile(t, filepath.Join(dir, "image.PNG"), "imgdata")
	createFile(t, filepath.Join(dir, "data.csv"), "a,b,c")

	s := New(Options{})
	files, err := s.Scan(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := len(files); got != 3 {
		t.Fatalf("expected 3 files, got %d", got)
	}

	// Sort by name for deterministic assertions.
	sort.Slice(files, func(i, j int) bool {
		return files[i].Name < files[j].Name
	})

	tests := []struct {
		name string
		ext  string
	}{
		{"data.csv", ".csv"},
		{"image.PNG", ".png"},
		{"test.txt", ".txt"},
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := files[i]
			if f.Name != tt.name {
				t.Errorf("Name = %q, want %q", f.Name, tt.name)
			}
			if f.Extension != tt.ext {
				t.Errorf("Extension = %q, want %q", f.Extension, tt.ext)
			}
			wantPath := filepath.Join(dir, tt.name)
			if f.Path != wantPath {
				t.Errorf("Path = %q, want %q", f.Path, wantPath)
			}
		})
	}
}

func TestScan_SkipsHiddenFiles(t *testing.T) {
	dir := t.TempDir()

	createFile(t, filepath.Join(dir, "visible.txt"), "v")
	createFile(t, filepath.Join(dir, ".hidden"), "h")

	s := New(Options{IncludeHidden: false})
	files, err := s.Scan(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := len(files); got != 1 {
		t.Fatalf("expected 1 file, got %d", got)
	}
	if files[0].Name != "visible.txt" {
		t.Errorf("Name = %q, want %q", files[0].Name, "visible.txt")
	}
}

func TestScan_IncludesHiddenFiles(t *testing.T) {
	dir := t.TempDir()

	createFile(t, filepath.Join(dir, "visible.txt"), "v")
	createFile(t, filepath.Join(dir, ".hidden"), "h")

	s := New(Options{IncludeHidden: true})
	files, err := s.Scan(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := len(files); got != 2 {
		t.Fatalf("expected 2 files, got %d", got)
	}

	names := make(map[string]bool)
	for _, f := range files {
		names[f.Name] = true
	}
	for _, want := range []string{"visible.txt", ".hidden"} {
		if !names[want] {
			t.Errorf("expected file %q in results", want)
		}
	}
}

func TestScan_NonRecursive(t *testing.T) {
	dir := t.TempDir()

	createFile(t, filepath.Join(dir, "top.txt"), "top")
	createFile(t, filepath.Join(dir, "sub", "nested.txt"), "nested")

	s := New(Options{Recursive: false})
	files, err := s.Scan(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := len(files); got != 1 {
		t.Fatalf("expected 1 file, got %d", got)
	}
	if files[0].Name != "top.txt" {
		t.Errorf("Name = %q, want %q", files[0].Name, "top.txt")
	}
}

func TestScan_Recursive(t *testing.T) {
	dir := t.TempDir()

	createFile(t, filepath.Join(dir, "top.txt"), "top")
	createFile(t, filepath.Join(dir, "sub", "nested.txt"), "nested")

	s := New(Options{Recursive: true})
	files, err := s.Scan(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := len(files); got != 2 {
		t.Fatalf("expected 2 files, got %d", got)
	}

	names := make(map[string]bool)
	for _, f := range files {
		names[f.Name] = true
	}
	for _, want := range []string{"top.txt", "nested.txt"} {
		if !names[want] {
			t.Errorf("expected file %q in results", want)
		}
	}
}

func TestScan_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	s := New(Options{})
	files, err := s.Scan(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if files == nil {
		// nil is acceptable; convert to empty for length check.
		files = []FileInfo{}
	}
	if got := len(files); got != 0 {
		t.Fatalf("expected 0 files, got %d", got)
	}
}

func TestScan_SkipsDirectories(t *testing.T) {
	dir := t.TempDir()

	createFile(t, filepath.Join(dir, "file.txt"), "content")
	if err := os.MkdirAll(filepath.Join(dir, "subdir"), 0o755); err != nil {
		t.Fatalf("creating subdir: %v", err)
	}

	s := New(Options{})
	files, err := s.Scan(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := len(files); got != 1 {
		t.Fatalf("expected 1 file, got %d", got)
	}
	if files[0].Name != "file.txt" {
		t.Errorf("Name = %q, want %q", files[0].Name, "file.txt")
	}
}

func TestScan_SourceNotExists(t *testing.T) {
	s := New(Options{})
	_, err := s.Scan("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for non-existent source, got nil")
	}
}

func TestScan_SourceIsFile(t *testing.T) {
	dir := t.TempDir()
	fp := filepath.Join(dir, "afile.txt")
	createFile(t, fp, "data")

	s := New(Options{})
	_, err := s.Scan(fp)
	if err == nil {
		t.Fatal("expected error when source is a file, got nil")
	}
}

func TestScan_FileMetadata(t *testing.T) {
	dir := t.TempDir()

	content := "hello, world!"
	createFile(t, filepath.Join(dir, "meta.txt"), content)

	s := New(Options{})
	files, err := s.Scan(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := len(files); got != 1 {
		t.Fatalf("expected 1 file, got %d", got)
	}

	f := files[0]
	wantSize := int64(len(content))
	if f.Size != wantSize {
		t.Errorf("Size = %d, want %d", f.Size, wantSize)
	}
	if f.ModTime.IsZero() {
		t.Error("ModTime should not be zero")
	}
	if f.Name != "meta.txt" {
		t.Errorf("Name = %q, want %q", f.Name, "meta.txt")
	}
	if f.Extension != ".txt" {
		t.Errorf("Extension = %q, want %q", f.Extension, ".txt")
	}
}
