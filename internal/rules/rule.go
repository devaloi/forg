// Package rules implements the file matching engine that evaluates files
// against configured rules to determine their destination directories.
package rules

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/jasonaloi/forg/internal/scanner"
)

// Matcher is the interface that wraps the Match method.
//
// Match reports whether the given file satisfies the matcher's criteria.
type Matcher interface {
	Match(file scanner.FileInfo) bool
}

// ExtensionMatcher matches files whose extension (case-insensitive) appears
// in the Extensions list.
type ExtensionMatcher struct {
	Extensions []string
}

// Match returns true if the file's extension matches any of the configured extensions.
func (m ExtensionMatcher) Match(file scanner.FileInfo) bool {
	ext := strings.ToLower(file.Extension)
	for _, e := range m.Extensions {
		if strings.ToLower(e) == ext {
			return true
		}
	}
	return false
}

// PatternMatcher matches files whose name matches a filepath.Match glob pattern.
type PatternMatcher struct {
	Pattern string
}

// Match returns true if the file's name matches the glob pattern.
func (m PatternMatcher) Match(file scanner.FileInfo) bool {
	matched, err := filepath.Match(m.Pattern, file.Name)
	if err != nil {
		return false
	}
	return matched
}

// MinSizeMatcher matches files whose size is at least MinBytes bytes.
type MinSizeMatcher struct {
	MinBytes int64
}

// Match returns true if the file's size is greater than or equal to MinBytes.
func (m MinSizeMatcher) Match(file scanner.FileInfo) bool {
	return file.Size >= m.MinBytes
}

// MaxSizeMatcher matches files whose size is at most MaxBytes bytes.
type MaxSizeMatcher struct {
	MaxBytes int64
}

// Match returns true if the file's size is less than or equal to MaxBytes.
func (m MaxSizeMatcher) Match(file scanner.FileInfo) bool {
	return file.Size <= m.MaxBytes
}

// OlderThanMatcher matches files whose modification time is before
// the current time minus the configured duration in seconds.
type OlderThanMatcher struct {
	Seconds int64
}

// Match returns true if the file's modification time is older than the threshold.
func (m OlderThanMatcher) Match(file scanner.FileInfo) bool {
	threshold := time.Now().Add(-time.Duration(m.Seconds) * time.Second)
	return file.ModTime.Before(threshold)
}

// NewerThanMatcher matches files whose modification time is after
// the current time minus the configured duration in seconds.
type NewerThanMatcher struct {
	Seconds int64
}

// Match returns true if the file's modification time is newer than the threshold.
func (m NewerThanMatcher) Match(file scanner.FileInfo) bool {
	threshold := time.Now().Add(-time.Duration(m.Seconds) * time.Second)
	return file.ModTime.After(threshold)
}

// Rule represents a named organization rule that maps matching files to a
// destination directory.
type Rule struct {
	Name        string
	Destination string
	Matchers    []Matcher
}

// Match returns true only if all of the rule's matchers match the given file.
// A rule with no matchers never matches.
func (r *Rule) Match(file scanner.FileInfo) bool {
	if len(r.Matchers) == 0 {
		return false
	}
	for _, m := range r.Matchers {
		if !m.Match(file) {
			return false
		}
	}
	return true
}
