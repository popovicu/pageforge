package main

import (
	"path/filepath"
	"testing"

	"github.com/popovicu/pageforge/internal/testutil"
)

func TestMergeOrder(t *testing.T) {
	dir := t.TempDir()
	out := filepath.Join(dir, "page.html")
	err := run([]string{
		"--site-config", testutil.Write(t, dir, "site.yaml", "a: site\nb: site\nc: site\nnested: {x: site, y: site}\n"),
		"--extra-data", testutil.Write(t, dir, "e1.yaml", "b: extra1\nc: extra1\n") + "," +
			testutil.Write(t, dir, "e2.yaml", "c: extra2\nd: extra2\nnested: {y: extra2}\n"),
		"--data", testutil.Write(t, dir, "data.yaml", "d: data\nnested: {z: data}\n"),
		"--templates", testutil.Write(t, dir, "page.tmpl",
			`{{.Metadata.a}} {{.Metadata.b}} {{.Metadata.c}} {{.Metadata.d}} {{.Metadata.nested.x}} {{.Metadata.nested.y}} {{.Metadata.nested.z}} {{.URL}}`),
		"--url-path", "/about/index.html",
		"--output", out,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := "site extra1 extra2 data site extra2 data /about/index.html"
	if got := testutil.Read(t, out); got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
