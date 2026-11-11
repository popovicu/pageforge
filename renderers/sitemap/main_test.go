package main

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/popovicu/pageforge/internal/testutil"
)

func TestSitemap(t *testing.T) {
	dir := t.TempDir()
	urlsFile := testutil.Write(t, dir, "posts.urls.json", `["/posts/", "/posts/a/"]`)
	metas := strings.Join([]string{
		testutil.WriteMeta(t, dir, "a.meta.json", map[string]any{
			"url_path": "/posts/a/index.html", "date": "2024-01-02T23:00:00-08:00",
		}),
		testutil.WriteMeta(t, dir, "elsewhere.meta.json", map[string]any{
			"url_path": "/not/in/sitemap/index.html", "date": "2024-01-01",
		}),
	}, ",")
	out := filepath.Join(dir, "sitemap.xml")
	err := run([]string{
		"--page-urls", "/index.html,/about/index.html,/posts/a/index.html",
		"--urls-files", urlsFile,
		"--meta-files", metas,
		"--base-url", "https://example.com/",
		"--output", out,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := `<?xml version="1.0" encoding="UTF-8"?>
<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
  <url>
    <loc>https://example.com/</loc>
  </url>
  <url>
    <loc>https://example.com/about/</loc>
  </url>
  <url>
    <loc>https://example.com/posts/</loc>
  </url>
  <url>
    <loc>https://example.com/posts/a/</loc>
    <lastmod>2024-01-03</lastmod>
  </url>
</urlset>
`
	if got := testutil.Read(t, out); got != want {
		t.Errorf("sitemap:\n got:\n%s\nwant:\n%s", got, want)
	}
}
