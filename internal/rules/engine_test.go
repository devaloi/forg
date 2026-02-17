package rules

import (
	"testing"
	"time"

	"github.com/jasonaloi/forg/internal/config"
	"github.com/jasonaloi/forg/internal/scanner"
)

func TestExtensionMatcher(t *testing.T) {
	tests := []struct {
		name       string
		extensions []string
		file       scanner.FileInfo
		want       bool
	}{
		{
			name:       "match jpg in list",
			extensions: []string{".jpg", ".png"},
			file:       scanner.FileInfo{Name: "photo.jpg", Extension: ".jpg"},
			want:       true,
		},
		{
			name:       "case insensitive match",
			extensions: []string{".pdf"},
			file:       scanner.FileInfo{Name: "DOC.PDF", Extension: ".PDF"},
			want:       true,
		},
		{
			name:       "no match",
			extensions: []string{".jpg", ".png"},
			file:       scanner.FileInfo{Name: "notes.txt", Extension: ".txt"},
			want:       false,
		},
		{
			name:       "no extension",
			extensions: []string{".jpg", ".png"},
			file:       scanner.FileInfo{Name: "Makefile", Extension: ""},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := ExtensionMatcher{Extensions: tt.extensions}
			got := m.Match(tt.file)
			if got != tt.want {
				t.Errorf("ExtensionMatcher.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPatternMatcher(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		file    scanner.FileInfo
		want    bool
	}{
		{
			name:    "Screenshot* matches Screenshot_2024.png",
			pattern: "Screenshot*",
			file:    scanner.FileInfo{Name: "Screenshot_2024.png"},
			want:    true,
		},
		{
			name:    "Screenshot* does not match photo.png",
			pattern: "Screenshot*",
			file:    scanner.FileInfo{Name: "photo.png"},
			want:    false,
		},
		{
			name:    "*.log matches app.log",
			pattern: "*.log",
			file:    scanner.FileInfo{Name: "app.log"},
			want:    true,
		},
		{
			name:    "*.log does not match app.txt",
			pattern: "*.log",
			file:    scanner.FileInfo{Name: "app.txt"},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := PatternMatcher{Pattern: tt.pattern}
			got := m.Match(tt.file)
			if got != tt.want {
				t.Errorf("PatternMatcher.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMinSizeMatcher(t *testing.T) {
	tests := []struct {
		name     string
		minBytes int64
		fileSize int64
		want     bool
	}{
		{
			name:     "file larger than min",
			minBytes: 100,
			fileSize: 200,
			want:     true,
		},
		{
			name:     "file equal to min",
			minBytes: 100,
			fileSize: 100,
			want:     true,
		},
		{
			name:     "file smaller than min",
			minBytes: 100,
			fileSize: 50,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MinSizeMatcher{MinBytes: tt.minBytes}
			got := m.Match(scanner.FileInfo{Size: tt.fileSize})
			if got != tt.want {
				t.Errorf("MinSizeMatcher.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMaxSizeMatcher(t *testing.T) {
	tests := []struct {
		name     string
		maxBytes int64
		fileSize int64
		want     bool
	}{
		{
			name:     "file smaller than max",
			maxBytes: 100,
			fileSize: 50,
			want:     true,
		},
		{
			name:     "file equal to max",
			maxBytes: 100,
			fileSize: 100,
			want:     true,
		},
		{
			name:     "file larger than max",
			maxBytes: 100,
			fileSize: 200,
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := MaxSizeMatcher{MaxBytes: tt.maxBytes}
			got := m.Match(scanner.FileInfo{Size: tt.fileSize})
			if got != tt.want {
				t.Errorf("MaxSizeMatcher.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOlderThanMatcher(t *testing.T) {
	tests := []struct {
		name    string
		seconds int64
		modTime time.Time
		want    bool
	}{
		{
			name:    "file modified 60 days ago, threshold 30 days",
			seconds: 2592000,
			modTime: time.Now().Add(-60 * 24 * time.Hour),
			want:    true,
		},
		{
			name:    "file modified 1 day ago, threshold 30 days",
			seconds: 2592000,
			modTime: time.Now().Add(-1 * 24 * time.Hour),
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := OlderThanMatcher{Seconds: tt.seconds}
			got := m.Match(scanner.FileInfo{ModTime: tt.modTime})
			if got != tt.want {
				t.Errorf("OlderThanMatcher.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewerThanMatcher(t *testing.T) {
	tests := []struct {
		name    string
		seconds int64
		modTime time.Time
		want    bool
	}{
		{
			name:    "file modified 1 day ago, threshold 30 days",
			seconds: 2592000,
			modTime: time.Now().Add(-1 * 24 * time.Hour),
			want:    true,
		},
		{
			name:    "file modified 60 days ago, threshold 30 days",
			seconds: 2592000,
			modTime: time.Now().Add(-60 * 24 * time.Hour),
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewerThanMatcher{Seconds: tt.seconds}
			got := m.Match(scanner.FileInfo{ModTime: tt.modTime})
			if got != tt.want {
				t.Errorf("NewerThanMatcher.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRule_ANDLogic(t *testing.T) {
	rule := Rule{
		Name:        "images-large",
		Destination: "/dest",
		Matchers: []Matcher{
			ExtensionMatcher{Extensions: []string{".jpg"}},
			MinSizeMatcher{MinBytes: 100},
		},
	}

	tests := []struct {
		name string
		file scanner.FileInfo
		want bool
	}{
		{
			name: "both matchers match",
			file: scanner.FileInfo{Name: "photo.jpg", Extension: ".jpg", Size: 200},
			want: true,
		},
		{
			name: "extension matches but size too small",
			file: scanner.FileInfo{Name: "tiny.jpg", Extension: ".jpg", Size: 50},
			want: false,
		},
		{
			name: "size matches but wrong extension",
			file: scanner.FileInfo{Name: "photo.png", Extension: ".png", Size: 200},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rule.Match(tt.file)
			if got != tt.want {
				t.Errorf("Rule.Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRule_NoMatchers(t *testing.T) {
	rule := Rule{
		Name:        "empty",
		Destination: "/dest",
		Matchers:    []Matcher{},
	}

	file := scanner.FileInfo{Name: "anything.txt", Extension: ".txt", Size: 100}
	if rule.Match(file) {
		t.Error("Rule with no matchers should never match")
	}
}

func TestEngine_FirstMatchWins(t *testing.T) {
	cfgRules := []config.RuleConfig{
		{
			Name:        "rule1",
			Match:       config.MatchConfig{Extensions: []string{".jpg"}},
			Destination: "/dest1",
		},
		{
			Name:        "rule2",
			Match:       config.MatchConfig{Extensions: []string{".jpg"}},
			Destination: "/dest2",
		},
	}

	engine, err := NewEngine(cfgRules)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	file := scanner.FileInfo{Name: "photo.jpg", Extension: ".jpg"}
	matched := engine.Match(file)
	if matched == nil {
		t.Fatal("expected a match, got nil")
	}
	if matched.Name != "rule1" {
		t.Errorf("expected rule1 to win, got %q", matched.Name)
	}
}

func TestEngine_NoMatch(t *testing.T) {
	cfgRules := []config.RuleConfig{
		{
			Name:        "images",
			Match:       config.MatchConfig{Extensions: []string{".jpg"}},
			Destination: "/images",
		},
	}

	engine, err := NewEngine(cfgRules)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	file := scanner.FileInfo{Name: "notes.txt", Extension: ".txt"}
	matched := engine.Match(file)
	if matched != nil {
		t.Errorf("expected nil match, got rule %q", matched.Name)
	}
}

func TestNewEngine_ValidConfig(t *testing.T) {
	cfgRules := []config.RuleConfig{
		{
			Name:        "images",
			Match:       config.MatchConfig{Extensions: []string{".jpg", ".png"}},
			Destination: "/images",
		},
		{
			Name:        "docs",
			Match:       config.MatchConfig{Extensions: []string{".pdf", ".docx"}},
			Destination: "/docs",
		},
	}

	engine, err := NewEngine(cfgRules)
	if err != nil {
		t.Fatalf("NewEngine() error = %v", err)
	}

	rules := engine.Rules()
	if len(rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(rules))
	}
}

func TestNewEngine_InvalidSize(t *testing.T) {
	cfgRules := []config.RuleConfig{
		{
			Name:        "bad-rule",
			Match:       config.MatchConfig{MinSize: "abc"},
			Destination: "/dest",
		},
	}

	_, err := NewEngine(cfgRules)
	if err == nil {
		t.Error("expected error for invalid min_size, got nil")
	}
}
