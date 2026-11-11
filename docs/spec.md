# Pageforge — Bazel-native Static Site Generator

## Overview

Pageforge is a static site generator built as a set of Bazel rules backed by Go binaries. BUILD files become the declarative language for describing a website.

**Design Principles:**
- BUILD files are the site's table of contents
- Every rule whose files are known at analysis time returns a `SiteOutputInfo` provider with a *list* of entries; aggregators (`site_collection`, `site_tags`) ship a `.tar` instead
- Go binaries do the heavy lifting (markdown rendering, template expansion, aggregation)
- Templates handle their own inheritance — page rules are agnostic of template stacking
- Bazel handles caching, incrementality, and the dependency graph

---

## Provider

```python
SiteFileEntry = struct(
    file,       # File. The generated output file.
    url_path,   # String. URL path, e.g. '/posts/my-article'.
    metadata,   # Dict. Arbitrary metadata (title, date, tags, etc.).
    kind,       # String. Built-in rules emit 'page', 'static' or 'feed'; site_sitemap skips 'static'.
)

SiteOutputInfo = provider(
    doc = "One or more files destined for the final static site.",
    fields = {
        "entries": "List of SiteFileEntry structs.",
    },
)
```

**Why a list?** A single rule often produces multiple files. `site_static` places
many files at once, `site_collection` generates paginated pages, `site_tags`
generates N tag pages + 1 index.

---

## Template System

**Key design: page-producing rules point to a single template target. They never
know about base templates, layouts, or inheritance chains. That's the template's
own concern.**

### `site_template` rule

Defines a template and its inheritance chain. Templates can extend other templates.

```python
# providers.bzl
TemplateInfo = provider(
    doc = "A resolved template with its full inheritance chain.",
    fields = {
        "files": "List of File. All template files in the chain, leaf first.",
    },
)
```

```python
# templates.bzl
def _site_template_impl(ctx):
    chain = [ctx.file.src]
    if ctx.attr.extends:
        parent_info = ctx.attr.extends[TemplateInfo]
        chain = chain + parent_info.files
    return [TemplateInfo(files = chain)]

site_template = rule(
    implementation = _site_template_impl,
    attrs = {
        "src": attr.label(allow_single_file = True, mandatory = True),
        "extends": attr.label(providers = [TemplateInfo]),
    },
)
```

Usage in BUILD files:

```python
# themes/lfg/BUILD.bazel

load("@pageforge//rules:defs.bzl", "site_template")

# Base layout — the outer shell (html, head, nav, footer)
site_template(
    name = "base",
    src = "base.html",
)

# Post template — extends base
site_template(
    name = "post",
    src = "post.html",
    extends = ":base",
)

# Chapter index — extends base
site_template(
    name = "chapter_index",
    src = "chapter_index.html",
    extends = ":base",
)

# Landing page — extends base
site_template(
    name = "landing",
    src = "landing.html",
    extends = ":base",
)
```

Then page-producing rules just say:

```python
site_markdown(
    name = "my-post",
    src = "my-post.md",
    template = "//themes/lfg:post",  # just one target, done
    url_path = "/posts/my-post/index.html",
)
```

The `site_markdown` rule reads the `TemplateInfo` provider from the template
target, gets the full chain of files, and passes them all to the Go renderer
which resolves `{{template "content" .}}` / `{{define "content"}}` directives
using Go's built-in template inheritance.

**How Go resolves the chain:**

```go
// The renderer receives all template files in the chain
tmpl, err := template.ParseFiles(chain...)
// Go's html/template natively handles {{define}} / {{template}} inheritance
// The leaf template defines "content", the base wraps it
```

This means:
- `site_markdown`, `site_page`, `site_collection` all accept a single `template` attr
- They never see `base_template` — that concept only exists in `site_template`
- You can have arbitrarily deep chains: base → section_layout → post
- Changing the base layout rebuilds all pages that transitively depend on it (Bazel tracks this)

---

## Data Model: How YAML Works

YAML serves two roles:

### 1. Front Matter (inside Markdown files)

