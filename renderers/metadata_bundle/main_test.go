package main

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/popovicu/pageforge/internal/testutil"
	"github.com/popovicu/pageforge/renderers/core"
)

func TestFilterSortLimitProject(t *testing.T) {
	dir := t.TempDir()
	var metas []string
	for name, meta := range map[string]map[string]any{
		"a": {"title": "A", "featured": true, "date": "2024-01-01", "url_path": "/posts/a/index.html", "body": "x"},
		"b": {"title": "B", "featured": false, "date": "2024-02-01", "url_path": "/posts/b/index.html"},
		"c": {"title": "C", "featured": "yes", "date": "2024-03-01", "url_path": "/posts/c/index.html"},
		"d": {"title": "D", "date": "2024-04-01", "url_path": "/posts/d/index.html"},
		"e": {"title": "E", "featured": true, "date": "2023-01-01", "url_path": "/posts/e/index.html"},
	} {
		metas = append(metas, testutil.WriteMeta(t, dir, name+".meta.json", meta))
	}
	out := filepath.Join(dir, "featured.yaml")
	err := run([]string{
		"--meta-files", strings.Join(metas, ","),
		"--key", "featured",
		"--filter", "featured",
		"--limit", "2",
		"--fields", "title, date",
		"--output", out,
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := core.LoadYAML(out)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]any{"featured": []any{
		map[string]any{"title": "C", "date": "2024-03-01", "url": "/posts/c/"},
		map[string]any{"title": "A", "date": "2024-01-01", "url": "/posts/a/"},
	}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("bundle:\n got: %#v\nwant: %#v", got, want)
	}
}

func TestNoPages(t *testing.T) {
	out := filepath.Join(t.TempDir(), "empty.yaml")
	if err := run([]string{"--meta-files", "", "--key", "recent", "--output", out}); err != nil {
		t.Fatal(err)
	}
	if got := strings.TrimSpace(testutil.Read(t, out)); got != "recent: []" {
		t.Errorf("got %q", got)
	}
}
