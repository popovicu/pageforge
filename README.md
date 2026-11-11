# Pageforge

A Bazel-native static site generator. The public surface is a small set of
Starlark rules in [`rules/`](rules/) backed by Go binaries in
[`renderers/`](renderers/). BUILD files describe the site; Bazel handles
caching, incrementality, and the dependency graph.

> **Note.** This codebase is AI-generated. Treat the rule contracts in
> [`docs/spec.md`](docs/spec.md) and [`docs/rules.md`](docs/rules.md) as the
> source of truth, and read code before relying on it in production.

## Quick start

```sh
bazel build //...                       # build everything
bazel test  //...                       # run all tests
bazel run   //examples/basic:dev        # serve the basic example on :8080
```

The end-to-end examples in [`examples/`](examples/) are the canonical reference
for how to wire the rules together — start with
[`examples/basic`](examples/basic/) and graduate to
[`blog`](examples/blog/), [`docs`](examples/docs/), and
[`portfolio`](examples/portfolio/).

## Rules

All loaded from `@pageforge//rules:defs.bzl`:

| Rule | Purpose |
|------|---------|
| `site_template` | Defines a Go `html/template` with optional `extends` for inheritance. |
| `site_static` | Passes static files (CSS, images, fonts) through to the site under a URL prefix. |
| `site_markdown` | Renders a markdown file with YAML front matter through a template chain. |
| `site_page` | Renders a YAML data file through a template (no markdown body). |
| `site_collection` | Aggregates pages into an index, with sorting and pagination. |
| `site_tags` | Builds per-tag listing pages and a tag index from any set of pages. |
| `site_build` | Stitches `deps + collections + static` into a single tarball. |
| `site_serve` | Dev server target — extracts the tarball and serves it on a port. |

See [`docs/rules.md`](docs/rules.md) for full attribute tables and provider
shapes.

## Layout

```
rules/        Starlark rules (defs.bzl is the public API)
renderers/    Go binaries invoked by the rules
internal/     Shared Go packages (templatefuncs, etc.)
third_party/  BUILD aliases for external deps
examples/     End-to-end examples and integration tests
docs/         spec.md, rules.md (authoritative)
```

## Conventions

- Rules are agnostic of template inheritance. Page-producing rules accept one
  `template` attribute; walking the chain is `site_template`'s job, exposed via
  `TemplateInfo.files` (leaf first).
- Every file-producing rule returns `SiteOutputInfo` with a list of
  `SiteFileEntry`. Some rules (`site_static`, `site_collection`, `site_tags`)
  produce many files from one target — preserve the list shape.
- Front matter is free-form. Pageforge does not enforce a schema; consumer
  projects define their own metadata fields.
- Go renderers are pure binaries with stdin/JSON or flag inputs. They never
  read the workspace directly — Bazel passes inputs as files.
- Aggregation rules read `.meta.json` sidecars, not rendered HTML.
  `site_markdown` emits both via `OutputGroupInfo` (`html`, `metadata`).
- `site_build` merges single-file outputs (`deps`, `static`) with `.tar`
  archives (`collections`) into one final tar. Don't add aggregating rules to
  `deps` — pass them via `collections`.

## Build environment

Bazel with bzlmod. Go SDK 1.26.2 is pinned in
[`MODULE.bazel`](MODULE.bazel). There is no `go build` workflow — go through
Bazel. Module path: `github.com/popovicu/pageforge`.