```markdown
---
title: "Booting Alpine for the First Time"
date: 2026-04-10
description: "What actually happens when Alpine boots."
tags: [alpine, qemu, boot]
order: 3
series: from-zero-to-root
releases: [1.0]
---

# Content here...
```

The `markdown_renderer` parses front matter and:
1. Populates template variables (`{{.Title}}`, `{{.Date}}`, etc.)
2. Writes a `.meta.json` sidecar so aggregation rules can read it cheaply

Unquoted YAML dates (`date: 2026-04-10`) are turned into RFC 3339 strings
(`2026-04-10T00:00:00Z`) on load, so templates and sidecars see the same
text.

**Fields are free-form.** Pageforge enforces no schema. Consumer projects
(like LFG) define their own fields (series, releases, chapter) without
modifying Pageforge.

### Site-wide YAML

Each page-producing rule (`site_markdown`, `site_page`, `site_collection`,
`site_tags`) takes an optional `site_config` label pointing at a YAML file.
The renderer loads it and deep-merges it with the per-page data:

- Site config is the base; per-page data is the overlay.
- Top-level merge of two `map[string]any`. Per-page wins on key collisions.
- When both sides hold a map at the same key, the merge recurses one level.
- Lists are replaced wholesale, not concatenated.

The merged map becomes `.Metadata` for the rendered template. The
per-page sidecar (today only `site_markdown` writes one) stores only the
per-page front matter — site-wide keys aren't duplicated into every sidecar
and don't invalidate sidecars when site.yaml changes.

The shared merge helper lives in `renderers/core` (`MergeMetadata`); each
renderer calls it after loading `--site-config` if the flag is supplied.

### 2. Data Files (for `site_page`)

Standalone YAML for pages that aren't Markdown:

```yaml
# about.yaml
title: "About The Linux Field Guide"
description: "Who we are and why this exists."
body: |
  You've installed a distro. You've opened a terminal...
links:
  - label: GitHub
    url: https://github.com/...
```

The `template_renderer` loads this YAML and passes every key as a template
variable. The body field can optionally be run through a markdown filter in
the template using a `markdownify` function.

### Template Data Context

All renderers provide this context to templates:

```go
type PageContext struct {
    // From front matter or YAML data file — all keys available as map entries
    Title       string
    Date        string
    Description string
    Tags        []string
    // ... any custom YAML key

    // Injected by the renderer:
    Content     template.HTML  // rendered markdown body (site_markdown only)
    URL         string         // the page's url_path
}
```

For `site_collection` templates, the context includes:

```go
type CollectionContext struct {
    Pages       []PageContext  // the pages in this batch
    CurrentPage int
    TotalPages  int
    PrevURL     string         // empty if first page
    NextURL     string         // empty if last page
    // ... extra_metadata keys
}
```

---

## Rule Taxonomy

Pageforge's rules sort into four layers, along two axes: how many sources
they read (one vs. many) and what kind of artifact they emit (an HTML
page, a terminal non-HTML file, or re-consumable data). Knowing which
layer a new rule belongs to is the fastest way to design it consistently
with the rest.

|                  | **HTML page out**             | **Non-HTML artifact out**                                  |
|------------------|-------------------------------|------------------------------------------------------------|
| **One source in**| Producers (`site_markdown`, `site_page`, `site_static`) | — (Markdown / YAML *is* the source) |
| **Many sidecars in** | Aggregators (`site_collection`, `site_tags`) | Derivatives — *terminal* (`site_feed`, `site_sitemap`, `site_search_index`, `site_chroma_css`) <br/> Bundles — *re-consumed* (`site_metadata_bundle`) |

The non-HTML / many-in cell splits along a second sub-axis — whether the
artifact is *terminal* (shipped, consumed by external clients) or
*re-consumed* (intermediate, fed back into another pageforge rule). That
split matters because terminal artifacts return `SiteOutputInfo` and
flow into `site_build`; re-consumed artifacts return only `DefaultInfo`
and stay out of the published tarball.

### 1. Producers

Rules that turn source files into output pages. They emit one
`SiteFileEntry` per page (HTML); `site_markdown` also writes a `.meta.json`
sidecar carrying the parsed front matter.

