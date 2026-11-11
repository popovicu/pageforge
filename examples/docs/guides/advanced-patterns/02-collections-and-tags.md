---
title: "Collections, tags, and the metadata sidecar"
description: "Why every page emits a .meta.json, and how site_collection consumes it."
date: 2026-04-26
order: "02"
guide: "advanced-patterns"
guide_url: "/guides/advanced-patterns/"
guide_title: "Advanced patterns"
---

`site_markdown` produces two files: the rendered HTML and a `.meta.json`
sidecar. The sidecar is exposed as the `metadata` output group, and it is the
contract that lets `site_collection` and `site_tags` aggregate pages without
re-parsing the source.

## What lands in the sidecar

Every key from the front matter, plus a synthesized `url_path`:

```json
{
  "title": "Hello",
  "date": "2026-04-15T00:00:00Z",
  "tags": ["test", "hello"],
  "url_path": "/posts/hello/index.html"
}
```

YAML dates round-trip through Go's `yaml.v3` and end up as RFC 3339 strings.
That is the format the `formatTime` template function expects, so dates always
work cleanly in collection templates.

## Sorting by an arbitrary field

`sort_by` accepts any front-matter key. For ordered chapters, set
`sort_by = "order"` and use string-typed values like `"01"`, `"02"`,
`"10"` — sorting is lexicographic, not numeric.

```python
site_collection(
    name = "guide",
    pages = [":ch1", ":ch2", ":ch3"],
    template = "//templates:guide_index",
    url_path = "/guides/foo",
    sort_by = "order",
    sort_order = "asc",
)
```

## Pagination

Set `paginate = N` to break the listing into pages of N items each. The first
page lands at `<url_path>/index.html`, subsequent pages at
`<url_path>/page/2/index.html`, `<url_path>/page/3/index.html`, and so on.
The collection template receives `.PrevURL`, `.NextURL`, `.CurrentPage`, and
`.TotalPages` to render the pager.
