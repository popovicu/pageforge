// metadata_bundle reads .meta.json sidecar files produced by site_markdown
// (and any other rule that emits an OutputGroupInfo "metadata" group),
// optionally filters and sorts the entries, projects a subset of fields,
// and writes a single YAML file with one top-level key whose value is the
// resulting list. The output slots into site_page's extra_data attribute.
//
// The renderer never reads the wall clock — Bazel actions must be
// reproducible.
//
// Each emitted entry always carries a synthesized `url` field: the
// canonical URL (url_path with any trailing /index.html stripped). This is
// the form templates want for <a href="">.
//
// Flags:
//
//	--meta-files   Comma-separated .meta.json files to aggregate
//	--key          Top-level YAML key wrapping the entry list
//	--filter       Metadata field; entries are kept only if it is truthy
//	--sort-by      Metadata field to sort by (default: 'date')
//	--sort-order   Sort order: 'asc' or 'desc' (default: 'desc')
//	--limit        Max entries; 0 means unlimited (default: 0)
//	--fields       Comma-separated metadata fields to project; empty = all
//	--output       Output YAML file path
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/popovicu/pageforge/internal/templatefuncs"
	"github.com/popovicu/pageforge/renderers/core"
	"gopkg.in/yaml.v3"
)

func run(args []string) error {
	fs := flag.NewFlagSet("metadata_bundle", flag.ContinueOnError)
	metaFiles := fs.String("meta-files", "", "Comma-separated .meta.json files")
	key := fs.String("key", "", "Top-level YAML key wrapping the entry list")
	filterKey := fs.String("filter", "", "Metadata field; entries kept only if truthy")
	sortBy := fs.String("sort-by", "date", "Metadata field to sort by")
	sortOrder := fs.String("sort-order", "desc", "Sort order: asc or desc")
	limit := fs.Int("limit", 0, "Max entries (0 = unlimited)")
	fieldsArg := fs.String("fields", "", "Comma-separated fields to project (empty = all)")
	output := fs.String("output", "", "Output YAML file path")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *key == "" || *output == "" {
		return fmt.Errorf("--key and --output are required")
	}

	pages, err := core.LoadMetaFiles(*metaFiles)
	if err != nil {
		return err
	}

	if *filterKey != "" {
		filtered := pages[:0]
		for _, p := range pages {
			if isTruthy(p.Metadata[*filterKey]) {
				filtered = append(filtered, p)
			}
		}
		pages = filtered
	}

	core.SortPages(pages, *sortBy, *sortOrder)

	if *limit > 0 && len(pages) > *limit {
		pages = pages[:*limit]
	}

	var fields []string
	if s := strings.TrimSpace(*fieldsArg); s != "" {
		for _, f := range strings.Split(s, ",") {
			fields = append(fields, strings.TrimSpace(f))
		}
	}

	entries := make([]map[string]any, 0, len(pages))
	for _, p := range pages {
		entries = append(entries, projectEntry(p.Metadata, fields))
	}

	doc := map[string]any{*key: entries}
	out, err := yaml.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshaling YAML: %w", err)
	}
	if err := os.WriteFile(*output, out, 0o644); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}
	return nil
}

// projectEntry returns a fresh map containing the requested fields from
// meta (or all of meta when fields is empty), plus a synthesized `url`
// derived from url_path (trailing /index.html stripped) so templates can
// link without restating the canonicalization dance on every iteration.
func projectEntry(meta map[string]any, fields []string) map[string]any {
	out := make(map[string]any)
	if len(fields) == 0 {
		for k, v := range meta {
			out[k] = v
		}
	} else {
		for _, f := range fields {
			if v, ok := meta[f]; ok {
				out[f] = v
			}
		}
	}
	if u := core.MetaString(meta, "url_path"); u != "" {
		out["url"] = templatefuncs.CanonicalURL(u)
	}
	return out
}

// isTruthy mirrors the loose YAML/JSON sense of "set and non-empty". Bool
// true, non-zero numbers, non-empty strings/slices/maps all pass; nil,
// false, zero, "" do not. This is what `filter = "featured"` callers
// expect — front matter `featured: true` passes, `featured: false` and
// missing keys do not.
func isTruthy(v any) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case string:
		return x != ""
	case int:
		return x != 0
	case int64:
		return x != 0
	case float64:
		return x != 0
	case []any:
		return len(x) > 0
	case map[string]any:
		return len(x) > 0
	default:
		return true
	}
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "metadata_bundle: %v\n", err)
		os.Exit(1)
	}
}
