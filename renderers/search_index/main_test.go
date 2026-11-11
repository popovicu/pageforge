package main

import (
	"path/filepath"
	"testing"

	"github.com/popovicu/pageforge/internal/testutil"
)

func TestIncludeContent(t *testing.T) {
	dir := t.TempDir()
	meta := testutil.WriteMeta(t, dir, "a.meta.json", map[string]any{
		"url_path": "/a/index.html", "title": "A", "secret": "not indexed",
	})
	body := testutil.Write(t, dir, "a.content.html",
		"<p>Hello &amp; welcome</p>\n<script>var x = \"<p>\";</script>"+
			"<STYLE type=\"text/css\">p { color: red }</STYLE><!-- a > b --><p>Bye</p>")
	out := filepath.Join(dir, "search.json")
	err := run([]string{
		"--meta-files", meta,
		"--html-files", body,
		"--include-content",
		"--fields", "title",
		"--output", out,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := `[
  {
    "body": "Hello \u0026 welcome Bye",
    "title": "A",
    "url": "/a/index.html"
  }
]`
	if got := testutil.Read(t, out); got != want {
		t.Errorf("index:\n got: %s\nwant: %s", got, want)
	}
}

func TestNoPages(t *testing.T) {
	out := filepath.Join(t.TempDir(), "search.json")
	if err := run([]string{"--meta-files", "", "--include-content", "--output", out}); err != nil {
		t.Fatal(err)
	}
	if got := testutil.Read(t, out); got != "[]" {
		t.Errorf("got %q, want []", got)
	}
}
