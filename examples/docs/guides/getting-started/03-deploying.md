---
title: "Deploying the tarball"
description: "Pageforge's only deliverable is a tar of the rendered site. Here is what to do with it."
date: 2026-04-22
order: "03"
guide: "getting-started"
guide_url: "/guides/getting-started/"
guide_title: "Getting started"
---

Pageforge produces exactly one artifact: a tar archive at
`bazel-bin/<your-target>.tar`. Every deployment story is a variation on
"unpack this somewhere a web server can read it."

## Local preview

The `site_serve` rule wraps a small dev server that extracts the tar in
memory and serves it on a port:

```python
load("@pageforge//rules:defs.bzl", "site_serve")

site_serve(
    name = "dev",
    site = ":site",
    port = 8080,
)
```

Then `bazel run //:dev` and open `http://localhost:8080/`.

## Static hosts

For Cloudflare Pages, Netlify, GitHub Pages, S3, or any other static host:

```bash
bazel build //:site
mkdir _public
tar xf bazel-bin/site.tar -C _public
# upload _public/ with your tool of choice
```

## A note on Content-Type

Pageforge writes every regular file with mode 0644 and every directory with
0755. This is on purpose: some CDNs (notably Cloudflare) refuse to set a
Content-Type header on files with the executable bit set, which silently
breaks social previews. The basic example contains a `tar_modes_test.sh`
that asserts this invariant — borrow it for your own site if you care.