- `site_markdown` — Markdown + front matter → HTML + sidecar.
- `site_page` — YAML data + template → HTML (no sidecar, so it can't feed
  aggregators).
- `site_static` — pass-through; emits one entry per file (no sidecar).

### 2. Aggregators

Rules that read sidecars from a list of producer targets, group/sort/paginate,
and emit further HTML pages.

- `site_collection` — sort + paginate a list of pages into listing pages.
- `site_tags` — group pages by their `tags` field; emit per-tag pages + an index.

### 3. Derivatives

Rules that read sidecars (or `SiteOutputInfo.entries`) and emit a single
*terminal* non-HTML artifact — terminal in the sense that the artifact
ships with the site and is consumed by external clients (browsers,
crawlers, feed readers), not by another pageforge rule. They share a
common implementation shape: take a list of page targets, transform
metadata, write one file.

- `site_chroma_css` — derived from a style name rather than from pages, but
  follows the same "emit one auxiliary artifact" pattern.
- `site_feed` — RSS 2.0 feed.
- `site_sitemap` — sitemap.xml.
- `site_search_index` — JSON search index for client-side search.

When adding a new rule that emits a single terminal artifact (OPML,
JSON-LD index, manpage manifest, …), it almost certainly belongs in this
layer. Mirror an existing derivative renderer's `main.go`: read sidecars,
transform, write one file. Don't invent a new contract.

### 4. Bundles

Rules that read sidecars and emit a non-HTML artifact intended to be
consumed *by another pageforge rule* — not by external clients. The
output never enters the site tarball; it's an intermediate data step
that lets one rule's output feed another rule's input.

- `site_metadata_bundle` — select / sort / project a slice of the catalog
  into a YAML file consumable by `site_page`'s `extra_data` attribute.

Bundles let you compose the build graph in ways that pure
producer→aggregator→derivative chains can't: a page can be driven by
catalog data without owning the selection logic. The motivating case is
a homepage with Featured + Recent sections that should both track the
post catalog automatically — see `site_metadata_bundle` in
`docs/rules.md` for the full rationale.

Bundles return only `DefaultInfo`, never `SiteOutputInfo`, so they
cannot accidentally land in `site_build`'s output. Downstream rules
depend on them via plain file dependencies (e.g.
`extra_data = [":my_bundle"]`).

When adding a new rule whose output is consumed by another pageforge rule
rather than shipped, it belongs in this layer — not in Derivatives.

#### What enters a metadata-driven aggregation

Aggregators, derivatives, and bundles all feed off `.meta.json` sidecars
through the same contract. Eligibility is defined by the contract, not a
rule name: **a target is metadata-eligible iff it exposes an
`OutputGroupInfo` with a `metadata` group whose files are JSON sidecars
carrying parsed front matter.**

Today only `site_markdown` meets that contract. `site_page` emits HTML only
and is *not* eligible. Any future rule that writes a sidecar plugs in
unchanged — no edits to consumer rules required.

`site_sitemap` uses a different contract because it doesn't need front
matter, only URLs: it reads `SiteOutputInfo.entries` from single-file
producers and the `urls` `OutputGroupInfo` group from aggregators.

### Composition

Producers feed aggregators, derivatives, and bundles. Aggregators
themselves can feed derivatives (e.g. a sitemap that includes
collection-page URLs). Bundles introduce a feedback path absent in the
other layers: aggregated catalog data flows back *into* a producer
(`site_page` via `extra_data`), so that producer's output reflects the
catalog without manual wiring. `site_build` is the terminal sink — it
consumes everything reachable via `SiteOutputInfo` and assembles the
final tarball; bundles are excluded from this intake by design.

---

## Rules

### `site_template`
Define a template with optional inheritance. See Template System above.

### `site_static`
Pass-through. Returns one `SiteFileEntry` per file in `srcs`.

```python
site_static(
    name = "styles",
    srcs = ["style.css", "reset.css"],
    url_prefix = "/css",
)
```

### `site_page`
Render a Go template with YAML data.

```python
site_page(
    name = "about",
    template = "//templates:page",
    url_path = "/about/index.html",
    data = "about.yaml",
)
```

### `site_markdown`
Parse Markdown + front matter, render through template.

