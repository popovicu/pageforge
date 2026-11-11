// search_index_builder reads .meta.json sidecars (and, when --include-content
// is set, each page's rendered body) from a list of site_markdown targets and
// emits a JSON search index that consumers can fetch and feed into a client-
// side search library (Fuse.js, MiniSearch, etc.). Pageforge does not ship JS
// bundling — the index is a flat JSON array, the consumer brings the JS.
//
// Output shape: a JSON array. Each item has a "url" key plus one entry per
// field in --fields. When --include-content is true, each item also has a
// "body" key with the rendered HTML stripped of tags, entities, scripts and
// styles.
//
// Flags:
//
//	--meta-files       Comma-separated .meta.json files
//	--html-files       Comma-separated rendered .html files (parallel to --meta-files; required only when --include-content)
//	--fields           Comma-separated metadata field names to include
//	--include-content  When true, add a "body" field with stripped HTML
//	--output           Output JSON file path
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"os"
	"regexp"
	"strings"

	"github.com/popovicu/pageforge/renderers/core"
)

func run(args []string) error {
	fs := flag.NewFlagSet("search_index", flag.ContinueOnError)
	metaFiles := fs.String("meta-files", "", "Comma-separated .meta.json files")
	htmlFiles := fs.String("html-files", "", "Comma-separated rendered .html files (required when --include-content)")
	fieldsArg := fs.String("fields", "", "Comma-separated metadata field names to include")
	includeContent := fs.Bool("include-content", false, "Include stripped rendered HTML as 'body'")
	output := fs.String("output", "", "Output JSON file path")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *output == "" {
		return fmt.Errorf("--output is required")
	}

	pages, err := core.LoadMetaFiles(*metaFiles)
	if err != nil {
		return err
	}

	var fields []string
	if *fieldsArg != "" {
		fields = strings.Split(*fieldsArg, ",")
	}

	var htmlPaths []string
	if *includeContent {
		if *htmlFiles != "" {
			htmlPaths = strings.Split(*htmlFiles, ",")
		}
		if len(htmlPaths) != len(pages) {
			return fmt.Errorf("html-files (%d) and meta-files (%d) lengths differ", len(htmlPaths), len(pages))
		}
	}

	index := make([]map[string]any, 0, len(pages))
	for i, p := range pages {
		item := map[string]any{
			"url": p.URL,
		}
		for _, field := range fields {
			if v, ok := p.Metadata[field]; ok {
				item[field] = v
			}
		}
		if *includeContent {
			raw, err := os.ReadFile(htmlPaths[i])
			if err != nil {
				return fmt.Errorf("reading %s: %w", htmlPaths[i], err)
			}
			item["body"] = stripHTML(string(raw))
		}
		index = append(index, item)
	}

	out, err := json.MarshalIndent(index, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling index: %w", err)
	}
	if err := os.WriteFile(*output, out, 0o644); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}
	return nil
}

// dropRE matches markup whose text is not prose: scripts, styles and
// comments. These go before tags are stripped so their bodies vanish too.
var dropRE = regexp.MustCompile(`(?is)<script\b.*?</script\s*>|<style\b.*?</style\s*>|<!--.*?-->`)

// tagRE matches HTML tags (greedy <...> sequences). Good enough for index
// text — we are not parsing structure, just removing markup.
var tagRE = regexp.MustCompile(`<[^>]*>`)

// wsRE collapses runs of whitespace (including newlines) to a single space.
var wsRE = regexp.MustCompile(`\s+`)

func stripHTML(s string) string {
	s = dropRE.ReplaceAllString(s, " ")
	s = tagRE.ReplaceAllString(s, " ")
	s = html.UnescapeString(s)
	s = wsRE.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "search_index_builder: %v\n", err)
		os.Exit(1)
	}
}
