// feed_builder reads .meta.json sidecar files produced by site_markdown and
// emits an RSS 2.0 feed XML file.
//
// The output is a single rss.xml file at the path given by --output.
//
// Channel-level <lastBuildDate> is the latest item date, omitted when the
// feed has no items. The renderer never reads the wall clock — Bazel actions
// must be reproducible.
//
// Flags:
//
//	--meta-files     Comma-separated .meta.json files to include
//	--title          Channel title
//	--description    Channel description
//	--base-url       Site base URL (e.g. 'https://example.com'); item links are base-url + url_path
//	--url-path       URL path of the feed itself (used in <atom:link rel="self">)
//	--output         Output XML file path
//	--sort-by        Metadata field to sort by (default: 'date')
//	--sort-order     Sort order: 'asc' or 'desc' (default: 'desc')
//	--limit          Max items; 0 means unlimited (default: 0)
//	--author         Channel-level author (optional): <managingEditor> when it
//	                 holds an email address, <dc:creator> for a plain name
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/popovicu/pageforge/internal/templatefuncs"
	"github.com/popovicu/pageforge/renderers/core"
)

func run(args []string) error {
	fs := flag.NewFlagSet("feed", flag.ContinueOnError)
	metaFiles := fs.String("meta-files", "", "Comma-separated .meta.json files")
	title := fs.String("title", "", "Channel title")
	description := fs.String("description", "", "Channel description")
	baseURL := fs.String("base-url", "", "Site base URL")
	urlPath := fs.String("url-path", "", "Feed's own URL path")
	output := fs.String("output", "", "Output XML file path")
	sortBy := fs.String("sort-by", "date", "Metadata field to sort by")
	sortOrder := fs.String("sort-order", "desc", "Sort order: asc or desc")
	limit := fs.Int("limit", 0, "Max items (0 = unlimited)")
	author := fs.String("author", "", "Channel author: an email address for <managingEditor>, or a name for <dc:creator>")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *title == "" || *baseURL == "" || *urlPath == "" || *output == "" {
		return fmt.Errorf("--title, --base-url, --url-path, and --output are required")
	}

	pages, err := core.LoadMetaFiles(*metaFiles)
	if err != nil {
		return err
	}

	// Undated entries have no place in a feed. Drop them before the limit
	// so they can't take slots from dated items.
	dated := pages[:0]
	for _, p := range pages {
		if _, ok := templatefuncs.ParseDate(core.MetaString(p.Metadata, "date")); ok {
			dated = append(dated, p)
		}
	}
	pages = dated

	core.SortPages(pages, *sortBy, *sortOrder)

	if *limit > 0 && len(pages) > *limit {
		pages = pages[:*limit]
	}

	base := strings.TrimRight(*baseURL, "/")

	// The channel is written first so the root element can declare the
	// Dublin Core namespace only when a dc:creator was emitted.
	var b strings.Builder
	usesDC := false
	b.WriteString("  <channel>\n")
	fmt.Fprintf(&b, "    <title>%s</title>\n", core.XMLEscape(*title))
	fmt.Fprintf(&b, "    <link>%s/</link>\n", core.XMLEscape(base))
	fmt.Fprintf(&b, "    <description>%s</description>\n", core.XMLEscape(*description))
	fmt.Fprintf(&b, "    <atom:link href=\"%s\" rel=\"self\" type=\"application/rss+xml\"/>\n", core.XMLEscape(core.AbsURL(base, *urlPath)))
	if *author != "" {
		usesDC = writeAuthor(&b, "    ", "managingEditor", *author) || usesDC
	}

	// lastBuildDate = latest item date; omitted for a feed with no items.
	var latest time.Time
	for _, p := range pages {
		if t, ok := templatefuncs.ParseDate(core.MetaString(p.Metadata, "date")); ok && t.After(latest) {
			latest = t
		}
	}
	if !latest.IsZero() {
		fmt.Fprintf(&b, "    <lastBuildDate>%s</lastBuildDate>\n", latest.UTC().Format(time.RFC1123Z))
	}

	for _, p := range pages {
		t, _ := templatefuncs.ParseDate(core.MetaString(p.Metadata, "date"))
		link := core.AbsURL(base, p.URL)
		b.WriteString("    <item>\n")
		fmt.Fprintf(&b, "      <title>%s</title>\n", core.XMLEscape(p.Title()))
		fmt.Fprintf(&b, "      <link>%s</link>\n", core.XMLEscape(link))
		fmt.Fprintf(&b, "      <guid isPermaLink=\"true\">%s</guid>\n", core.XMLEscape(link))
		if desc := p.Description(); desc != "" {
			fmt.Fprintf(&b, "      <description>%s</description>\n", core.XMLEscape(desc))
		}
		if itemAuthor := core.MetaString(p.Metadata, "author"); itemAuthor != "" {
			usesDC = writeAuthor(&b, "      ", "author", itemAuthor) || usesDC
		}
		fmt.Fprintf(&b, "      <pubDate>%s</pubDate>\n", t.UTC().Format(time.RFC1123Z))
		b.WriteString("    </item>\n")
	}

	b.WriteString("  </channel>\n")
	b.WriteString(`</rss>` + "\n")

	ns := `xmlns:atom="http://www.w3.org/2005/Atom"`
	if usesDC {
		ns += ` xmlns:dc="http://purl.org/dc/elements/1.1/"`
	}
	doc := `<?xml version="1.0" encoding="UTF-8"?>` + "\n" +
		`<rss version="2.0" ` + ns + ">\n" + b.String()
	if err := os.WriteFile(*output, []byte(doc), 0o644); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}
	return nil
}

// writeAuthor emits an author in a form RSS 2.0 accepts. managingEditor and
// author must hold an email address, so a plain name goes in dc:creator
// instead. It reports whether dc:creator was used.
func writeAuthor(b *strings.Builder, indent, emailElem, value string) bool {
	elem := emailElem
	if !strings.Contains(value, "@") {
		elem = "dc:creator"
	}
	fmt.Fprintf(b, "%s<%s>%s</%s>\n", indent, elem, core.XMLEscape(value), elem)
	return elem == "dc:creator"
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "feed_builder: %v\n", err)
		os.Exit(1)
	}
}
