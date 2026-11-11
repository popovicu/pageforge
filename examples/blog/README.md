# blog

A small blog ("Forge Notes"). Builds on [basic](../basic/) by adding pagination,
syntax highlighting, and richer front matter.

```sh
bazel run //examples/blog:dev     # → http://localhost:8081
```

## What it adds over `basic`

- **Chroma syntax highlighting** via `site_chroma_css` (monokai). The rule
  emits `/css/chroma.css`; templates reference it from `<head>`.
- **Pagination** on the post collection (`paginate = 2`). Pageforge writes
  `/posts/index.html` plus `/posts/page/2/index.html`, etc., and exposes
  `.PrevURL` / `.NextURL` to the template.
- **Custom front-matter fields** (`author`, `reading_minutes`) read off
  `.Metadata` in `post.html`.
- **`extra_metadata` on the collection** — title and description pinned in
  `BUILD.bazel` so the listing page has chrome without a sidecar yaml file.
- **Yaml-driven home page** with a `featured` array iterated in `home.html`.
- **`sort_by = "date"`, `sort_order = "desc"`** — newest posts first.
- **Auto-generated TOC.** Three sibling posts demonstrate the markdown
  renderer's Table-of-contents behavior:
  - `auto-toc.md` — happy path: `## Table of contents` + H2/H3 sections →
    a `<details>` block linking every subsequent heading.
  - `no-toc.md` — `auto_toc: false` in front matter suppresses injection.
  - `custom-toc-marker.md` — `toc_marker` / `toc_summary` override the
    detected heading text and the disclosure label (also settable
    site-wide via `site_config`).

## Layout

```
content/      8 posts + home.yaml + about.yaml + BUILD
templates/    base, home, page, post, post_list, tag_list, tag_index + BUILD
static/       blog.css + BUILD
```

## Tests

`tar_modes_test` (reused from `//examples/basic`) plus `artifacts_test`
which asserts feed/sitemap/search shape and the auto-TOC behaviors above.
