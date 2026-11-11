# docs

A documentation site with multiple guides. Each guide is its own collection
with ordered chapters.

```sh
bazel run //examples/docs:dev     # → http://localhost:8082
```

## What it shows

- **Multiple sibling collections.** Two guides (`getting-started`,
  `advanced-patterns`) each live in their own package and produce their own
  `site_collection` index. `site_build` pulls them in via `collections = [...]`.
- **Deterministic chapter order without a custom rule.** Front matter carries
  `order: "01"` (string-padded) so `sort_by = "order", sort_order = "asc"`
  yields lexicographic-but-correct ordering. The chapter template renders
  `{{add $i 1}}` over the loop index for display.
- **`extra_metadata` per collection** sets the guide's title and description in
  `BUILD.bazel`, keeping the chapter markdown files free of layout chrome.
- **`guide_url` / `guide_title` in chapter front matter** lets the chapter
  template link back up to its parent guide without a separate lookup table.
- **Chroma highlighting** with the `github` style.

## Layout

```
content/                       home.yaml, about.yaml + BUILD
guides/getting-started/        3 chapters + BUILD (one site_collection)
guides/advanced-patterns/      3 chapters + BUILD (one site_collection)
templates/                     base, chapter, guide_index, home, page + BUILD
static/                        docs.css + BUILD
```

## Tests

`tar_modes_test`.
