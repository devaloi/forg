// Package scanner walks a source directory and collects file metadata.
package scanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileInfo holds metadata about a single file discovered during a scan.
type FileInfo struct {
	Path      string
	Name      string
	Extension string
	Size      int64
	ModTime   time.Time
}

// Options controls the behaviour of a Scanner.
type Options struct {
	// Recursive enables recursive directory traversal. When false, only
	// top-level files in the source directory are collected.
	Recursive bool
	// IncludeHidden includes files whose names start with ".".
	IncludeHidden bool
}

// Scanner walks a directory and collects file metadata according to the
// configured options.
type Scanner struct {
	opts Options
}

// New creates a Scanner with the given options.
func New(opts Options) *Scanner {
	return &Scanner{opts: opts}
}

// Scan walks source and returns metadata for every file that matches the
// scanner's options. Directories themselves are never included in the results.
func (s *Scanner) Scan(source string) ([]FileInfo, error) {
	info, err := os.Stat(source)
	if err != nil {
		return nil, fmt.Errorf("scanner: stat source %q: %w", source, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("scanner: source %q is not a directory", source)
	}

	var files []FileInfo

	if s.opts.Recursive {
		err = filepath.WalkDir(source, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return fmt.Errorf("scanner: walk %q: %w", path, walkErr)
			}

			name := d.Name()

			// Skip hidden entries unless configured otherwise.
			if !s.opts.IncludeHidden && strings.HasPrefix(name, ".") {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}

			if d.IsDir() {
				return nil
			}

			fi, infoErr := d.Info()
			if infoErr != nil {
				return fmt.Errorf("scanner: file info %q: %w", path, infoErr)
			}

			files = append(files, FileInfo{
				Path:      path,
				Name:      name,
				Extension: strings.ToLower(filepath.Ext(name)),
				Size:      fi.Size(),
				ModTime:   fi.ModTime(),
			})
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		entries, readErr := os.ReadDir(source)
		if readErr != nil {
			return nil, fmt.Errorf("scanner: read dir %q: %w", source, readErr)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			name := entry.Name()
			if !s.opts.IncludeHidden && strings.HasPrefix(name, ".") {
				continue
			}

			fi, infoErr := entry.Info()
			if infoErr != nil {
				return nil, fmt.Errorf("scanner: file info %q: %w", name, infoErr)
			}

			files = append(files, FileInfo{
				Path:      filepath.Join(source, name),
				Name:      name,
				Extension: strings.ToLower(filepath.Ext(name)),
				Size:      fi.Size(),
				ModTime:   fi.ModTime(),
			})
		}
	}

	return files, nil
}
