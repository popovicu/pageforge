// template_renderer loads a YAML data file and renders it through a Go
// html/template chain. Use this for pages that are not Markdown-based,
// such as landing pages or about pages where the content is structured data.
//
// All YAML keys are available in templates via .Metadata, with convenience
// accessors like .Title, .Date, and .Description.
//
// Flags:
//
//	--data         YAML data file
//	--templates    Comma-separated template files (base first, leaf last)
//	--url-path     URL path for this page in the final site
//	--output       Output HTML file path
//	--site-config  Optional site-wide YAML merged into Metadata as a base
//	--extra-data   Optional comma-separated YAML files merged in between
//	               site_config (base) and the primary data (top overlay).
//	               Merged in list order; the primary data file still wins.
package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/popovicu/pageforge/renderers/core"
)

func run(args []string) error {
	fs := flag.NewFlagSet("template", flag.ContinueOnError)
	data := fs.String("data", "", "Path to YAML data file")
	templateFiles := fs.String("templates", "", "Comma-separated template files (base first, leaf last)")
	urlPath := fs.String("url-path", "", "URL path for this page")
	output := fs.String("output", "", "Output HTML file path")
	siteConfig := fs.String("site-config", "", "Optional site-wide YAML merged into Metadata as a base")
	extraData := fs.String("extra-data", "", "Optional comma-separated YAML files merged between site-config and data")
	if err := fs.Parse(args); err != nil {
		return err
	}

	if *data == "" || *templateFiles == "" || *output == "" {
		return fmt.Errorf("--data, --templates, and --output are required")
	}

	dataMeta, err := core.LoadYAML(*data)
	if err != nil {
		return err
	}

	metadata := map[string]any{}
	if *siteConfig != "" {
		base, err := core.LoadYAML(*siteConfig)
		if err != nil {
			return fmt.Errorf("loading site config: %w", err)
		}
		metadata = core.MergeMetadata(metadata, base)
	}
	if s := strings.TrimSpace(*extraData); s != "" {
		for _, path := range strings.Split(s, ",") {
			path = strings.TrimSpace(path)
			if path == "" {
				continue
			}
			extra, err := core.LoadYAML(path)
			if err != nil {
				return fmt.Errorf("loading extra-data %s: %w", path, err)
			}
			metadata = core.MergeMetadata(metadata, extra)
		}
	}
	metadata = core.MergeMetadata(metadata, dataMeta)

	ctx := core.PageContext{
		URL:      *urlPath,
		Metadata: metadata,
	}

	chain := strings.Split(*templateFiles, ",")
	return core.RenderTemplate(chain, ctx, *output)
}

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "template_renderer: %v\n", err)
		os.Exit(1)
	}
}
