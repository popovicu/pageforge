---
title: "When to write a macro"
description: "The repetition you feel writing many site_markdown rules is a feature, not a bug — until it isn't."
date: 2026-04-27
order: "03"
guide: "advanced-patterns"
guide_url: "/guides/advanced-patterns/"
guide_title: "Advanced patterns"
---

Pageforge's rules are intentionally low-level. Each `site_markdown` declares
one page and one URL. Each `site_collection` enumerates its pages by label.
For a small site this reads well — the BUILD file is a literal table of
contents.

For a docs site with twelve guides and six chapters each, the same pattern
becomes 72 nearly-identical `site_markdown` invocations. That is when you
write a macro.

## A worked example

```python
def chapter(name, order, guide, guide_url, guide_title):
    site_markdown(
        name = name,
        src = name + ".md",
        template = "//templates:chapter",
        url_path = guide_url + name + "/index.html",
    )

def guide(name, slug, title, description, chapters):
    [chapter(c.name, c.order, slug, "/guides/{}/".format(slug), title) for c in chapters]
    site_collection(
        name = "guide",
        pages = [":" + c.name for c in chapters],
        template = "//templates:guide_index",
        url_path = "/guides/" + slug,
        sort_by = "order",
        sort_order = "asc",
        extra_metadata = {"title": title, "description": description},
    )
```

Now each guide's BUILD file shrinks to a single `guide(...)` call.

## What you should not do

Resist hiding the page-by-page structure inside the macro. The advantage of
"BUILD file as table of contents" is real — a contributor can add a chapter
by adding a line. If `chapter()` becomes the place where front matter is
*generated* rather than written, you have lost the trail from BUILD file to
rendered output.

The right rule of thumb: macros may compose existing primitives, but should
not invent metadata that does not appear in the source files. Front matter
lives next to the prose it describes.
