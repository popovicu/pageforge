---
title: "Template inheritance, three levels deep"
description: "A base, a layout, and a leaf. How site_template chains compose them."
date: 2026-04-25
order: "01"
guide: "advanced-patterns"
guide_url: "/guides/advanced-patterns/"
guide_title: "Advanced patterns"
---

`site_template` exposes a single inheritance attribute: `extends`. Chains of
arbitrary depth fall out of stacking it.

## The chain

```python
site_template(name = "base",   src = "base.html")
site_template(name = "layout", src = "layout.html", extends = ":base")
site_template(name = "leaf",   src = "leaf.html",   extends = ":layout")
```

Behind the scenes, each `site_template` adds its source to a `depset` of
template files. The page-producing rules walk that chain, base first, and hand
the list to Go's `template.ParseFiles`.

## How the slots compose

Go's `html/template` resolves `{{block}}` / `{{define}}` lazily. The base
defines outer slots; intermediate templates fill some and re-expose others;
the leaf fills the innermost slots.

```html
<!-- base.html -->
<html><body>{{block "page" .}}default page{{end}}</body></html>

<!-- layout.html -->
{{define "page"}}<main class="layout">{{block "main" .}}{{end}}</main>{{end}}

<!-- leaf.html -->
{{define "main"}}<h1>{{.Title}}</h1>{{end}}
```

The leaf only knows about `main`. The base only knows about `page`. The
layout is the bridge — and is reusable across many leaves.

## When to reach for a deeper chain

Two levels (base → leaf) is enough for a small blog. Three becomes useful when
multiple leaves share structure that does not belong in the base — a
documentation layout with sidebar, a marketing layout without one, both
sitting on top of the same `<html>` skeleton.

Past three levels you are usually re-inventing partials and should ask
whether a shared template included via `{{template "name"}}` reads better.
