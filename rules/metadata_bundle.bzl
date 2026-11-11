"""site_metadata_bundle rule.

Aggregates page metadata into a YAML data artifact that another rule
consumes. The missing slot in pageforge's rule matrix: an aggregator
that produces *re-consumable data*, neither a rendered page (like
site_collection / site_tags) nor a terminal artifact (like site_feed /
site_sitemap / site_search_index).

# What it solves

The motivating case is a homepage with a "Featured" section and a
"Recent posts" section. Both are derivative views of the same catalog —
Featured filters by `featured = True`, Recent sorts by date and slices
the top N — and both should track the catalog automatically. Without
this rule the alternatives are all bad:

  - Hand-curate the lists in the page's YAML data. The catalog and the
    page drift: flipping `featured = True` on a post also requires
    editing home.yaml, or the post never surfaces.
  - Use site_collection. But site_collection owns the *whole rendered
    page* — paginated, listing-shaped — so a homepage with a hero plus
    multiple curated slices plus a CTA doesn't fit, and you can't have
    two collections drive one page anyway.
  - Reach into the build with shell concatenation (LFG's
    lfg_series_bundle does this for a similar problem). Works, but no
    filter/sort and the output format is welded to whatever the
    consumer happens to expect.

site_metadata_bundle is the clean factoring: it does the
"select + sort + project" half of site_collection, writes the result
as YAML, and stops there. The consuming rule (today, site_page via its
extra_data attribute) picks the rendering.

# Why it composes

  - One page can pull in N bundles — Featured + Recent + a future
    "latest in tag X" — because each bundle is just a YAML overlay on
    the page's Metadata. site_page's extra_data merges them in order.
  - Templates address entries as `.Metadata.<key>` and never see the
    filter logic. Catalog metadata stays the source of truth; the
    bundle is the seam that turns "rows in the catalog" into "data the
    page sees."
  - Flipping a flag in the catalog (e.g. `featured = True` in a
    posts.bzl entry) re-runs the bundle, which re-runs the page. No
    template touches, no YAML hand-editing.

# Output shape

A single YAML file with one top-level key (default = target name)
wrapping a list of entry dicts. Each entry contains the requested
`fields` plus a synthesized `url` — the entry's `url_path` with any
trailing `/index.html` stripped — so templates can write
`<a href="{{.url}}">` without restating the canonicalization on every
iteration.
"""

load(":common.bzl", "metadata_files")

def _site_metadata_bundle_impl(ctx):
    out = ctx.actions.declare_file(ctx.label.name + ".yaml")

    meta_files = metadata_files(ctx.attr.pages)

    key = ctx.attr.key if ctx.attr.key else ctx.label.name

    args = ctx.actions.args()
    args.add("--meta-files", ",".join([f.path for f in meta_files]))
    args.add("--key", key)
    args.add("--output", out)
    args.add("--sort-by", ctx.attr.sort_by)
    args.add("--sort-order", ctx.attr.sort_order)
    if ctx.attr.filter:
        args.add("--filter", ctx.attr.filter)
    if ctx.attr.limit > 0:
        args.add("--limit", str(ctx.attr.limit))
    if ctx.attr.fields:
        args.add("--fields", ",".join(ctx.attr.fields))

    ctx.actions.run(
        executable = ctx.executable._builder,
        arguments = [args],
        inputs = meta_files,
        outputs = [out],
        mnemonic = "MetadataBundle",
        progress_message = "Bundling metadata %{label}",
    )

    # DefaultInfo is enough: site_page's extra_data is a label_list that
    # reads ctx.files (the default outputs). The bundle is data, not a
    # site artifact, so it stays out of SiteOutputInfo and never enters
    # site_build's intake.
    return [DefaultInfo(files = depset([out]))]

site_metadata_bundle = rule(
    implementation = _site_metadata_bundle_impl,
    doc = "Aggregate page metadata into a single YAML artifact for consumption " +
          "by site_page's extra_data attribute.",
    attrs = {
        "pages": attr.label_list(
            mandatory = True,
            doc = "Targets that expose an OutputGroupInfo 'metadata' group of " +
                  ".meta.json sidecars (today, only site_markdown qualifies; " +
                  "any future rule that writes a sidecar plugs in unchanged).",
        ),
        "key": attr.string(
            doc = "Top-level YAML key wrapping the entry list. Defaults to the " +
                  "target name; templates address entries as .Metadata.<key>.",
        ),
        "filter": attr.string(
            doc = "Metadata field name; entries are kept only if its value is " +
                  "truthy (bool true, non-empty string, non-zero number, " +
                  "non-empty list/map). Empty = no filter.",
        ),
        "sort_by": attr.string(
            default = "date",
            doc = "Metadata field to sort by.",
        ),
        "sort_order": attr.string(
            default = "desc",
            doc = "Sort order: 'asc' or 'desc'.",
        ),
        "limit": attr.int(
            default = 0,
            doc = "Maximum number of entries. 0 = unlimited.",
        ),
        "fields": attr.string_list(
            doc = "Metadata fields to project into each entry. Empty = include " +
                  "all metadata. A synthesized `url` field (canonicalized " +
                  "url_path with any trailing /index.html stripped) is always " +
                  "added so templates can link without restating the strip.",
        ),
        "_builder": attr.label(
            default = "//renderers/metadata_bundle",
            executable = True,
            cfg = "exec",
        ),
    },
)
