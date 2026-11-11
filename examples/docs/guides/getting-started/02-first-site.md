---
title: "Your first site"
description: "A minimal three-file site: one template, one Markdown post, one site_build."
date: 2026-04-21
order: "02"
guide: "getting-started"
guide_url: "/guides/getting-started/"
guide_title: "Getting started"
---

The smallest interesting Pageforge site is three files: a template, a piece of
Markdown content, and a `site_build` to glue them together.

## Template

```html
<!-- templates/base.html -->
<!DOCTYPE html>
<html>
<head><title>{{template "title" .}}</title></head>
<body>{{template "content" .}}</body>
</html>
```

```python
# templates/BUILD.bazel
load("@pageforge//rules:defs.bzl", "site_template")

site_template(name = "base", src = "base.html")
```

## Content

```markdown
---
title: "Hello"
---

# Hello

A first post.
```

```python
# content/BUILD.bazel
load("@pageforge//rules:defs.bzl", "site_markdown")

site_markdown(
    name = "hello",
    src = "hello.md",
    template = "//templates:base",
    url_path = "/hello/index.html",
)
```

## Build

```python
# BUILD.bazel
load("@pageforge//rules:defs.bzl", "site_build")

site_build(
    name = "site",
    deps = ["//content:hello"],
)
```

Now `bazel build //:site` produces `bazel-bin/site.tar`, which is your entire
site as a tar archive. Untar it anywhere, point a static host at the contents,
and you are done.
