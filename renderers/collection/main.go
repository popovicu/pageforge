// collection_builder reads .meta.json sidecar files produced by
// site_markdown, sorts them by a configurable field, optionally paginates
// them, and renders listing pages through a Go html/template chain.
//
// The output is a tar archive containing the rendered HTML pages at their
// final URL paths (e.g. posts/index.html, posts/page/2/index.html).
// This tar is consumed by site_build via pkg_tar's deps mechanism.
//
// Flags:
//
//	--meta-files       Comma-separated .meta.json files to aggregate
//	--templates        Comma-separated template files (base first, leaf last)
//	--url-path         Base URL path for the collection (e.g. '/posts')
//	--output           Output tar file path
//	--urls-output      Output JSON file listing every URL the tar contains
//	--sort-by          Metadata field to sort by (default: 'date')
//	--sort-order       Sort order: 'asc' or 'desc' (default: 'desc')
//	--paginate         Items per page; 0 means all on one page (default: 0)
//	--extra-metadata   JSON object of extra metadata for the template
//	--site-config      Optional site-wide YAML merged into Metadata as a base;
//	                   --extra-metadata overlays on top.
package main

import (
	"archive/tar"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/popovicu/pageforge/renderers/core"
)

func run(args []string) error {
	fs := flag.NewFlagSet("collection", flag.ContinueOnError)
	metaFiles := fs.String("meta-files", "", "Comma-separated .meta.json files")
	templateFiles := fs.String("templates", "", "Comma-separated template files (base first, leaf last)")
	urlPath := fs.String("url-path", "", "Base URL path for this collection")
	output := fs.String("output", "", "Output tar file path")
	urlsOutput := fs.String("urls-output", "", "Output JSON file listing all generated URLs")
	sortBy := fs.String("sort-by", "date", "Metadata field to sort by")
	sortOrder := fs.String("sort-order", "desc", "Sort order: asc or desc")
	paginate := fs.Int("paginate", 0, "Items per page (0 = no pagination)")
	extraMeta := fs.String("extra-metadata", "{}", "JSON object of extra metadata")
	siteConfig := fs.String("site-config", "", "Optional site-wide YAML merged into Metadata as a base")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *templateFiles == "" || *urlPath == "" || *output == "" {
		return fmt.Errorf("--templates, --url-path, and --output are required")
	}

	pages, err := core.LoadMetaFiles(*metaFiles)
	if err != nil {
		return err
	}

	core.SortPages(pages, *sortBy, *sortOrder)

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

	// Paginate.
	pageSize := *paginate
	if pageSize <= 0 {
		pageSize = len(pages)
	}

	var batches [][]core.PageContext
	for i := 0; i < len(pages); i += pageSize {
		end := i + pageSize
		if end > len(pages) {
			end = len(pages)
		}
		batches = append(batches, pages[i:end])
	}
	if len(batches) == 0 {
		batches = [][]core.PageContext{{}}
	}

	chain := strings.Split(*templateFiles, ",")

	// Create output tar.
	tarFile, err := os.Create(*output)
	if err != nil {
		return fmt.Errorf("creating tar: %w", err)
	}
	defer tarFile.Close()
	tw := tar.NewWriter(tarFile)

	// Trim trailing slashes so url_path "/" (or "/posts/") yields "/" and
	// "index.html" rather than "//" and an absolute tar path.
	root := strings.TrimRight(*urlPath, "/")
	baseURL := strings.TrimPrefix(root, "/")

	var urls []string
	for i, batch := range batches {
		pageURL := root + "/"
		tarPath := path.Join(baseURL, "index.html")
		if i > 0 {
			pageURL = fmt.Sprintf("%s/page/%d/", root, i+1)
			tarPath = path.Join(baseURL, fmt.Sprintf("page/%d/index.html", i+1))
		}

		prevURL := ""
		if i > 0 {
			if i == 1 {
				prevURL = root + "/"
			} else {
				prevURL = fmt.Sprintf("%s/page/%d/", root, i)
			}
		}
		nextURL := ""
		if i < len(batches)-1 {
			nextURL = fmt.Sprintf("%s/page/%d/", root, i+2)
		}

		ctx := core.CollectionContext{
			Pages:       batch,
			CurrentPage: i + 1,
			TotalPages:  len(batches),
			PrevURL:     prevURL,
			NextURL:     nextURL,
			URL:         pageURL,
			Metadata:    extra,
		}

		rendered, err := core.RenderTemplateToBytes(chain, ctx)
		if err != nil {
			return fmt.Errorf("rendering page %d: %w", i+1, err)
		}

		if err := core.WriteTarEntry(tw, tarPath, rendered); err != nil {
			return err
		}
		urls = append(urls, pageURL)
	}
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
		fmt.Fprintf(os.Stderr, "collection_builder: %v\n", err)
		os.Exit(1)
	}
}
