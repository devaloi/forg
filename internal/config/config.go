// Package config handles parsing and validation of .forg.yaml configuration files.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the top-level forg configuration.
type Config struct {
	Source   string       `yaml:"source"`
	Conflict string       `yaml:"conflict"`
	Rules    []RuleConfig `yaml:"rules"`
}

// RuleConfig represents a single organization rule.
type RuleConfig struct {
	Name        string      `yaml:"name"`
	Match       MatchConfig `yaml:"match"`
	Destination string      `yaml:"destination"`
}

// MatchConfig defines the criteria for matching files in a rule.
type MatchConfig struct {
	Extensions []string `yaml:"extensions,omitempty"`
	Pattern    string   `yaml:"pattern,omitempty"`
	MinSize    string   `yaml:"min_size,omitempty"`
	MaxSize    string   `yaml:"max_size,omitempty"`
	OlderThan  string   `yaml:"older_than,omitempty"`
	NewerThan  string   `yaml:"newer_than,omitempty"`
}

// sizePattern matches size strings like "100MB", "1.5GB", "500KB".
var sizePattern = regexp.MustCompile(`(?i)^(\d+(?:\.\d+)?)\s*(B|KB|MB|GB|TB)$`)

// durationPattern matches duration strings like "30d", "2w", "6m", "1y".
var durationPattern = regexp.MustCompile(`(?i)^(\d+)\s*(d|w|m|y)$`)

// sizeMultipliers maps size unit suffixes to their byte multipliers.
var sizeMultipliers = map[string]int64{
	"b":  1,
	"kb": 1024,
	"mb": 1024 * 1024,
	"gb": 1024 * 1024 * 1024,
	"tb": 1024 * 1024 * 1024 * 1024,
}

// durationMultipliers maps duration unit suffixes to their second multipliers.
var durationMultipliers = map[string]int64{
	"d": 86400,
	"w": 604800,
	"m": 2592000,
	"y": 31536000,
}

// Load reads and parses a .forg.yaml configuration file from the given path.
func Load(path string) (*Config, error) {
	expanded, err := ExpandPath(path)
	if err != nil {
		return nil, fmt.Errorf("expanding config path: %w", err)
	}

	data, err := os.ReadFile(expanded)
	if err != nil {
		return nil, fmt.Errorf("reading config file %s: %w", expanded, err)
	}

	cfg, err := Parse(data)
	if err != nil {
		return nil, fmt.Errorf("parsing config file %s: %w", expanded, err)
	}

	return cfg, nil
}

// Parse unmarshals YAML data into a Config and validates it.
func Parse(data []byte) (*Config, error) {
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("unmarshaling YAML: %w", err)
	}

	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("validating config: %w", err)
	}

	return &cfg, nil
}

