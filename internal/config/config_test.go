package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseSize(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		tests := []struct {
			input    string
			expected int64
		}{
			{"100MB", 104857600},
			{"1KB", 1024},
			{"1GB", 1073741824},
			{"0B", 0},
			{"1.5GB", 1610612736},
			{"100mb", 104857600},
		}

		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				got, err := ParseSize(tt.input)
				if err != nil {
					t.Errorf("ParseSize(%q) unexpected error: %v", tt.input, err)
				}
				if got != tt.expected {
					t.Errorf("ParseSize(%q) = %d, want %d", tt.input, got, tt.expected)
				}
			})
		}
	})

	t.Run("errors", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"empty string", ""},
			{"letters only", "abc"},
			{"number only", "100"},
			{"unit only", "MB"},
			{"invalid unit", "100XB"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := ParseSize(tt.input)
				if err == nil {
					t.Errorf("ParseSize(%q) expected error, got nil", tt.input)
				}
			})
		}
	})
}

func TestParseDuration(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		tests := []struct {
			input    string
			expected int64
		}{
			{"30d", 2592000},
			{"2w", 1209600},
			{"6m", 15552000},
			{"1y", 31536000},
		}

		for _, tt := range tests {
			t.Run(tt.input, func(t *testing.T) {
				got, err := ParseDuration(tt.input)
				if err != nil {
					t.Errorf("ParseDuration(%q) unexpected error: %v", tt.input, err)
				}
				if got != tt.expected {
					t.Errorf("ParseDuration(%q) = %d, want %d", tt.input, got, tt.expected)
				}
			})
		}
	})

	t.Run("errors", func(t *testing.T) {
		tests := []struct {
			name  string
			input string
		}{
			{"empty string", ""},
			{"letters only", "abc"},
			{"number only", "30"},
			{"unit only", "d"},
			{"invalid unit", "30x"},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := ParseDuration(tt.input)
				if err == nil {
					t.Errorf("ParseDuration(%q) expected error, got nil", tt.input)
				}
			})
		}
	})
}

func TestExpandPath(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home directory: %v", err)
	}

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"tilde prefix", "~/test", filepath.Join(home, "test")},
		{"absolute path", "/absolute/path", "/absolute/path"},
		{"relative path", "relative/path", "relative/path"},
		{"empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExpandPath(tt.input)
			if err != nil {
				t.Errorf("ExpandPath(%q) unexpected error: %v", tt.input, err)
			}
			if got != tt.expected {
				t.Errorf("ExpandPath(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParse_ValidConfig(t *testing.T) {
	srcDir := t.TempDir()

	yamlData := fmt.Sprintf("source: %s\nconflict: rename\nrules:\n  - name: images\n    match:\n      extensions:\n        - .jpg\n        - .png\n    destination: /tmp/sorted/images\n", srcDir)

	cfg, err := Parse([]byte(yamlData))
	if err != nil {
		t.Fatalf("Parse() unexpected error: %v", err)
	}
	if cfg.Source != srcDir {
		t.Errorf("Source = %q, want %q", cfg.Source, srcDir)
	}
	if cfg.Conflict != "rename" {
		t.Errorf("Conflict = %q, want %q", cfg.Conflict, "rename")
	}
	if len(cfg.Rules) != 1 {
		t.Fatalf("len(Rules) = %d, want 1", len(cfg.Rules))
	}
	if cfg.Rules[0].Name != "images" {
		t.Errorf("Rules[0].Name = %q, want %q", cfg.Rules[0].Name, "images")
	}
	if cfg.Rules[0].Destination != "/tmp/sorted/images" {
		t.Errorf("Rules[0].Destination = %q, want %q", cfg.Rules[0].Destination, "/tmp/sorted/images")
	}
	if len(cfg.Rules[0].Match.Extensions) != 2 {
		t.Errorf("len(Rules[0].Match.Extensions) = %d, want 2", len(cfg.Rules[0].Match.Extensions))
	}
}

func TestParse_Errors(t *testing.T) {
	srcDir := t.TempDir()

	tmpFile := filepath.Join(t.TempDir(), "notadir.txt")
	if err := os.WriteFile(tmpFile, []byte("hello"), 0o600); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	tests := []struct {
		name      string
		yaml      string
		wantError string
	}{
		{
			name:      "missing source",
			yaml:      "rules:\n  - name: test\n    match:\n      extensions: [.jpg]\n    destination: /tmp/out\n",
			wantError: "source directory is required",
		},
		{
			name:      "source does not exist",
			yaml:      "source: /nonexistent/path/that/does/not/exist\nrules:\n  - name: test\n    match:\n      extensions: [.jpg]\n    destination: /tmp/out\n",
			wantError: "source directory",
		},
		{
			name:      "source is a file not dir",
			yaml:      fmt.Sprintf("source: %s\nrules:\n  - name: test\n    match:\n      extensions: [.jpg]\n    destination: /tmp/out\n", tmpFile),
			wantError: "is not a directory",
		},
		{
			name:      "no rules",
			yaml:      fmt.Sprintf("source: %s\n", srcDir),
			wantError: "at least one rule is required",
		},
		{
			name:      "rule missing name",
			yaml:      fmt.Sprintf("source: %s\nrules:\n  - match:\n      extensions: [.jpg]\n    destination: /tmp/out\n", srcDir),
			wantError: "name is required",
		},
		{
			name:      "rule missing destination",
			yaml:      fmt.Sprintf("source: %s\nrules:\n  - name: test\n    match:\n      extensions: [.jpg]\n", srcDir),
			wantError: "destination is required",
		},
		{
			name:      "rule missing match criteria",
			yaml:      fmt.Sprintf("source: %s\nrules:\n  - name: test\n    match: {}\n    destination: /tmp/out\n", srcDir),
			wantError: "at least one match criterion is required",
		},
		{
			name:      "invalid conflict strategy",
			yaml:      fmt.Sprintf("source: %s\nconflict: invalid\nrules:\n  - name: test\n    match:\n      extensions: [.jpg]\n    destination: /tmp/out\n", srcDir),
			wantError: "invalid conflict strategy",
		},
		{
			name:      "invalid min_size",
			yaml:      fmt.Sprintf("source: %s\nrules:\n  - name: test\n    match:\n      min_size: badsize\n    destination: /tmp/out\n", srcDir),
			wantError: "invalid min_size",
		},
		{
			name:      "invalid older_than",
			yaml:      fmt.Sprintf("source: %s\nrules:\n  - name: test\n    match:\n      older_than: badtime\n    destination: /tmp/out\n", srcDir),
			wantError: "invalid older_than",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Parse([]byte(tt.yaml))
			if err == nil {
				t.Fatalf("Parse() expected error containing %q, got nil", tt.wantError)
			}
			if !strings.Contains(err.Error(), tt.wantError) {
				t.Errorf("Parse() error = %q, want it to contain %q", err.Error(), tt.wantError)
			}
		})
	}
}

func TestLoad_FileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Fatal("Load() expected error for missing file, got nil")
	}
}

func TestSampleConfig(t *testing.T) {
	sample := SampleConfig()
	if sample == "" {
		t.Fatal("SampleConfig() returned empty string")
	}
	if !strings.Contains(sample, "source:") {
		t.Error("SampleConfig() missing 'source:' field")
	}
	if !strings.Contains(sample, "rules:") {
		t.Error("SampleConfig() missing 'rules:' field")
	}
}
