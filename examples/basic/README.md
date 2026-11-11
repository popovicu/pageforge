# basic

The smallest end-to-end site. Read this first.

```sh
bazel run //examples/basic:dev    # → http://localhost:8080/posts/
```

There is no home page; the site is just the posts, tags and about page.

## What it shows

- `site_template` chain: every leaf template extends `base.html`.
- `site_markdown` for two posts (`hello`, `goodbye`) and `site_page` (yaml-driven) for `about`.
- `site_collection` builds `/posts/` from those two markdown pages.
- `site_tags` builds `/tags/` (index) and `/tags/<tag>/` (per-tag list).
- `site_static` ships `styles.css`.
- `site_build` stitches `deps + collections + static` into one tarball.

## Layout

```
content/      hello.md, goodbye.md, about.yaml + BUILD
templates/    base, page, post, post_list, tag_list, tag_index + BUILD
static/       styles.css + BUILD
testdata/     golden output for the integration test
```

## Tests

- `testdata:*_test` — diff each rendered page and `.meta.json` sidecar
  against its golden file in `testdata/` (`bazel test //examples/basic/...`).
- `tar_modes_test` — assert 0644 files / 0755 dirs (Cloudflare Pages-safe).

The tar-modes script is `exports_files`'d so the other examples can reuse it.
