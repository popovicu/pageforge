---
title: "Five template functions you actually use"
date: 2026-04-15
description: "A short tour of the helpers Pageforge attaches to every html/template render."
tags: [go, templates]
author: "Mira"
reading_minutes: 3
---

Go's `html/template` ships with a tiny standard library of functions. Pageforge
adds five more, and you reach for them surprisingly often:

```go
template.FuncMap{
    "formatTime": formatTime,
    "hasPrefix":  hasPrefix,
    "hasSuffix":  hasSuffix,
    "add":        add,
    "sub":        sub,
}
```

## Dates: `formatTime`

YAML front matter stores dates as ISO 8601 strings. Renderers serialize them
to RFC 3339 in the metadata sidecar. By the time they reach a template they
are strings like `2026-04-15T00:00:00Z` — readable, but not pretty.

```html
<time datetime="{{.Date}}">{{formatTime "Jan 2, 2006" .Date}}</time>
```

`formatTime` parses RFC 3339 and re-formats with Go's reference layout. On
parse failure it returns the value unchanged, so a typo in front matter never
takes the page down.

## Pagination: `add` and `sub`

Pagination math is the most boring use of `add` you will ever write, and also
the one that proves you needed it:

```html
{{range $i, $page := .Pages}}
  <li>{{add $i 1}}. {{$page.Title}}</li>
{{end}}
```

## Path matching: `hasPrefix` / `hasSuffix`

These come up most when a navigation block needs to know which section the
current URL lives under:

```html
<a class="nav {{if hasPrefix "/posts/" .URL}}active{{end}}" href="/posts/">Posts</a>
```

That is the whole library. Anything more elaborate belongs in the renderer,
not the template.
