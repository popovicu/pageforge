---
title: "Installing Pageforge"
description: "Add Pageforge as a Bazel module dependency and verify the rules load."
date: 2026-04-20
order: "01"
guide: "getting-started"
guide_url: "/guides/getting-started/"
guide_title: "Getting started"
---

Pageforge ships as a Bazel module. Add it to your `MODULE.bazel`:

```python
bazel_dep(name = "pageforge", version = "0.1.0")
```

If you are tracking the source repository directly, add an `archive_override`
or `git_override` after the `bazel_dep`. The rules become available under
`@pageforge//rules:defs.bzl`.

## A smoke test

The fastest way to confirm the install worked is to load a single rule and
look at it:

```python
load("@pageforge//rules:defs.bzl", "site_template")

site_template(
    name = "smoke",
    src = "base.html",
)
```

If `bazel query //...` returns the target, you are wired up. The next chapter
turns this into a real page.
