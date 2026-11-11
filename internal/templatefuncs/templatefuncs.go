// Package templatefuncs provides the shared html/template FuncMap attached to
// every template pageforge instantiates.
//
// All functions are pure, total (no panics on bad input), and pipe-friendly:
// the piped value is the last argument.
package templatefuncs

import (
	"html/template"
	"strings"
	"time"
)

// Default returns a fresh FuncMap containing pageforge's built-in template
// functions. A new map is returned on each call so callers can freely mutate
// or extend it without affecting other templates.
func Default() template.FuncMap {
	return template.FuncMap{
		"formatTime":   formatTime,
		"canonicalURL": CanonicalURL,
		"hasPrefix":    hasPrefix,
		"hasSuffix":    hasSuffix,
		"add":          add,
		"sub":          sub,
	}
}

// ParseDate accepts the two date shapes pageforge understands: RFC 3339
// (what pageforge turns an unquoted YAML date into) and date-only YYYY-MM-DD.
// Renderers and templates share it so "is this a date" has one answer.
func ParseDate(s string) (time.Time, bool) {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, true
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t, true
	}
	return time.Time{}, false
}

// CanonicalURL returns the public form of a url_path: rooted, with a
// trailing "index.html" dropped so "/posts/x/index.html" becomes
// "/posts/x/". Feeds, sitemaps and metadata bundles emit this form.
func CanonicalURL(p string) string {
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	if strings.HasSuffix(p, "/index.html") {
		p = strings.TrimSuffix(p, "index.html")
	}
	return p
}

// formatTime parses value with ParseDate and formats it using Go's reference
// layout. On parse failure, value is returned unchanged so a bad metadata
// field does not abort the render.
func formatTime(layout, value string) string {
	t, ok := ParseDate(value)
	if !ok {
		return value
	}
	return t.Format(layout)
}

func hasPrefix(prefix, s string) bool {
	return strings.HasPrefix(s, prefix)
}

func hasSuffix(suffix, s string) bool {
	return strings.HasSuffix(s, suffix)
}

func add(a, b int) int { return a + b }

func sub(a, b int) int { return a - b }
