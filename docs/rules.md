# Pageforge Rules Reference

All rules are loaded from `@pageforge//rules:defs.bzl`:

```starlark
load("@pageforge//rules:defs.bzl", "site_template", "site_static", "site_markdown", "site_page", "site_collection", "site_metadata_bundle", "site_tags", "site_feed", "site_sitemap", "site_search_index", "site_build", "site_zip", "site_serve")
```

## URL paths

A page's `url_path` is the path of its file inside the site tarball, so give
it an extension. Prefer `/posts/my-post/index.html`: static hosts serve it at
`/posts/my-post/` as `text/html`. An extensionless `/posts/my-post` ships as a
file with no extension, which strict hosts (Cloudflare Pages, for one) serve
without a Content-Type. Feeds, sitemaps and metadata bundles emit the
canonical form with the trailing `index.html` dropped; templates get the same
with the [`canonicalURL`](#template-functions) function.

## Dates and sorting

Unquoted YAML dates (`date: 2026-04-15`) reach templates and sidecars as
RFC 3339 strings (`2026-04-15T00:00:00Z`); quoted values stay as written.
Pageforge treats a value as a date when it is RFC 3339 or `YYYY-MM-DD`.

Rules with `sort_by` (`site_collection`, `site_tags`, `site_feed`,
`site_metadata_bundle`) compare that metadata field across pages. When every
non-empty value is a date, pages sort chronologically, so mixed UTC offsets
order correctly. Otherwise values compare as strings; use zero-padded keys
(`order: "01"`) for numeric order. Pages missing the field sort lowest. Equal
keys break ties on `url_path` ascending, so every listing orders them the
same way.

Aggregating rules accept an empty `pages` list: collections render one empty
listing page, tags render only the index, feeds have no items.

---

## site_template

```starlark
site_template(name, src, extends)
```

Defines a template with optional inheritance. Templates use Go's `html/template`
syntax. Child templates fill parent slots via `{{define "block_name"}}` /
`{{block "block_name" .}}` directives. Every file in a chain needs a distinct
file name: Go keys parsed templates by name, so a second `page.html` would
replace the first rather than extend it, and pageforge fails the build
instead.

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | Name | yes | Target name |
| `src` | Label | yes | Go `html/template` source file |
| `extends` | Label | no | Parent `site_template` target. This template's `{{define}}` blocks fill the parent's `{{block}}` slots |

**Example:**

```starlark
site_template(name = "base", src = "base.html")
site_template(name = "post", src = "post.html", extends = ":base")
```

**Providers returned:** `TemplateInfo`

---

## site_static

```starlark
site_static(name, srcs, url_prefix)
```

Pass-through rule for static files. Each file appears in the final site at
`url_prefix/filename`.

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | Name | yes | Target name |
| `srcs` | List of labels | yes | Static files (CSS, images, fonts, etc.) |
| `url_prefix` | String | yes | URL prefix (e.g. `"/css"`). Files appear at `prefix/filename` |

**Example:**

```starlark
site_static(
    name = "styles",
    srcs = ["style.css", "reset.css"],
    url_prefix = "/css",
)
```

**Providers returned:** `SiteOutputInfo`

---

## site_markdown

```starlark
site_markdown(name, src, template, url_path, site_config)
```

Parses a Markdown file with YAML front matter, renders it through a template
chain, and produces HTML. Also emits a `.meta.json` sidecar for use by
aggregation rules.

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | Name | yes | Target name |
| `src` | Label | yes | Markdown source file (`.md`) with optional YAML front matter |
| `template` | Label | yes | `site_template` target to render through |
| `url_path` | String | yes | URL path in the final site (e.g. `"/posts/my-post"`) |
| `site_config` | Label | no | YAML file merged into `Metadata` as a base (see [Site config](#site-config)) |

**Template context:** `PageContext` with `.Title`, `.Date`, `.Description`, `.Tags`, `.Content` (rendered HTML), `.URL`, `.Metadata` (full map).

**Markdown features:** Chroma syntax highlighting on code fences, GitHub
Flavored Markdown (pipe tables, `~~strikethrough~~`, `- [ ]` task lists,
auto-linked bare `http(s)://` URLs), `[^1]` footnotes, raw HTML passthrough
(sources are treated as trusted), plus an
[auto-generated TOC](#auto-generated-toc) under any `## Table of contents`
heading (configurable, opt-out per post).

**Front matter** opens with a first line of `---` and closes at the next line
that is exactly `---` (trailing whitespace and CRLF endings are fine). A file
without both lines has no front matter and is rendered whole.

**Output groups:**
- `html` — the rendered HTML file
- `metadata` — the `.meta.json` sidecar
- `content` — the rendered Markdown body alone, without the template (what
  `site_search_index` indexes)

**Example:**

```starlark
site_markdown(
    name = "my-post",
    src = "my-post.md",
    template = "//templates:post",
    url_path = "/posts/my-post/index.html",
)
```

**Providers returned:** `SiteOutputInfo`, `OutputGroupInfo`

### Auto-generated TOC

A literal `## Table of contents` heading in a post body is treated as a
TOC anchor. The renderer preserves the heading verbatim and inserts a
collapsible `<details>` block right after it that lists every subsequent
`<h2>` and `<h3>` with anchor links. Match is case-insensitive on the
trimmed heading text. Headings before the marker are not included; only
what comes after.

#### Configuration keys

All three keys live on `.Metadata` — read from the merged
`site_config` + per-page front matter. Set in `site_config` for a
site-wide value; override in front matter for a single post. Front
matter wins on conflict.

| Key | Type | Default | Effect |
|-----|------|---------|--------|
| `auto_toc` | bool | `true` | When `false`, suppresses TOC injection for the page. The literal heading still renders |
| `toc_marker` | string | `Table of contents` | Heading text the renderer matches against (case-insensitive, trimmed). The H2 in your markdown source must use this exact text |
| `toc_summary` | string | `Open ` + `toc_marker` | Label inside the rendered `<summary>`. Tracks `toc_marker` automatically when only the marker is overridden |

#### Common customizations

**Rename site-wide (e.g. for a non-English blog).** Add the keys to
`site.yaml` and the matching markdown heading in every post:

```yaml
# site.yaml
site:
  base_url: "https://example.es"
toc_marker: "Tabla de contenidos"
toc_summary: "Abrir tabla de contenidos"
```

```markdown
## Tabla de contenidos
```

`toc_summary` is optional — leaving it out yields
`Abrir Tabla de contenidos` automatically.

**Override for one post.** Front matter wins, so a single post can use a
different heading text without touching `site.yaml`:

```yaml
---
title: "Long-form essay"
toc_marker: "Contents"
toc_summary: "Jump to a section"
---
```

```markdown
## Contents
```

**Opt a single post out entirely.** Keep the literal heading visible but
suppress the generated block:

```yaml
---
title: "Short note"
auto_toc: false
---
```

Posts without any matching heading are unaffected — no opt-out needed
and no markup is added.

#### Emitted markup

```html
<h2 id="table-of-contents">Table of contents</h2>
<details class="pageforge-toc"><summary>Open Table of contents</summary>
<ul>
<li><a href="#first-section">First section</a>
<ul>
<li><a href="#a-nested-point">A nested point</a></li>
</ul>
</li>
<li><a href="#second-section">Second section</a></li>
</ul>
</details>
```

The collapsible UX comes from native `<details>` / `<summary>` — no JS.
Anchor ids are whatever `parser.WithAutoHeadingID()` already emits for
each heading, so the TOC links and the heading anchors stay in lockstep
across renames or content edits.

#### Styling

Pageforge ships no CSS — every consumer styles the TOC to match its own
typography. The `pageforge-toc` class on the outer `<details>` is the
single stable hook; the inner `<summary>`, `<ul>`, `<li>`, and `<a>` are
plain HTML and inherit whatever the surrounding article styles set.

A minimal starter (place in your existing post stylesheet):

```css
.pageforge-toc {
  margin: 0 0 32px;
  padding: 16px 20px;
  background: #f5f5f5;
  border: 1px solid #ddd;
  border-radius: 6px;
}
.pageforge-toc summary {
  cursor: pointer;
  font-weight: 600;
  list-style: none;
}
.pageforge-toc summary::-webkit-details-marker { display: none; }
.pageforge-toc summary::before {
  content: "▸ ";
  display: inline-block;
  transition: transform 0.15s ease;
}
.pageforge-toc[open] summary::before { transform: rotate(90deg); }
.pageforge-toc ul { margin: 12px 0 0 20px; }
.pageforge-toc li { margin-bottom: 4px; }
```

Two presentational choices worth knowing:

- **Always-open list** — drop the `<details>` collapsibility entirely by
  styling `.pageforge-toc summary { display: none }` and
  `.pageforge-toc[open], .pageforge-toc { display: block }`. Or, if you
  want the open state by default but still allow toggle, set
  `.pageforge-toc[open]` styles and let users click to collapse.
- **Floating sidebar** — `position: sticky` on `.pageforge-toc` plus a
  grid/flex layout on the article container turns the TOC into a
  side-panel without any markup change.

The renderer doesn't auto-open the `<details>`; if you want the TOC
expanded by default, add the `[open]` attribute via a small post-render
script or accept that users click once.

#### Reference & limits

- TOC walks H2 and H3 only. H4 and below are deliberately omitted; if a
  post needs deeper navigation, write the TOC by hand and set
  `auto_toc: false`.
- Source headings always render as-is — the renderer never strips or
  rewrites them. The marker H2 is preserved alongside the generated
  block.
- A worked end-to-end demo lives at `examples/blog/content/auto-toc.md`,
  the opt-out at `examples/blog/content/no-toc.md`, and the marker
  override at `examples/blog/content/custom-toc-marker.md`.

---

## site_page

```starlark
site_page(name, data, template, url_path, site_config, extra_data)
```

Renders a template chain with data from a YAML file. For non-Markdown pages
like landing pages or about pages.

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | Name | yes | Target name |
| `data` | Label | yes | YAML data file (`.yaml` or `.yml`). All keys become template variables |
| `template` | Label | yes | `site_template` target to render through |
| `url_path` | String | yes | URL path in the final site (e.g. `"/about"`) |
| `site_config` | Label | no | YAML file merged into `Metadata` as a base (see [Site config](#site-config)) |
| `extra_data` | List of labels | no | Additional YAML files merged into `Metadata` between `site_config` and `data`. Files merge in list order; the primary `data` file still wins on key collisions. Designed to feed `site_metadata_bundle` outputs (homepage widgets, featured slices, etc.) into a page without hand-editing its YAML |

**Template context:** `PageContext` with `.Title`, `.Date`, `.Description`, `.URL`, `.Metadata` (full map). No `.Content` field.

**Example:**

```starlark
site_page(
    name = "about",
    data = "about.yaml",
    template = "//templates:page",
    url_path = "/about",
)
```

**Providers returned:** `SiteOutputInfo`

---

## site_collection

```starlark
site_collection(name, pages, template, url_path, paginate, sort_by, sort_order, extra_metadata)
```

Aggregates metadata from page targets, sorts and paginates, and renders listing
pages. Produces a `.tar` archive; pass it to `site_build` via the `collections`
attribute.

| Attribute | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `name` | Name | yes | | Target name |
| `pages` | List of labels | yes | | `site_markdown` targets to aggregate. Must provide `OutputGroupInfo` with `metadata` group |
| `template` | Label | yes | | `site_template` target for listing pages |
| `url_path` | String | yes | | Base URL path (e.g. `"/posts"`, or `"/"` for a root listing). Paginated pages at `/posts/page/2`, etc. Trailing slashes are ignored |
| `paginate` | Integer | no | `0` | Items per page. `0` = all on one page |
| `sort_by` | String | no | `"date"` | Metadata field to sort by |
| `sort_order` | String | no | `"desc"` | `"asc"` or `"desc"`. See [Dates and sorting](#dates-and-sorting) |
| `extra_metadata` | Dict (string→string) | no | | Additional metadata passed to the template |
| `site_config` | Label | no | | YAML file merged into `Metadata` as a base (see [Site config](#site-config)) |

**Template context:** `CollectionContext` with `.Pages` (list of `PageContext`), `.CurrentPage`, `.TotalPages`, `.PrevURL`, `.NextURL`, `.URL`, `.Metadata`.

**Output groups:**
- `urls` — a `urls.json` sidecar listing every URL the tar contains. Consumed by `site_sitemap` via the `collections` attribute.

**Example:**

```starlark
site_collection(
    name = "all_posts",
    pages = [":post1", ":post2"],
    template = "//templates:post_list",
    url_path = "/posts",
    paginate = 10,
    sort_by = "date",
    sort_order = "desc",
)
```

---

## site_metadata_bundle

```starlark
site_metadata_bundle(name, pages, key, filter, sort_by, sort_order, limit, fields)
```

Aggregates page metadata into a YAML data artifact that another rule
consumes. Today's only consumer is `site_page` via its
[`extra_data`](#site_page) attribute, which merges the bundle into the
page's `Metadata` so templates can iterate `.Metadata.<key>`.

### Why this rule exists

Pageforge's other aggregators either render their own page
(`site_collection`, `site_tags`) or emit a terminal artifact you don't
re-consume (`site_feed`, `site_sitemap`, `site_search_index`). What was
missing was an aggregator that emits *re-consumable data* — so a page
template can be driven by a slice of the catalog without owning the
selection logic.

The motivating case is a homepage with a "Featured" section (filter
`featured = true`) and a "Recent posts" section (sort by date, take the
top N). Both are derivative views of the same catalog, both should track
the catalog automatically. Without this rule, the alternatives are:

- **Hand-curate the lists** in `home.yaml`. Catalog and page drift —
  flipping `featured = True` on a post also requires editing the
  homepage YAML, or the post never surfaces.
- **Use `site_collection`.** But `site_collection` owns the *whole
  rendered page* — paginated, listing-shaped — so a homepage with a
  hero plus multiple curated slices plus a CTA doesn't fit, and you
  can't have two collections drive one page anyway.
- **Concatenate YAMLs in a shell `genrule`.** Works, but no filtering or
  sorting, and the output format is welded to one specific consumer.

`site_metadata_bundle` is the clean factoring: it does the
"select + sort + project" half of `site_collection`, writes the result
as YAML, and stops. The consuming rule decides the rendering.

### What it composes into

- One page can pull in N bundles — Featured + Recent + a future
  "latest in tag X" — because each bundle is a YAML overlay on the
  page's `Metadata`. `site_page`'s `extra_data` merges them in order.
- Templates address entries as `.Metadata.<key>` and never see the
  filter logic. Catalog metadata stays the source of truth; the bundle
  is the seam between "rows in the catalog" and "data the page sees."
- Flipping a flag in the catalog (e.g. `featured = True` in a
  `posts.bzl` entry) re-runs the bundle, which re-runs the page. No
  template touches, no YAML hand-editing.

| Attribute | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `name` | Name | yes | | Target name |
| `pages` | List of labels | yes | | Targets exposing an `OutputGroupInfo` `metadata` group of `.meta.json` sidecars (today, only `site_markdown`) |
| `key` | String | no | target name | Top-level YAML key wrapping the entry list. Templates address entries as `.Metadata.<key>` |
| `filter` | String | no | `""` | Metadata field name; entries are kept only if its value is truthy (bool `true`, non-empty string, non-zero number, non-empty list/map). Empty = no filter |
| `sort_by` | String | no | `"date"` | Metadata field to sort by |
| `sort_order` | String | no | `"desc"` | `"asc"` or `"desc"` |
| `limit` | Integer | no | `0` | Maximum entries. `0` = unlimited |
| `fields` | List of strings | no | all | Metadata fields to project into each entry. Empty includes everything from the sidecar. A synthesized `url` field (canonical URL: `url_path` with any trailing `/index.html` stripped) is always included so templates can link without restating the strip |

**Output:** a single `<name>.yaml` file shaped as

```yaml
<key>:
  - title: "..."
    description: "..."
    date: "..."
    url: "/posts/.../"
```

**Example:**

```starlark
site_metadata_bundle(
    name = "featured_posts",
    pages = _ALL_POSTS,
    filter = "featured",
    sort_by = "date",
    sort_order = "desc",
    fields = ["title", "description", "date"],
)

site_page(
    name = "home",
    data = "home.yaml",
    extra_data = [":featured_posts"],
    template = "//themes:home",
    url_path = "/index.html",
)
```

The home template iterates `.Metadata.featured_posts` as `[{title, description, date, url}, ...]`.

**Providers returned:** `DefaultInfo` only — the bundle is data, not a site artifact. `site_page`'s `extra_data` reads it as a plain file; the bundle never enters `site_build`'s intake and so cannot accidentally land in the published tarball.

---

## site_tags

```starlark
site_tags(name, pages, tag_template, index_template, url_prefix, sort_by, sort_order, extra_metadata)
```

Groups pages by their `tags` front-matter field (a list, or a single string
for one tag), renders a listing page per tag and an index page listing all
tags. Produces a `.tar` archive; pass it to
`site_build` via the `collections` attribute.

| Attribute | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `name` | Name | yes | | Target name |
| `pages` | List of labels | yes | | `site_markdown` targets to aggregate. Must provide `OutputGroupInfo` with `metadata` group |
| `tag_template` | Label | yes | | `site_template` target for per-tag listing pages |
| `index_template` | Label | yes | | `site_template` target for the tag index page |
| `url_prefix` | String | yes | | URL prefix (e.g. `"/tags"`). Individual tags at `/tags/<tag-name>`. Each tag becomes a path segment, so a tag that is empty, `.`, `..`, or contains `/` or `\` fails the build |
| `sort_by` | String | no | `"date"` | Metadata field to sort pages within each tag by |
| `sort_order` | String | no | `"desc"` | `"asc"` or `"desc"` |
| `extra_metadata` | Dict (string→string) | no | | Additional metadata passed to templates |
| `site_config` | Label | no | | YAML file merged into `Metadata` as a base (see [Site config](#site-config)) |

**Per-tag template context:** `TagPageContext` with `.Tag` (string), `.Pages` (list of `PageContext`), `.URL`, `.Metadata`. `.Title` returns the tag name.

**Index template context:** `TagIndexContext` with `.Tags` (list of `TagIndexEntry` with `.Tag`, `.Count`, `.URL`), `.URL`, `.Metadata`.

**Output groups:**
- `urls` — a `urls.json` sidecar listing every URL the tar contains. Consumed by `site_sitemap` via the `collections` attribute.

**Example:**

```starlark
site_tags(
    name = "all_tags",
    pages = [":post1", ":post2"],
    tag_template = "//templates:tag_list",
    index_template = "//templates:tag_index",
    url_prefix = "/tags",
)
```

---

## site_feed

```starlark
site_feed(name, pages, title, description, base_url, url_path, sort_by, sort_order, limit, author)
```

Renders an RSS 2.0 feed XML file from a list of `site_markdown` targets.

| Attribute | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `name` | Name | yes | | Target name |
| `pages` | List of labels | yes | | Targets exposing an `OutputGroupInfo` `metadata` group of `.meta.json` sidecars. Today only `site_markdown` qualifies; `site_page` emits HTML only and is not feed-eligible. Any future rule that writes a sidecar plugs in unchanged |
| `title` | String | yes | | Channel title |
| `description` | String | no | `""` | Channel description |
| `base_url` | String | yes | | Site base URL (e.g. `"https://example.com"`). Item links are `base_url + url_path` |
| `url_path` | String | yes | | URL path of the feed in the final site (e.g. `"/rss.xml"`) |
| `sort_by` | String | no | `"date"` | Metadata field to sort by |
| `sort_order` | String | no | `"desc"` | `"asc"` or `"desc"` |
| `limit` | Integer | no | `0` | Maximum number of items. `0` = unlimited |
| `author` | String | no | `""` | Channel-level author. Optional. A value with an email address (`"me@example.com (Me)"`) goes in `<managingEditor>`, as RSS 2.0 requires; a plain name goes in `<dc:creator>`. Per-item `author` front matter follows the same rule (`<author>` or `<dc:creator>`) |

Channel `<lastBuildDate>` is the latest item date — the renderer never reads
the wall clock, so the feed is reproducible. It is omitted when the feed has
no items. Items without a parseable `date`
field (RFC 3339 or `YYYY-MM-DD`) are silently skipped; an undated entry has
no place in a feed. They are dropped before `limit` applies, so `limit`
counts dated items only.

**Example:**

```starlark
site_feed(
    name = "rss",
    pages = [":welcome", ":hello"],
    title = "Forge Notes",
    description = "Posts on the forge.",
    base_url = "https://example.com",
    url_path = "/rss.xml",
    limit = 20,
)
```

**Providers returned:** `SiteOutputInfo` (one entry, `kind = "feed"`)

---

## site_sitemap

```starlark
site_sitemap(name, pages, collections, base_url, url_path)
```

Generates a `sitemap.xml` listing every URL produced by the supplied targets.
Mirrors `site_build`'s two-channel intake: `pages` for single-file producers,
`collections` for aggregator outputs.

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | Name | yes | Target name |
| `pages` | List of labels | no | Single-file page producers (`site_markdown`, `site_page`). Entries marked `kind="static"` (i.e. `site_static` assets) are skipped — sitemaps list pages, not CSS/images. URLs come from `SiteOutputInfo.entries`; when a dep also exposes an `OutputGroupInfo` `metadata` group (today, only `site_markdown` does), its `.meta.json` sidecars feed `<lastmod>` |
| `collections` | List of labels | no | Aggregator targets (`site_collection`, `site_tags`) exposing a `urls` `OutputGroupInfo` group with a `urls.json` sidecar |
| `base_url` | String | yes | Site base URL (e.g. `"https://example.com"`). Each emitted `<loc>` is `base_url + url_path` |
| `url_path` | String | yes | URL path of the sitemap in the final site (e.g. `"/sitemap.xml"`) |

`<lastmod>` is emitted as the W3C Datetime date-only form (`YYYY-MM-DD`,
what Google recommends without a precise modification time) for any page
whose metadata sidecar carries a parseable `date` field (RFC 3339 or
`YYYY-MM-DD`). Pages without a sidecar — `site_page`, `site_static`,
collection URLs — get a `<url>` entry with no `<lastmod>`. No `<priority>`
elements are emitted (deprecated by Google). URLs are merged by canonical
form (`/x/` and `/x/index.html` are one entry); entries are sorted
lexicographically for reproducibility.

**Example:**

```starlark
site_sitemap(
    name = "sitemap",
    pages = [":home", ":about", ":welcome"],
    collections = [":all_posts", ":all_tags"],
    base_url = "https://example.com",
    url_path = "/sitemap.xml",
)
```

**Providers returned:** `SiteOutputInfo` (one entry, `kind = "static"`)

---

## site_search_index

```starlark
site_search_index(name, pages, fields, include_content, url_path)
```

Reads `.meta.json` sidecars from a list of `site_markdown` targets and emits a
JSON search index. The consumer ships their own JS to fetch the index and
feed it into Fuse.js / MiniSearch / etc. Pageforge does not bundle JS.

| Attribute | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `name` | Name | yes | | Target name |
| `pages` | List of labels | yes | | `site_markdown` targets to index. Must provide `OutputGroupInfo` with a `metadata` group (and `content`, or failing that `html`, when `include_content = True`) |
| `fields` | List of strings | no | `[]` | Metadata field names to copy from each page into the index. `url` is always included |
| `include_content` | Bool | no | `False` | When `True`, additionally include each page's body text as a `body` field: the Markdown content without the surrounding template, with tags, scripts, styles and comments stripped |
| `url_path` | String | yes | | URL path of the index in the final site (e.g. `"/search.json"`) |

**Output shape:** a JSON array. Each item has `url` plus the requested
fields. With `include_content = True`, each item also has `body`.

**Example:**

```starlark
site_search_index(
    name = "search",
    pages = [":welcome", ":hello"],
    fields = ["title", "description", "tags"],
    url_path = "/search.json",
)
```

**Providers returned:** `SiteOutputInfo` (one entry, `kind = "static"`)

---

## site_build

```starlark
site_build(name, deps, static, collections)
```

Assembles all site outputs into a single `.tar` archive. This is the final
build target for a Pageforge site.

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | Name | yes | Target name. Produces `<name>.tar` |
| `deps` | List of labels | no | Targets providing `SiteOutputInfo` (pages) |
| `static` | List of labels | no | Targets providing `SiteOutputInfo` (static files) |
| `collections` | List of labels | no | Targets producing `.tar` archives (`site_collection`, `site_tags`) |

**How it works:** Single-file targets (`deps`, `static`) are collected via
`PackageFilesInfo` with a `dest_src_map` keyed by URL path. Collection tars
are passed to `pkg_tar`'s `deps` which unpacks and merges them. The final
output is a single tar with all files at their URL paths.

Two `deps` / `static` entries that resolve to the same URL with different
files fail analysis. Collisions involving a collection tar (a page at
`/posts/index.html` next to a collection at `/posts`, or two collections
sharing a path) fail the build too, through a validation action; left to
itself, `pkg_tar` would keep the first copy and only print a warning.

**Example:**

```starlark
site_build(
    name = "site",
    deps = [":about", ":hello"],
    static = ["//static:styles"],
    collections = [":all_posts"],
)
```

---

## site_zip

```starlark
site_zip(name, tar)
```

Repacks a `site_build`'s `.tar` archive as a `.zip` with the same files at the
same paths. Use for deploy targets that require zip (e.g. Cloudflare Pages).
The `tar` attribute names the source archive explicitly: this rule does not
build a site from scratch, it rewraps an existing tar.

| Attribute | Type | Required | Description |
|-----------|------|----------|-------------|
| `name` | Name | yes | Target name. Produces `<name>.zip` |
| `tar` | Label | yes | A `site_build` target (its `.tar` output is the input) |

**How it works:** Runs a shell action that extracts the tar into a temp
directory and re-archives it with `zip`. Requires `tar` and `zip` on the
build host's `PATH` (universally available on macOS and standard Linux
distros).

**Example:**

```starlark
site_build(name = "site", ...)

site_zip(
    name = "site_zip",
    tar = ":site",
)
```

`bazel build //:site_zip` materializes `bazel-bin/site_zip.zip`.

---

## site_serve

```starlark
site_serve(name, site, port)
```

Creates a `bazel run` target that serves a built site locally over HTTP.
Extracts the site tar into a temporary directory, removed when the server
stops, and starts a file server. It listens on loopback only; pass
`-- --host=0.0.0.0` to reach it from other devices.

| Attribute | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `name` | Name | yes | | Target name |
| `site` | Label | yes | | A `site_build` target producing a `.tar` archive |
| `port` | Integer | no | `8080` | Port to listen on |

**Example:**

```starlark
site_serve(
    name = "dev",
    site = ":site",
    port = 8080,
)
```

**Usage:** `bazel run //my/site:dev`

---

## Site config

`site_markdown`, `site_page`, `site_collection`, and `site_tags` accept an
optional `site_config` label pointing at a YAML file. The renderer loads it
and uses it as a base `Metadata` map; per-page data overlays on top — the
overlay wins on key collisions. Lists are replaced wholesale, not
concatenated. When both sides hold a map at the same key, the merge recurses
one level (so a top-level `site:` block with sub-keys merges sensibly).

The intended use is site-wide values that every page needs in templates —
`base_url`, `author`, `twitter_handle`, default OG image, etc. Group them
under a single top-level key (commonly `site:`) so per-page front matter
doesn't have to dance around them, and so templates address them via
`{{.Metadata.site.base_url}}` rather than a flat namespace.

```yaml
# site.yaml
site:
  base_url: "https://example.com"
  author: "Alice"
  twitter_handle: "@example"
```

```starlark
filegroup(name = "site_config", srcs = ["site.yaml"])

site_markdown(
    name = "my-post",
    src = "my-post.md",
    template = "//templates:post",
    url_path = "/posts/my-post/index.html",
    site_config = ":site_config",
)
```

**Why a `filegroup` and not a custom rule?** The attribute is just
`attr.label(allow_single_file = [".yaml", ".yml"])` — anything that resolves
to one YAML file works. `filegroup` is the lightest Bazel primitive that
fits, and it earns its keep in two ways:

- **Cross-package visibility.** A bare `srcs = ["site.yaml"]` only works
  inside the declaring package. Wrapping the file in a named filegroup with
  `visibility = [...]` exposes it to other packages without scattering
  `exports_files` declarations.
- **Prod/dev swap with `select()`.** A common pattern is to point the
  filegroup at `site.yaml` by default and `site.dev.yaml` under a
  `--define=env=dev` `config_setting`, so `bazel run //:dev` rewrites
  canonical / OG URLs to localhost without changing any consumer:

  ```starlark
  config_setting(name = "env_dev", define_values = {"env": "dev"})

  filegroup(
      name = "site_config",
      srcs = select({
          ":env_dev": ["site.dev.yaml"],
          "//conditions:default": ["site.yaml"],
      }),
      visibility = ["//visibility:public"],
  )
  ```

A custom rule would only earn its keep if it had real work to do (schema
validation, multi-file merge, emitting a provider). Pageforge's renderers
just want a label that resolves to one YAML file; `filegroup` is the right
amount of machinery.

The attribute is purely additive: when unset, the renderer behaves exactly
as before — existing builds with no `site_config` keep their current
output, byte-for-byte. Per-page overrides remain in front matter (for
`site_markdown`), the data YAML (for `site_page`), or `extra_metadata`
(for `site_collection` and `site_tags`).

The merged result is what the page template sees as `.Metadata`. The
`site_markdown` `.meta.json` sidecar keeps only the per-page front matter —
site-wide keys live in one place and aren't duplicated into every sidecar.

---

## Template functions

Every `html/template` rendered by pageforge (via `site_page`, `site_markdown`,
`site_collection`, `site_tags`, and any template they extend) has the following
built-in functions registered. All are pure, total (no panics on bad input),
and pipe-friendly — the piped value is the last argument.

| Function | Signature | Description |
|----------|-----------|-------------|
| `formatTime` | `formatTime(layout, value string) string` | Parses `value` as RFC 3339 or `YYYY-MM-DD` and formats using Go's reference layout. Returns `value` unchanged on parse failure |
| `canonicalURL` | `canonicalURL(path string) string` | Roots `path` and drops a trailing `index.html`: `/posts/x/index.html` → `/posts/x/`. Use it for links to pages from listings |
| `hasPrefix` | `hasPrefix(prefix, s string) bool` | `strings.HasPrefix(s, prefix)` |
| `hasSuffix` | `hasSuffix(suffix, s string) bool` | `strings.HasSuffix(s, suffix)` |
| `add` | `add(a, b int) int` | Integer addition |
| `sub` | `sub(a, b int) int` | Integer subtraction |

**Example:**

```go-html-template
<time datetime="{{.Metadata.pub_datetime}}">
  {{.Metadata.pub_datetime | formatTime "January 2, 2006"}}
</time>

{{if .URL | hasPrefix "/series/"}}<nav>...</nav>{{end}}

Page {{.CurrentPage}} of {{.TotalPages}} — next is {{add .CurrentPage 1}}
```
