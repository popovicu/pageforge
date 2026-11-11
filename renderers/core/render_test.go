package core

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRenderTemplateToBytes_AppliesFuncMap renders a fixture that exercises
// every function in templatefuncs.Default(), proving the shared FuncMap is
// attached to every template constructed via the core render helper.
func TestRenderTemplateToBytes_AppliesFuncMap(t *testing.T) {
	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "page.tmpl")

	// Two pipe forms per function are exercised:
	//   - formatTime via pipe      ({{.Date | formatTime ...}})
	//   - hasPrefix / hasSuffix via pipe
	//   - add / sub called directly (they take ints, not a piped string)
	fixture := `URL_PREFIX={{if .URL | hasPrefix "/series/"}}yes{{else}}no{{end}};` +
		`URL_SUFFIX={{if .URL | hasSuffix ".html"}}yes{{else}}no{{end}};` +
		`DATE={{.Date | formatTime "January 2, 2006"}};` +
		`ADD={{add 2 3}};` +
		`SUB={{sub 10 4}}`

	if err := os.WriteFile(tmplPath, []byte(fixture), 0o644); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}

	ctx := struct {
		URL  string
		Date string
	}{
		URL:  "/series/foo.html",
		Date: "2026-04-23T10:30:00Z",
	}

	got, err := RenderTemplateToBytes([]string{tmplPath}, ctx)
	if err != nil {
		t.Fatalf("RenderTemplateToBytes: %v", err)
	}

	want := "URL_PREFIX=yes;URL_SUFFIX=yes;DATE=April 23, 2026;ADD=5;SUB=6"
	if strings.TrimSpace(string(got)) != want {
		t.Errorf("rendered output mismatch:\n got: %q\nwant: %q", string(got), want)
	}
}

// TestRenderTemplateToBytes_FormatTimeFallback confirms formatTime's total
// behavior survives the end-to-end render path: a malformed date does not
// abort rendering, it passes through unchanged.
func TestRenderTemplateToBytes_FormatTimeFallback(t *testing.T) {
	dir := t.TempDir()
	tmplPath := filepath.Join(dir, "page.tmpl")

	if err := os.WriteFile(tmplPath, []byte(`{{.Date | formatTime "2006"}}`), 0o644); err != nil {
		t.Fatalf("writing fixture: %v", err)
	}

	ctx := struct{ Date string }{Date: "not a date"}

	got, err := RenderTemplateToBytes([]string{tmplPath}, ctx)
	if err != nil {
		t.Fatalf("RenderTemplateToBytes: %v", err)
	}
	if string(got) != "not a date" {
		t.Errorf("formatTime fallback: got %q, want %q", string(got), "not a date")
	}
}
