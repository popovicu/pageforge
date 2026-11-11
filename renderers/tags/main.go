// tag_builder reads .meta.json sidecar files produced by site_markdown,
// groups pages by their "tags" metadata field, and renders a per-tag listing
// page for each tag plus an index page listing all tags.
//
// The output is a tar archive containing the rendered HTML pages at their
// final URL paths (e.g. tags/index.html, tags/go/index.html).
// This tar is consumed by site_build via pkg_tar's deps mechanism.
//
// Flags:
//
//	--meta-files         Comma-separated .meta.json files to aggregate
//	--tag-templates      Comma-separated template files for per-tag pages (base first)
//	--index-templates    Comma-separated template files for the tag index page (base first)
//	--url-prefix         URL prefix for tag pages (e.g. '/tags')
//	--output             Output tar file path
//	--urls-output        Output JSON file listing every URL the tar contains
//	--sort-by            Metadata field to sort pages within each tag (default: 'date')
//	--sort-order         Sort order within each tag: 'asc' or 'desc' (default: 'desc')
//	--extra-metadata     JSON object of extra metadata for templates
//	--site-config        Optional site-wide YAML merged into Metadata as a base;
//	                     --extra-metadata overlays on top.
package main

import (
	"archive/tar"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"sort"
	"strings"

	"github.com/popovicu/pageforge/renderers/core"
)

func run(args []string) error {
	fs := flag.NewFlagSet("tags", flag.ContinueOnError)
	metaFiles := fs.String("meta-files", "", "Comma-separated .meta.json files")
	tagTemplates := fs.String("tag-templates", "", "Comma-separated template files for per-tag pages (base first)")
	indexTemplates := fs.String("index-templates", "", "Comma-separated template files for tag index (base first)")
	urlPrefix := fs.String("url-prefix", "", "URL prefix for tags (e.g. '/tags')")
	output := fs.String("output", "", "Output tar file path")
	urlsOutput := fs.String("urls-output", "", "Output JSON file listing all generated URLs")
	sortBy := fs.String("sort-by", "date", "Metadata field to sort by")
	sortOrder := fs.String("sort-order", "desc", "Sort order: asc or desc")
	extraMeta := fs.String("extra-metadata", "{}", "JSON object of extra metadata")
	siteConfig := fs.String("site-config", "", "Optional site-wide YAML merged into Metadata as a base")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *tagTemplates == "" || *indexTemplates == "" || *urlPrefix == "" || *output == "" {
		return fmt.Errorf("--tag-templates, --index-templates, --url-prefix, and --output are required")
	}

	pages, err := core.LoadMetaFiles(*metaFiles)
	if err != nil {
		return err
	}

	// Group pages by tag.
	tagMap := map[string][]core.PageContext{}
	for _, p := range pages {
		for _, tag := range p.Tags() {
			// Tags become path segments; anything that could escape the
			// prefix or vanish from the path would misplace the page.
			if tag == "" || tag == "." || tag == ".." || strings.ContainsAny(tag, `/\`) {
				return fmt.Errorf("tag %q on %s cannot be used as a URL path segment", tag, p.URL)
			}
			tagMap[tag] = append(tagMap[tag], p)
		}
	}

	// Sort pages within each tag.
	for _, tagPages := range tagMap {
		core.SortPages(tagPages, *sortBy, *sortOrder)
	}

	// Sort tag names for deterministic output.
	var tagNames []string
	for tag := range tagMap {
		tagNames = append(tagNames, tag)
	}
	sort.Strings(tagNames)

	extra, err := core.ParseExtraMetadata(*extraMeta)
	if err != nil {
		return err
	}

	if *siteConfig != "" {
		base, err := core.LoadYAML(*siteConfig)
		if err != nil {
			return fmt.Errorf("loading site config: %w", err)
		}
		extra = core.MergeMetadata(base, extra)
	}

	tagChain := strings.Split(*tagTemplates, ",")
	indexChain := strings.Split(*indexTemplates, ",")

	// Create output tar.
	tarFile, err := os.Create(*output)
	if err != nil {
		return fmt.Errorf("creating tar: %w", err)
	}
	defer tarFile.Close()
	tw := tar.NewWriter(tarFile)

	prefix := strings.TrimRight(*urlPrefix, "/")
	baseURL := strings.TrimPrefix(prefix, "/")

	var urls []string

	// Render per-tag pages.
	for _, tag := range tagNames {
		tagURL := fmt.Sprintf("%s/%s/", prefix, tag)
		tarPath := path.Join(baseURL, tag, "index.html")

		ctx := core.TagPageContext{
			Tag:      tag,
			Pages:    tagMap[tag],
			URL:      tagURL,
			Metadata: extra,
		}

		rendered, err := core.RenderTemplateToBytes(tagChain, ctx)
		if err != nil {
			return fmt.Errorf("rendering tag page %q: %w", tag, err)
		}

		if err := core.WriteTarEntry(tw, tarPath, rendered); err != nil {
			return err
		}
		urls = append(urls, tagURL)
	}

	// Render tag index.
	var entries []core.TagIndexEntry
	for _, tag := range tagNames {
		entries = append(entries, core.TagIndexEntry{
			Tag:   tag,
			Count: len(tagMap[tag]),
			URL:   fmt.Sprintf("%s/%s/", prefix, tag),
		})
	}

	indexURL := prefix + "/"
	indexCtx := core.TagIndexContext{
		Tags:     entries,
		URL:      indexURL,
		Metadata: extra,
	}

	rendered, err := core.RenderTemplateToBytes(indexChain, indexCtx)
	if err != nil {
		return fmt.Errorf("rendering tag index: %w", err)
	}

	if err := core.WriteTarEntry(tw, path.Join(baseURL, "index.html"), rendered); err != nil {
		return err
	}
	urls = append(urls, indexURL)
	if err := core.CloseTar(tw, tarFile); err != nil {
		return err
	}

	if *urlsOutput != "" {
		urlsJSON, err := json.MarshalIndent(urls, "", "  ")
		if err != nil {
			return fmt.Errorf("marshaling urls: %w", err)
		}
		if err := os.WriteFile(*urlsOutput, urlsJSON, 0o644); err != nil {
			return fmt.Errorf("writing urls: %w", err)
		}
	}

	return nil
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "tag_builder: %v\n", err)
		os.Exit(1)
	}
}
