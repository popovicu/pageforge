# AGENTS.md

This file provides guidance to AI coding agents (Claude Code, Gemini CLI, and
others) when working with code in this repository. `CLAUDE.md` and `GEMINI.md`
are symlinks to it, so there is one copy to edit.

## What this is

Pageforge is a Bazel-native static site generator. The public surface is a set
of Starlark rules in `rules/` backed by Go binaries in `renderers/`. BUILD
files describe the site; Bazel handles caching, incrementality, and the
dependency graph.

[`docs/spec.md`](docs/spec.md) and [`docs/rules.md`](docs/rules.md) are the
authoritative references for rule contracts and provider shapes — read them
before making non-trivial changes.

## Build & test

Bazel with bzlmod. There is **no `go build` workflow** — every Go invocation
goes through Bazel. Go SDK 1.26.2 is pinned in `MODULE.bazel`.

| Command | What it does |
|---------|--------------|
| `bazel build //...` | Build everything |
| `bazel test //...` | Run all tests |
| `bazel test //renderers/feed:feed_test` | Run one specific test |
| `bazel run //examples/basic:dev` | Serve the basic example on :8080 (other examples on :8081–:8083) |
| `bazel run //examples/blog:site` | Materialize an example's tarball under `bazel-bin/` |

After editing a renderer or rule, the right validation loop is
`bazel test //...`: each renderer's `main_test.go` covers its CLI contract,
and the four examples plus their golden + tar-modes tests cover the public
surface end-to-end.

## Architecture (the parts that span multiple files)

**Two layers, one boundary.** Starlark rules in `rules/` orchestrate; Go
binaries in `renderers/` do the work. Rules invoke renderers as `ctx.actions.run`
with files (never workspace paths) as inputs. Renderers are pure binaries with
stdin/JSON or flag inputs and never read the workspace directly.

**`SiteOutputInfo` is the universal currency for shipped files.** Every
rule whose files are known at analysis time
(`site_static`, `site_markdown`, `site_page`, `site_feed`, `site_sitemap`,
`site_search_index`, `site_chroma_css`) returns a `SiteOutputInfo` whose
`entries` is a *list* of `SiteFileEntry`. Multi-file producers (static
bundles) preserve the list shape — never collapse to a single file.
`site_build` consumes the list and stitches everything into one tarball,
failing analysis if two entries claim the same URL (a validation action,
`renderers/sitecheck`, catches collisions with files inside collection
tars). Aggregators whose file
set is only known at action time (`site_collection`, `site_tags`) ship a
`.tar` plus a `urls.json` sidecar instead (see the intake channels below).
The deliberate exception is
`site_metadata_bundle`: its output is intermediate data consumed by
another rule (`site_page` via `extra_data`), never shipped, so it returns
only `DefaultInfo` and stays out of `site_build`'s intake by design.

**Template inheritance lives in `site_template`, not in page rules.** Page
rules (`site_markdown`, `site_page`, `site_collection`, `site_tags`) accept one
`template` attribute. Walking the chain is `site_template`'s job — it returns
`TemplateInfo.files` with the entire chain, leaf first. Don't teach page rules
about inheritance.

**`site_build` has two intake channels.** `deps` and `static` are single-file
outputs (one `SiteFileEntry` each). `collections` is for rules that produce
`.tar` archives (`site_collection`, `site_tags`). Don't pass aggregating rules
through `deps` — they go in `collections`.

**Aggregation reads sidecars, not HTML.** `site_markdown` emits both rendered
HTML and a `.meta.json` sidecar via `OutputGroupInfo` (`html`, `metadata`).
`site_collection` and `site_tags` consume the metadata, not the rendered
output. When extending aggregation behavior, work with the sidecars.

**Front matter is free-form.** Pageforge enforces no schema. Consumer projects
(e.g. LFG) define their own metadata fields; the renderers pass front matter
through to templates as `.Metadata`. Don't add validation or required fields.
The one normalization: unquoted YAML dates are turned into RFC 3339 strings
when YAML is loaded, so pages, sidecars and `formatTime` all see the same
text.

**Markdown renderer ships built-in goldmark behaviors.** Today: syntax
highlighting (Chroma), GFM (tables, strikethrough, task lists, autolinks),
footnotes, and an auto-TOC that turns a literal
`## Table of contents` heading into a `<details>` block linking every
subsequent H2/H3. Each behavior is a sibling `.go` file under
`renderers/markdown/` (e.g. `toc.go`) wired in `main.go`. Configuration
flows through the same channel as everything else: per-page front matter
for opt-out (`auto_toc: false`), and the merged `site_config` for
site-wide knobs (`toc_marker`, `toc_summary`). No new BUILD attrs — the
renderer takes one `site_config` label and reads everything from there.
When adding another behavior, mirror this shape: it keeps the rule
surface stable and makes overrides discoverable in one place.

**Sort keys are dates or strings.** `sort_by` (collections, tags, feeds,
bundles) compares the metadata value chronologically when every non-empty
value parses as a date, and as a string otherwise. For numeric ordering,
content must use string-padded keys (e.g. `order: "01"`) — see
`examples/docs/guides/getting-started/`. Equal keys break ties on `url_path`
ascending, so every listing orders them identically.

**Pages ship as `<path>/index.html`.** A `url_path` is the file's path in the
tarball; extensionless paths get no Content-Type on strict hosts. Feeds,
sitemaps and bundles emit the canonical form (trailing `index.html` dropped)
via `templatefuncs.CanonicalURL`, which templates reach as `canonicalURL`.

## When adding a new rule or renderer

- **New rule:** mirror `rules/markdown.bzl` (the most representative example),
  add the load to `rules/defs.bzl`, and document the attribute table and
  provider shape in `docs/rules.md`. The public API is `defs.bzl` — keep it in sync.
- **New renderer:** add `renderers/<name>/main.go` with a
  `run(args []string) error` entry point and hand-edit
  `renderers/<name>/BUILD.bazel` mirroring an existing one: a `go_library`,
  a `go_binary` embedding it, and a `go_test` that calls `run` with temp
  files (helpers in `internal/testutil`). BUILD files for Go binaries are not
  gazelle-generated. Shared template helpers go in `internal/templatefuncs`,
  as do the date and canonical-URL rules every renderer must agree on
  (`ParseDate`, `CanonicalURL`).

## Examples are the canonical reference

The four examples in `examples/` are integration tests and the wiring
reference for every public rule. Their READMEs call out exactly which
primitives each one exercises:

- [`basic`](examples/basic/) — minimal site; has a golden test.
- [`blog`](examples/blog/) — pagination, chroma, custom front matter.
- [`docs`](examples/docs/) — multiple sibling collections, padded sort keys, `extra_metadata`.
- [`portfolio`](examples/portfolio/) — three-level template inheritance via a default `{{block "shell"}}`.

For UI/dev-server changes, run an example with `:dev` and verify in a browser
before declaring done.

## Tone for code & comments

Match the project: short, declarative comments only when the *why* isn't
obvious. The spec and rules.md are the long-form reference — code stays lean.
Don't add backwards-compatibility shims unless explicitly asked.

## AI-coded provenance

This repository is AI-generated (noted in [README.md](README.md)). Treat
`docs/spec.md` and `docs/rules.md` as the source of truth and verify
non-trivial behavior against the code before relying on it.