// validate checks that the config is well-formed.
func validate(cfg *Config) error {
	if cfg.Source == "" {
		return fmt.Errorf("source directory is required")
	}

	srcExpanded, err := ExpandPath(cfg.Source)
	if err != nil {
		return fmt.Errorf("expanding source path: %w", err)
	}

	info, err := os.Stat(srcExpanded)
	if err != nil {
		return fmt.Errorf("source directory %s: %w", srcExpanded, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("source path %s is not a directory", srcExpanded)
	}

	if cfg.Conflict != "" {
		valid := map[string]bool{"skip": true, "rename": true, "overwrite": true}
		if !valid[cfg.Conflict] {
			return fmt.Errorf("invalid conflict strategy %q: must be skip, rename, or overwrite", cfg.Conflict)
		}
	}

	if len(cfg.Rules) == 0 {
		return fmt.Errorf("at least one rule is required")
	}

	for i, rule := range cfg.Rules {
		if err := validateRule(i, rule); err != nil {
			return err
		}
	}

	return nil
}

// validateRule checks that a single rule has all required fields and valid values.
func validateRule(index int, rule RuleConfig) error {
	if rule.Name == "" {
		return fmt.Errorf("rule %d: name is required", index)
	}

	if rule.Destination == "" {
		return fmt.Errorf("rule %q: destination is required", rule.Name)
	}

	hasMatch := len(rule.Match.Extensions) > 0 ||
		rule.Match.Pattern != "" ||
		rule.Match.MinSize != "" ||
		rule.Match.MaxSize != "" ||
		rule.Match.OlderThan != "" ||
		rule.Match.NewerThan != ""

	if !hasMatch {
		return fmt.Errorf("rule %q: at least one match criterion is required", rule.Name)
	}

	if rule.Match.Pattern != "" {
		if _, err := regexp.Compile(rule.Match.Pattern); err != nil {
			return fmt.Errorf("rule %q: invalid pattern %q: %w", rule.Name, rule.Match.Pattern, err)
		}
	}

	if rule.Match.MinSize != "" {
		if _, err := ParseSize(rule.Match.MinSize); err != nil {
			return fmt.Errorf("rule %q: invalid min_size: %w", rule.Name, err)
		}
	}

	if rule.Match.MaxSize != "" {
		if _, err := ParseSize(rule.Match.MaxSize); err != nil {
			return fmt.Errorf("rule %q: invalid max_size: %w", rule.Name, err)
		}
	}

	if rule.Match.OlderThan != "" {
		if _, err := ParseDuration(rule.Match.OlderThan); err != nil {
			return fmt.Errorf("rule %q: invalid older_than: %w", rule.Name, err)
		}
	}

	if rule.Match.NewerThan != "" {
		if _, err := ParseDuration(rule.Match.NewerThan); err != nil {
			return fmt.Errorf("rule %q: invalid newer_than: %w", rule.Name, err)
		}
	}

	return nil
}

// ParseSize converts a human-readable size string (e.g. "100MB", "1.5GB") to bytes.
func ParseSize(s string) (int64, error) {
	matches := sizePattern.FindStringSubmatch(strings.TrimSpace(s))
	if matches == nil {
		return 0, fmt.Errorf("invalid size format %q: expected number followed by B, KB, MB, GB, or TB", s)
	}

	value, err := strconv.ParseFloat(matches[1], 64)
	if err != nil {
		return 0, fmt.Errorf("parsing size number %q: %w", matches[1], err)
	}

	unit := strings.ToLower(matches[2])
	multiplier, ok := sizeMultipliers[unit]
	if !ok {
		return 0, fmt.Errorf("unknown size unit %q", matches[2])
	}

	return int64(value * float64(multiplier)), nil
}

// ParseDuration converts a human-readable duration string (e.g. "30d", "2w", "6m", "1y") to seconds.
func ParseDuration(s string) (int64, error) {
	matches := durationPattern.FindStringSubmatch(strings.TrimSpace(s))
	if matches == nil {
		return 0, fmt.Errorf("invalid duration format %q: expected number followed by d, w, m, or y", s)
	}

	value, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, fmt.Errorf("parsing duration number %q: %w", matches[1], err)
	}

	unit := strings.ToLower(matches[2])
	multiplier, ok := durationMultipliers[unit]
	if !ok {
		return 0, fmt.Errorf("unknown duration unit %q", matches[2])
	}

	return value * multiplier, nil
}

// ExpandPath expands a leading ~ in a path to the user's home directory.
func ExpandPath(path string) (string, error) {
	if path == "" {
		return path, nil
	}

	if strings.HasPrefix(path, "~/") || path == "~" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("resolving home directory: %w", err)
		}
		if path == "~" {
			return home, nil
		}
		return filepath.Join(home, path[2:]), nil
	}

	return path, nil
}

// SampleConfig returns a sample .forg.yaml configuration string.
func SampleConfig() string {
	return "# forg configuration file\n" +
		"# See https://github.com/jasonaloi/forg for documentation\n" +
		"\n" +
		"source: ~/Downloads\n" +
		"conflict: rename\n" +
		"\n" +
		"rules:\n" +
		"  - name: images\n" +
		"    match:\n" +
		"      extensions:\n" +
		"        - .jpg\n" +
		"        - .jpeg\n" +
		"        - .png\n" +
		"        - .gif\n" +
		"        - .webp\n" +
		"    destination: ~/Pictures/Sorted\n" +
		"\n" +
		"  - name: documents\n" +
		"    match:\n" +
		"      extensions:\n" +
		"        - .pdf\n" +
		"        - .doc\n" +
		"        - .docx\n" +
		"        - .txt\n" +
		"    destination: ~/Documents/Sorted\n" +
		"\n" +
		"  - name: large-videos\n" +
		"    match:\n" +
		"      extensions:\n" +
		"        - .mp4\n" +
		"        - .mov\n" +
		"        - .avi\n" +
		"      min_size: 100MB\n" +
		"    destination: ~/Videos/Large\n" +
		"\n" +
		"  - name: old-archives\n" +
		"    match:\n" +
		"      extensions:\n" +
		"        - .zip\n" +
		"        - .tar.gz\n" +
		"        - .rar\n" +
		"      older_than: 30d\n" +
		"    destination: ~/Archives/Old\n"
}
