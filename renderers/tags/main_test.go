package main

import (
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/popovicu/pageforge/internal/testutil"
)

func runTags(t *testing.T, pages []map[string]any) (map[string]string, string, error) {
	t.Helper()
	dir := t.TempDir()
	var metas []string
	for i, p := range pages {
		metas = append(metas, testutil.WriteMeta(t, dir, string(rune('a'+i))+".meta.json", p))
	}
	out := filepath.Join(dir, "out.tar")
	urls := filepath.Join(dir, "urls.json")
	err := run([]string{
		"--meta-files", strings.Join(metas, ","),
		"--tag-templates", testutil.Write(t, dir, "tag.html", `{{.Tag}}:{{.URL}}:{{range .Pages}}{{.URL}},{{end}}`),
		"--index-templates", testutil.Write(t, dir, "index.html", `{{range .Tags}}{{.Tag}}={{.Count}}@{{.URL}};{{end}}`),
		"--url-prefix", "/tags/",
		"--output", out,
		"--urls-output", urls,
	})
	if err != nil {
		return nil, "", err
	}
	return testutil.ReadTar(t, out), testutil.Read(t, urls), nil
}

func TestGroupsAndIndex(t *testing.T) {
	entries, urls, err := runTags(t, []map[string]any{
		{"url_path": "/a", "date": "2024-01-01", "tags": []any{"go", "bazel"}},
		{"url_path": "/b", "date": "2024-02-01", "tags": []any{"go"}},
		{"url_path": "/c", "date": "2024-03-01"},
	})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"tags/bazel/index.html": "bazel:/tags/bazel/:/a,",
		"tags/go/index.html":    "go:/tags/go/:/b,/a,",
		"tags/index.html":       "bazel=1@/tags/bazel/;go=2@/tags/go/;",
	}
	if !reflect.DeepEqual(entries, want) {
		t.Errorf("entries:\n got: %v\nwant: %v", entries, want)
	}
	if got := strings.Join(strings.Fields(urls), ""); got != `["/tags/bazel/","/tags/go/","/tags/"]` {
		t.Errorf("urls.json = %s", urls)
	}
}

func TestRejectsTagsThatAreNotPathSegments(t *testing.T) {
	for _, tag := range []string{"../escape", "a/b", `a\b`, ".", ".."} {
		_, _, err := runTags(t, []map[string]any{{"url_path": "/a", "tags": []any{tag}}})
		if err == nil || !strings.Contains(err.Error(), "cannot be used as a URL path segment") {
			t.Errorf("tag %q: err = %v", tag, err)
		}
	}
}
