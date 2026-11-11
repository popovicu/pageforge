// markdown_renderer parses a Markdown file with YAML front matter, converts
// the Markdown body to HTML (with syntax highlighting via Chroma), renders it
// through a Go html/template chain, and writes the final HTML output.
//
// It also produces a .meta.json sidecar containing the extracted front matter
// metadata, used by downstream aggregation rules (site_collection, site_tags).
//
// Flags:
//
//	--src          Markdown source file with optional YAML front matter
//	--templates    Comma-separated template files (base first, leaf last)
//	--url-path     URL path for this page in the final site
//	--output       Output HTML file path
//	--meta-output  Output metadata JSON file path (optional)
//	--content-output  Output path for the rendered Markdown body alone,
//	               before templating (optional)
//	--site-config  Optional site-wide YAML merged into Metadata as a base;
//	               front matter overlays on top.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"os"
	"strings"

	chromahtml "github.com/alecthomas/chroma/v2/formatters/html"
	"github.com/popovicu/pageforge/renderers/core"
	"github.com/yuin/goldmark"
	highlighting "github.com/yuin/goldmark-highlighting/v2"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

func run(args []string) error {
	fs := flag.NewFlagSet("markdown", flag.ContinueOnError)
	src := fs.String("src", "", "Path to markdown source file")
	templateFiles := fs.String("templates", "", "Comma-separated template files (base first, leaf last)")
	urlPath := fs.String("url-path", "", "URL path for this page")
	output := fs.String("output", "", "Output HTML file path")
	metaOutput := fs.String("meta-output", "", "Output metadata JSON file path")
	contentOutput := fs.String("content-output", "", "Output path for the rendered body before templating")
	siteConfig := fs.String("site-config", "", "Optional site-wide YAML merged into Metadata as a base")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *src == "" || *templateFiles == "" || *output == "" {
		return fmt.Errorf("--src, --templates, and --output are required")
	}

	raw, err := os.ReadFile(*src)
	if err != nil {
		return fmt.Errorf("reading source: %w", err)
	}

	metadata, body, err := core.SplitFrontMatter(raw)
	if err != nil {
		return fmt.Errorf("parsing front matter: %w", err)
	}

	// renderMeta is the metadata passed to the template. It is the deep merge
	// of site config (base) and per-page front matter (overlay). The sidecar
	// keeps only the per-page front matter so downstream aggregators don't
	// see duplicated site keys and don't invalidate when site.yaml changes.
	renderMeta := metadata
	if *siteConfig != "" {
		base, err := core.LoadYAML(*siteConfig)
		if err != nil {
			return fmt.Errorf("loading site config: %w", err)
		}
		renderMeta = core.MergeMetadata(base, metadata)
	}

	// Raw HTML in markdown is allowed: pageforge treats sources as
	// trusted (single-author, version-controlled, not user-generated).
	// Without this, goldmark replaces blocks like <iframe>/<img>/<video>
	// with <!-- raw HTML omitted --> placeholders.
	extensions := []goldmark.Extender{
		highlighting.NewHighlighting(
			highlighting.WithFormatOptions(
				chromahtml.WithClasses(true),
			),
		),
		// GitHub Flavored Markdown: pipe tables, ~~strikethrough~~,
		// "- [ ]" task lists, and auto-linked bare http(s):// URLs so
		// authors can paste a URL without wrapping it in [text](url).
		extension.GFM,
		// [^1] footnotes, collected into a list at the end of the body.
		extension.Footnote,
	}
	// Auto-TOC is on by default; an explicit `auto_toc: false` in front
	// matter suppresses it. Per-page metadata is what matters for the
	// opt-out (per-post decision); marker / summary text is read from
	// renderMeta so site-wide values in site_config feed in too.
	if autoTOCEnabled(metadata) {
		marker := tocString(renderMeta, "toc_marker", "Table of contents")
		// Default summary tracks the marker so an overridden marker yields a
		// matching summary without forcing the consumer to set both keys.
		summary := tocString(renderMeta, "toc_summary", "Open "+marker)
		extensions = append(extensions, tocExtension{
			Marker:  marker,
			Summary: summary,
		})
	}
	md := goldmark.New(
		goldmark.WithParserOptions(parser.WithAutoHeadingID()),
		goldmark.WithRendererOptions(html.WithUnsafe()),
		goldmark.WithExtensions(extensions...),
	)

	var rendered bytes.Buffer
	if err := md.Convert(body, &rendered); err != nil {
		return fmt.Errorf("rendering markdown: %w", err)
	}

	ctx := core.PageContext{
		Content:  template.HTML(rendered.String()),
		URL:      *urlPath,
		Metadata: renderMeta,
	}

	chain := strings.Split(*templateFiles, ",")
	if err := core.RenderTemplate(chain, ctx, *output); err != nil {
		return err
	}

	if *contentOutput != "" {
		if err := os.WriteFile(*contentOutput, rendered.Bytes(), 0o644); err != nil {
			return fmt.Errorf("writing content: %w", err)
		}
	}

	if *metaOutput != "" {
		metadata["url_path"] = *urlPath
		metaJSON, err := json.MarshalIndent(metadata, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling metadata: %w", err)
		}
		if err := os.WriteFile(*metaOutput, metaJSON, 0o644); err != nil {
			return fmt.Errorf("writing metadata: %w", err)
		}
	}

	return nil
}

// tocString reads a string-valued key from the merged metadata map,
// falling back to def when missing or non-string. Empty strings count as
// missing — an explicit empty marker would silently disable matching, and
// an empty summary would render an unlabeled disclosure widget; neither is
// a useful state worth supporting.
func tocString(meta map[string]interface{}, key, def string) string {
	v, ok := meta[key]
	if !ok {
		return def
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return def
	}
	return s
}

// autoTOCEnabled reports whether auto-TOC injection should run for this
// page. The default is on; only an explicit boolean `false` (or string
// "false") in front matter suppresses it.
func autoTOCEnabled(metadata map[string]interface{}) bool {
	v, ok := metadata["auto_toc"]
	if !ok {
		return true
	}
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return !strings.EqualFold(t, "false")
	default:
		return true
	}
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "markdown_renderer: %v\n", err)
		os.Exit(1)
	}
}
