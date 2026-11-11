package main

import (
	"encoding/json"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/popovicu/pageforge/internal/testutil"
)

type rendered struct {
	html, content string
	meta          map[string]any
}

// render runs the markdown renderer on src through a template that prints
// the date and the content, with an optional site config.
func render(t *testing.T, src, siteConfig string) rendered {
	t.Helper()
	dir := t.TempDir()
	out := filepath.Join(dir, "page.html")
	metaOut := filepath.Join(dir, "page.meta.json")
	contentOut := filepath.Join(dir, "page.content.html")
	args := []string{
		"--src", testutil.Write(t, dir, "page.md", src),
		"--templates", testutil.Write(t, dir, "page.tmpl", `[{{.Date}}]{{.Content}}`),
		"--url-path", "/posts/p/index.html",
		"--output", out,
		"--meta-output", metaOut,
		"--content-output", contentOut,
	}
	if siteConfig != "" {
		args = append(args, "--site-config", testutil.Write(t, dir, "site.yaml", siteConfig))
	}
	if err := run(args); err != nil {
		t.Fatal(err)
	}
	var meta map[string]any
	if err := json.Unmarshal([]byte(testutil.Read(t, metaOut)), &meta); err != nil {
		t.Fatal(err)
	}
	return rendered{testutil.Read(t, out), testutil.Read(t, contentOut), meta}
}

const tocDoc = `## Table of contents

## First

### Sub

## Second
`

func TestTOC(t *testing.T) {
	got := render(t, tocDoc, "").content
	for _, want := range []string{
		`<h2 id="table-of-contents">Table of contents</h2>`,
		`<details class="pageforge-toc"><summary>Open Table of contents</summary>`,
		"<li><a href=\"#first\">First</a>\n<ul>\n<li><a href=\"#sub\">Sub</a></li>\n</ul>\n</li>",
		`<li><a href="#second">Second</a></li>`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestTOCOptOut(t *testing.T) {
	got := render(t, "---\nauto_toc: false\n---\n"+tocDoc, "").content
	if strings.Contains(got, "pageforge-toc") {
		t.Errorf("auto_toc: false still injected a TOC:\n%s", got)
	}
}

func TestTOCMarkerFromSiteConfig(t *testing.T) {
	got := render(t, "## Contents\n\n## A\n", "toc_marker: Contents\n").content
	if !strings.Contains(got, "<summary>Open Contents</summary>") {
		t.Errorf("site_config toc_marker was not honored:\n%s", got)
	}
}

func TestTOCOrphanH3(t *testing.T) {
	got := render(t, "## Table of contents\n\n### Early\n\n## Later\n", "").content
	if !strings.Contains(got, "<li>\n<ul>\n<li><a href=\"#early\">Early</a></li>") {
		t.Errorf("H3 before any H2 was not nested under a synthetic item:\n%s", got)
	}
}

func TestMarkdownFeatures(t *testing.T) {
	src := "| a | b |\n|---|---|\n| 1 | 2 |\n\nSee https://example.com now.\n\n<div class=\"raw\">kept</div>\n\n```go\nfunc f() {}\n```\n"
	got := render(t, src, "").content
	for _, want := range []string{
		"<table>",
		`<a href="https://example.com">https://example.com</a>`,
		`<div class="raw">kept</div>`,
		`class="chroma"`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestSidecarAndContentOutputs(t *testing.T) {
	r := render(t, "---\ntitle: T\ndate: 2026-04-15\n---\nbody\n", "site_name: S\ntitle: Site\n")
	// The page sees the merged metadata with the date as an RFC 3339 string.
	if !strings.HasPrefix(r.html, "[2026-04-15T00:00:00Z]<p>body</p>") {
		t.Errorf("html = %q", r.html)
	}
	// The sidecar holds only front matter plus url_path; site keys stay out.
	want := map[string]any{"title": "T", "date": "2026-04-15T00:00:00Z", "url_path": "/posts/p/index.html"}
	if !reflect.DeepEqual(r.meta, want) {
		t.Errorf("sidecar:\n got: %v\nwant: %v", r.meta, want)
	}
	// The content output is the body alone, without the template.
	if r.content != "<p>body</p>\n" {
		t.Errorf("content = %q", r.content)
	}
}

func TestGFMExtrasAndFootnotes(t *testing.T) {
	src := "Was ~~wrong~~ right.\n\n- [x] done\n- [ ] todo\n\nA claim.[^1]\n\n[^1]: The source.\n"
	got := render(t, src, "").content
	for _, want := range []string{
		"<del>wrong</del>",
		`<input checked="" disabled="" type="checkbox"> done`,
		`<input disabled="" type="checkbox"> todo`,
		`<sup id="fnref:1"><a href="#fn:1" class="footnote-ref" role="doc-noteref">1</a></sup>`,
		`<li id="fn:1">`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestSingleTildesStayLiteral(t *testing.T) {
	got := render(t, "Packed into ~10 MB, about ~45 minutes, see ~/notes.\n", "").content
	if strings.Contains(got, "<del>") {
		t.Errorf("single tildes in prose became strikethrough:\n%s", got)
	}
}
