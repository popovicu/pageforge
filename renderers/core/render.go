package core

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"

	"github.com/popovicu/pageforge/internal/templatefuncs"
)

// NewTemplate constructs a fresh html/template.Template with pageforge's
// shared FuncMap attached. Route every template construction through this
// helper so no render path misses the FuncMap.
//
// TODO(funcmap-plugins): let downstream sites register their own template
// funcs without editing upstream pageforge. Sketch of the right shape:
//
//  1. Keep `templatefuncs.Default()` as the baseline. Extend `NewTemplate`
//     (or add `NewTemplateWith(extra template.FuncMap)`) so renderer `main.go`
//     files can merge an additional FuncMap at construction time. Merge
//     order: Default first, user overrides win — gives sites a way to swap
//     `formatTime` without a fork.
//  2. Make the renderer binaries composable. `renderers/core` already does
//     the work; the `main.go` per renderer is a thin driver. Factor each
//     main into a `Run(extra template.FuncMap) error` entry point so a
//     downstream repo can write its own `main.go` that imports
//     `renderers/markdown` + its own funcs package and calls `Run(myFuncs)`.
//  3. Bazel side: add an optional `renderer` label attr to the page-producing
//     rules (`site_markdown`, `site_page`, `site_collection`, `site_tags`)
//     that defaults to the upstream `//renderers/...` binary. Sites that
//     want custom funcs point the attr at a `go_binary` they built using
//     step 2. No new plugin API, no shared-object plugins, no reflection —
//     just "swap the binary, it's the same CLI contract."
//  4. Optional ergonomic layer: a `site_funcs(name, deps)` macro that wraps
//     the boilerplate of generating the custom `main.go` + `go_binary` +
//     forwarding the binary to `renderer = ...` on downstream rules.
//
// Avoid: Go `plugin` package (Linux-only, incompatible with pure=on), global
// init()-driven registries (test isolation pain), and a Starlark-level func
// DSL (cannot express anything real without reinventing `text/template`).
func NewTemplate(name string) *template.Template {
	return template.New(name).Funcs(templatefuncs.Default())
}

// RenderTemplateToBytes parses the template chain, executes it with the given
// context, and returns the rendered bytes.
func RenderTemplateToBytes(chain []string, ctx any) ([]byte, error) {
	if len(chain) == 0 {
		return nil, fmt.Errorf("parsing templates: empty template chain")
	}

	tmpl, err := NewTemplate(filepath.Base(chain[0])).ParseFiles(chain...)
	if err != nil {
		return nil, fmt.Errorf("parsing templates: %w", err)
	}

	var out bytes.Buffer
	if err := tmpl.Execute(&out, ctx); err != nil {
		return nil, fmt.Errorf("executing template: %w", err)
	}

	return out.Bytes(), nil
}

// RenderTemplate parses the template chain and executes it with the given
// context, writing the result to outputPath.
func RenderTemplate(chain []string, ctx any, outputPath string) error {
	rendered, err := RenderTemplateToBytes(chain, ctx)
	if err != nil {
		return err
	}

	if err := os.WriteFile(outputPath, rendered, 0o644); err != nil {
		return fmt.Errorf("writing output: %w", err)
	}

	return nil
}
