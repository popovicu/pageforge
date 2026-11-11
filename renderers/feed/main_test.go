package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/popovicu/pageforge/internal/testutil"
)

func runFeed(t *testing.T, metaFiles string, extra ...string) string {
	t.Helper()
	out := filepath.Join(t.TempDir(), "rss.xml")
	args := append([]string{
		"--meta-files", metaFiles,
		"--title", "Site",
		"--base-url", "https://example.com/",
		"--url-path", "/rss.xml",
		"--output", out,
	}, extra...)
	if err := run(args); err != nil {
		t.Fatal(err)
	}
	return testutil.Read(t, out)
}

func TestFeed(t *testing.T) {
	dir := t.TempDir()
	metas := strings.Join([]string{
		testutil.WriteMeta(t, dir, "new.meta.json", map[string]any{
			"url_path": "/posts/new/index.html", "date": "2024-03-01", "title": "New & <shiny>",
		}),
		testutil.WriteMeta(t, dir, "old.meta.json", map[string]any{
			"url_path": "/posts/old/index.html", "date": "2024-01-01T10:00:00-08:00", "title": "Old",
		}),
		testutil.WriteMeta(t, dir, "undated.meta.json", map[string]any{
			"url_path": "/posts/undated/index.html", "title": "Undated",
		}),
	}, ",")

	// Ascending puts the undated page first; it must not take a limit slot.
	got := runFeed(t, metas, "--sort-order", "asc", "--limit", "2")

	for _, want := range []string{
		"<link>https://example.com/</link>",
		`<atom:link href="https://example.com/rss.xml" rel="self"`,
		"<lastBuildDate>Fri, 01 Mar 2024 00:00:00 +0000</lastBuildDate>",
		"<title>New &amp; &lt;shiny&gt;</title>",
		"<link>https://example.com/posts/new/</link>",
		"<pubDate>Mon, 01 Jan 2024 18:00:00 +0000</pubDate>",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("feed is missing %q:\n%s", want, got)
		}
	}
	if n := strings.Count(got, "<item>"); n != 2 {
		t.Errorf("got %d items, want 2", n)
	}
	if strings.Contains(got, "Undated") {
		t.Error("undated page made it into the feed")
	}
	if strings.Index(got, "posts/old/") > strings.Index(got, "posts/new/") {
		t.Error("items are not in ascending date order")
	}
}

func TestEmptyFeed(t *testing.T) {
	got := runFeed(t, "")
	if strings.Contains(got, "<item>") || strings.Contains(got, "<lastBuildDate>") {
		t.Errorf("empty feed should have no items and no lastBuildDate:\n%s", got)
	}
	if !strings.Contains(got, "</channel>") {
		t.Errorf("empty feed is incomplete:\n%s", got)
	}
}

func TestAuthors(t *testing.T) {
	dir := t.TempDir()
	meta := testutil.WriteMeta(t, dir, "a.meta.json", map[string]any{
		"url_path": "/a/index.html", "date": "2024-01-01", "author": "Ann <ann@example.com>",
	})

	named := runFeed(t, meta, "--author", "Uros Popovic")
	for _, want := range []string{
		`xmlns:dc="http://purl.org/dc/elements/1.1/"`,
		"    <dc:creator>Uros Popovic</dc:creator>",
		"      <author>Ann &lt;ann@example.com&gt;</author>",
	} {
		if !strings.Contains(named, want) {
			t.Errorf("missing %q in:\n%s", want, named)
		}
	}
	if strings.Contains(named, "managingEditor") {
		t.Errorf("a plain name must not go in managingEditor:\n%s", named)
	}

	emailed := runFeed(t, meta, "--author", "me@example.com (Me)")
	if !strings.Contains(emailed, "<managingEditor>me@example.com (Me)</managingEditor>") {
		t.Errorf("email author should use managingEditor:\n%s", emailed)
	}
	if strings.Contains(emailed, "xmlns:dc") {
		t.Errorf("no dc:creator was emitted, so no dc namespace is needed:\n%s", emailed)
	}
}