```python
site_markdown(
    name = "my-post",
    src = "my-post.md",
    template = "//templates:post",
    url_path = "/posts/my-post/index.html",
)
```

### `site_collection`
Aggregation. Sort, paginate, render listings.

```python
site_collection(
    name = "all_posts",
    pages = [":post1", ":post2", ":post3"],
    template = "//templates:post_list",
    url_path = "/posts",
    paginate = 5,
    sort_by = "date",
    sort_order = "desc",
    extra_metadata = {},
)
```

Returns multiple entries (one per paginated page).

### `site_tags`
Tag aggregation.

```python
site_tags(
    name = "tags",
    pages = [":post1", ":post2"],
    index_template = "//templates:tag_index",
    tag_template = "//templates:tag_page",
    url_prefix = "/tags",
)
```

### `site_metadata_bundle`
Bundle. Select / sort / project a slice of the catalog into a YAML data
file consumed by another rule (today, `site_page` via `extra_data`). No
HTML, no `SiteOutputInfo`.

```python
site_metadata_bundle(
    name = "featured_posts",
    pages = [":post1", ":post2"],
    filter = "featured",
    sort_by = "date",
    sort_order = "desc",
    fields = ["title", "description", "date"],
)
```

### `site_build`
Collect all `SiteOutputInfo`, flatten entries, assemble final directory tree.

```python
site_build(
    name = "site",
    deps = ["//content/posts:all_posts", "//content/pages:about"],
    static = ["//static:styles", "//static:images"],
    homepage = "//content:homepage",
)
```

### `site_zip`
Repack a `site_build` tar as a zip. The source is named explicitly so the
repack relationship is obvious at the call site.

```python
site_zip(
    name = "site_zip",
    tar = ":site",
)
```

### `site_serve`
Dev server. `bazel run` target.

```python
site_serve(
    name = "dev",
    site = ":site",
    port = 8080,
)
```

---

## Go Binaries

| Binary | Purpose |
|--------|---------|
| `markdown_renderer` | Parse .md + front matter, resolve template chain, render HTML |
| `template_renderer` | Fill Go template chain with YAML data |
| `collection_builder` | Sort, paginate, render listing pages |
| `tag_builder` | Group by tag, render tag pages |
| `metadata_bundle` | Select / sort / project page metadata into a YAML data file |
| `site_assembler` | Copy all outputs into final directory tree |
| `devserver` | HTTP server with optional live reload |

### Go Dependencies

- `github.com/yuin/goldmark` — Markdown rendering
- `github.com/alecthomas/chroma/v2` — Syntax highlighting
- `gopkg.in/yaml.v3` — YAML parsing
- Standard library: `html/template`, `net/http`, `encoding/json`

---

## Plugin System (future)

Plugins are Go binaries with a JSON stdin/stdout contract:

```
plugin_binary --manifest=input.json --output-dir=out/
```

Examples: RSS, sitemap, search index, social card generator, manpage fetcher.

---

## Directory Structure

```
pageforge/
  MODULE.bazel
  rules/
    defs.bzl              # Public API
    providers.bzl         # SiteOutputInfo, SiteFileEntry, TemplateInfo
    templates.bzl         # site_template rule
    static.bzl
    page.bzl
    markdown.bzl
    collection.bzl
    tags.bzl
    feed.bzl
    sitemap.bzl
    search_index.bzl
    chromacss.bzl
    build.bzl
    serve.bzl
  renderers/
    markdown/        (main.go)
    template/        (main.go)
    collection/      (main.go)
    tags/            (main.go)
    feed/            (main.go)
    sitemap/         (main.go)
    search_index/    (main.go)
    chromacss/       (main.go)
    devserver/       (main.go)
```

---

## Implementation Priority

1. **Phase 1:** Providers, `site_template`, `site_static`, `site_markdown`, `site_build`, `site_serve`
2. **Phase 2:** `site_page`, `site_collection` (pagination)
3. **Phase 3:** `site_tags`, homepage template with featured posts
4. **Phase 4:** `site_feed`, `site_sitemap`, `site_search_index`
5. **Phase 5:** `site_pdf`, plugin system, Atom feed

