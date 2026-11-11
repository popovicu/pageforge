package main

import (
	"fmt"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/popovicu/pageforge/internal/testutil"
)

const listTmpl = `{{.URL}}|{{.CurrentPage}}/{{.TotalPages}}|{{.PrevURL}}|{{.NextURL}}|{{range .Pages}}{{.URL}},{{end}}|{{.Title}}`

// build runs the collection builder over n dated pages and returns the tar
// entries plus the urls.json listing.
func build(t *testing.T, n int, extra ...string) (map[string]string, string) {
	t.Helper()
	dir := t.TempDir()
	var metas []string
	for i := 1; i <= n; i++ {
		metas = append(metas, testutil.WriteMeta(t, dir, fmt.Sprintf("p%d.meta.json", i), map[string]any{
			"url_path": fmt.Sprintf("/p%d", i),
			"date":     fmt.Sprintf("2024-01-0%d", i),
		}))
	}
	out := filepath.Join(dir, "out.tar")
	urls := filepath.Join(dir, "urls.json")
	args := append([]string{
		"--meta-files", strings.Join(metas, ","),
		"--templates", testutil.Write(t, dir, "list.html", listTmpl),
		"--output", out,
		"--urls-output", urls,
		"--extra-metadata", `{"title":"All"}`,
	}, extra...)
	if err := run(args); err != nil {
		t.Fatal(err)
	}
	return testutil.ReadTar(t, out), testutil.Read(t, urls)
}

func TestPagination(t *testing.T) {
	entries, urls := build(t, 5, "--url-path", "/posts", "--paginate", "2")
	want := map[string]string{
		"posts/index.html":        "/posts/|1/3||/posts/page/2/|/p5,/p4,|All",
		"posts/page/2/index.html": "/posts/page/2/|2/3|/posts/|/posts/page/3/|/p3,/p2,|All",
		"posts/page/3/index.html": "/posts/page/3/|3/3|/posts/page/2/||/p1,|All",
	}
	if !reflect.DeepEqual(entries, want) {
		t.Errorf("entries:\n got: %v\nwant: %v", entries, want)
	}
	if got := strings.Join(strings.Fields(urls), ""); got != `["/posts/","/posts/page/2/","/posts/page/3/"]` {
		t.Errorf("urls.json = %s", urls)
	}
}

func TestRootAndTrailingSlash(t *testing.T) {
	for _, urlPath := range []string{"/", "//"} {
		entries, _ := build(t, 1, "--url-path", urlPath)
		if got := entries["index.html"]; got != "/|1/1|||/p1,|All" {
			t.Errorf("url_path %q: entries = %v", urlPath, entries)
		}
	}
	entries, _ := build(t, 1, "--url-path", "/posts/")
	if _, ok := entries["posts/index.html"]; !ok || len(entries) != 1 {
		t.Errorf("url_path /posts/: entries = %v", entries)
	}
}

func TestNoPages(t *testing.T) {
	entries, _ := build(t, 0, "--url-path", "/posts")
	if got := entries["posts/index.html"]; got != "/posts/|1/1||||All" {
		t.Errorf("entries = %v", entries)
	}
}
