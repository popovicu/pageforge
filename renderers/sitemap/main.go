// sitemap_builder emits a sitemap.xml conforming to the sitemaps.org schema.
//
// URLs come from two sources:
//
//   - --page-urls: a comma-separated list of URL paths (the rule fills these
//     directly from each dep's SiteOutputInfo.entries).
//   - --urls-files: a comma-separated list of urls.json sidecar files
//     produced by aggregator rules (site_collection, site_tags). Each file
//     is a JSON array of strings.
//
// When --meta-files supplies .meta.json sidecars (today, only site_markdown
// emits them), the renderer keys by url_path and emits a date-only
// <lastmod>YYYY-MM-DD</lastmod> for any entry whose `date` field is a
// parseable RFC3339 timestamp or YYYY-MM-DD value. Date-only is the W3C
// Datetime form Google recommends when a precise modification time isn't
// available. URLs without metadata get no <lastmod> — that's a no-op for
// the schema, not a regression.
//
// All URLs are absolutized against --base-url. No <priority> elements are
// emitted (deprecated by Google).
//
// Flags:
//
//	--page-urls    Comma-separated URL paths from single-file producers
//	--urls-files   Comma-separated urls.json files from aggregators
//	--meta-files   Comma-separated .meta.json files (provides <lastmod>)
//	--base-url     Site base URL (e.g. 'https://example.com')
//	--output       Output XML file path
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/popovicu/pageforge/internal/templatefuncs"
	"github.com/popovicu/pageforge/renderers/core"
)

func run(args []string) error {
	fs := flag.NewFlagSet("sitemap", flag.ContinueOnError)
	pageURLs := fs.String("page-urls", "", "Comma-separated URL paths from single-file producers")
	urlsFiles := fs.String("urls-files", "", "Comma-separated urls.json files from aggregators")
	metaFiles := fs.String("meta-files", "", "Comma-separated .meta.json files (for <lastmod>)")
	baseURL := fs.String("base-url", "", "Site base URL")
	output := fs.String("output", "", "Output XML file path")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *baseURL == "" || *output == "" {
		return fmt.Errorf("--base-url and --output are required")
	}

	base := strings.TrimRight(*baseURL, "/")

	// Canonical url -> lastmod (date-only YYYY-MM-DD); empty = no lastmod.
	// Keying on the canonical form keeps /x/ and /x/index.html from
	// producing two identical <loc>s.
	urls := map[string]string{}
	if *pageURLs != "" {
		for _, u := range strings.Split(*pageURLs, ",") {
			if u == "" {
				continue
			}
			urls[templatefuncs.CanonicalURL(u)] = ""
		}
	}
	if *urlsFiles != "" {
		for _, path := range strings.Split(*urlsFiles, ",") {
			if path == "" {
				continue
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("reading %s: %w", path, err)
			}
			var us []string
			if err := json.Unmarshal(raw, &us); err != nil {
				return fmt.Errorf("parsing %s: %w", path, err)
			}
			for _, u := range us {
				u = templatefuncs.CanonicalURL(u)
				if _, ok := urls[u]; !ok {
					urls[u] = ""
				}
			}
		}
	}
	if *metaFiles != "" {
		for _, path := range strings.Split(*metaFiles, ",") {
			if path == "" {
				continue
			}
			raw, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("reading %s: %w", path, err)
			}
			var meta map[string]any
			if err := json.Unmarshal(raw, &meta); err != nil {
				return fmt.Errorf("parsing %s: %w", path, err)
			}
			urlPath, _ := meta["url_path"].(string)
			if urlPath == "" {
				continue
			}
			urlPath = templatefuncs.CanonicalURL(urlPath)
			lastmod := formatLastmod(core.MetaString(meta, "date"))
			if lastmod == "" {
				continue
			}
			if _, known := urls[urlPath]; known {
				urls[urlPath] = lastmod
			}
		}
	}

	type urlEntry struct {
		Loc     string
		Lastmod string
	}
	entries := make([]urlEntry, 0, len(urls))
	for u, lm := range urls {
		entries = append(entries, urlEntry{Loc: core.AbsURL(base, u), Lastmod: lm})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Loc < entries[j].Loc })

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	for _, e := range entries {
		b.WriteString("  <url>\n")
		fmt.Fprintf(&b, "    <loc>%s</loc>\n", core.XMLEscape(e.Loc))
		if e.Lastmod != "" {
			fmt.Fprintf(&b, "    <lastmod>%s</lastmod>\n", e.Lastmod)
		}
		b.WriteString("  </url>\n")
	}
	b.WriteString(`</urlset>` + "\n")

	if err := os.WriteFile(*output, []byte(b.String()), 0o644); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}
	return nil
}

// formatLastmod returns a W3C Datetime (date-only) suitable for <lastmod>,
// or "" when the value is not a date.
func formatLastmod(s string) string {
	t, ok := templatefuncs.ParseDate(s)
	if !ok {
		return ""
	}
	return t.UTC().Format("2006-01-02")
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "sitemap_builder: %v\n", err)
		os.Exit(1)
	}
}
