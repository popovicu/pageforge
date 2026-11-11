# Examples

End-to-end Pageforge sites. Each is a working `site_build` target plus a
`site_serve` dev target and a `tar_modes_test` that exercises the Cloudflare
Pages-friendly file modes.

| Example | Port | What it shows |
|---------|------|---------------|
| [basic](basic/) | 8080 | The minimum: markdown leaves, a collection, a tag index, static files. Has golden tests. |
| [blog](blog/) | 8081 | Adds chroma syntax highlighting, paginated post lists, and richer front matter (author, reading time). |
| [docs](docs/) | 8082 | Multiple guides, each its own collection. Padded `order` keys for deterministic chapter sort; `extra_metadata` on collections. |
| [portfolio](portfolio/) | 8083 | Three-level template inheritance (`base` → `layout` → `project`). Status badges and tech-stack tags. |

```sh
bazel build //examples/...        # build every example
bazel test  //examples/...        # run all integration tests
bazel run   //examples/blog:dev   # serve any example on its port
```

Every example uses only the public rules from `//rules:defs.bzl`. They are the
canonical references for how to wire those rules together.
