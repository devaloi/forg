package rules

import (
	"fmt"

	"github.com/jasonaloi/forg/internal/config"
	"github.com/jasonaloi/forg/internal/scanner"
)

// Engine evaluates files against an ordered set of rules and returns
// the first matching rule for each file.
type Engine struct {
	rules []Rule
}

// NewEngine creates an Engine from the given configuration rules. It returns
// an error if any rule cannot be built (e.g. invalid size or duration format).
func NewEngine(cfgRules []config.RuleConfig) (*Engine, error) {
	rules := make([]Rule, 0, len(cfgRules))
	for i, cr := range cfgRules {
		r, err := buildRule(cr)
		if err != nil {
			return nil, fmt.Errorf("building rule %d (%q): %w", i, cr.Name, err)
		}
		rules = append(rules, r)
	}
	return &Engine{rules: rules}, nil
}

// Match returns the first rule that matches the given file, or nil if no
// rule matches.
func (e *Engine) Match(file scanner.FileInfo) *Rule {
	for i := range e.rules {
		if e.rules[i].Match(file) {
			return &e.rules[i]
		}
	}
	return nil
}

// Rules returns all rules loaded into the engine.
func (e *Engine) Rules() []Rule {
	return e.rules
}

// buildRule converts a config.RuleConfig into a Rule by creating the
// appropriate matchers for each configured match criterion.
func buildRule(cr config.RuleConfig) (Rule, error) {
	dest, err := config.ExpandPath(cr.Destination)
	if err != nil {
		return Rule{}, fmt.Errorf("expanding destination path: %w", err)
	}

	r := Rule{
		Name:        cr.Name,
		Destination: dest,
	}

	if len(cr.Match.Extensions) > 0 {
		r.Matchers = append(r.Matchers, ExtensionMatcher{
			Extensions: cr.Match.Extensions,
		})
	}

	if cr.Match.Pattern != "" {
		r.Matchers = append(r.Matchers, PatternMatcher{
			Pattern: cr.Match.Pattern,
		})
	}

	if cr.Match.MinSize != "" {
		bytes, err := config.ParseSize(cr.Match.MinSize)
		if err != nil {
			return Rule{}, fmt.Errorf("parsing min_size: %w", err)
		}
		r.Matchers = append(r.Matchers, MinSizeMatcher{MinBytes: bytes})
	}

	if cr.Match.MaxSize != "" {
		bytes, err := config.ParseSize(cr.Match.MaxSize)
		if err != nil {
			return Rule{}, fmt.Errorf("parsing max_size: %w", err)
		}
		r.Matchers = append(r.Matchers, MaxSizeMatcher{MaxBytes: bytes})
	}

	if cr.Match.OlderThan != "" {
		secs, err := config.ParseDuration(cr.Match.OlderThan)
		if err != nil {
			return Rule{}, fmt.Errorf("parsing older_than: %w", err)
		}
		r.Matchers = append(r.Matchers, OlderThanMatcher{Seconds: secs})
	}

	if cr.Match.NewerThan != "" {
		secs, err := config.ParseDuration(cr.Match.NewerThan)
		if err != nil {
			return Rule{}, fmt.Errorf("parsing newer_than: %w", err)
		}
		r.Matchers = append(r.Matchers, NewerThanMatcher{Seconds: secs})
	}

	return r, nil
}
