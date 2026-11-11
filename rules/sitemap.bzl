"""site_sitemap rule.

Emits a sitemap.xml from URL paths supplied by two intake channels — same
shape as site_build:

  - `pages`: single-file producers (site_markdown, site_page, site_static).
    URL paths are read directly from SiteOutputInfo.entries at analysis time.
    When a dep also exposes an OutputGroupInfo "metadata" group of
    .meta.json sidecars (today, that's site_markdown), the renderer harvests
    the parsed `date` field and emits <lastmod> for the matching <url>.
  - `collections`: aggregators (site_collection, site_tags) whose URL set is
    only known at action time. They expose a urls.json sidecar via
    OutputGroupInfo's "urls" group; the renderer reads it.
"""

load(":providers.bzl", "SiteFileEntry", "SiteOutputInfo")

def _site_sitemap_impl(ctx):
    out = ctx.actions.declare_file(ctx.label.name + ".xml")

    page_urls = []
    meta_files = []
    for dep in ctx.attr.pages:
        for entry in dep[SiteOutputInfo].entries:
            # Skip non-page assets — site_static targets (CSS, OG card PNGs,
            # etc.) carry kind="static" and don't belong in a sitemap. Page
            # producers (site_markdown, site_page) carry kind="page".
            if entry.kind == "static":
                continue
            page_urls.append(entry.url_path)
        if hasattr(dep[OutputGroupInfo], "metadata"):
            for f in dep[OutputGroupInfo].metadata.to_list():
                meta_files.append(f)

    urls_files = []
    for dep in ctx.attr.collections:
        for f in dep[OutputGroupInfo].urls.to_list():
            urls_files.append(f)

    args = ctx.actions.args()
    args.add("--page-urls", ",".join(page_urls))
    args.add("--urls-files", ",".join([f.path for f in urls_files]))
    args.add("--meta-files", ",".join([f.path for f in meta_files]))
    args.add("--base-url", ctx.attr.base_url)
    args.add("--output", out)

    ctx.actions.run(
        executable = ctx.executable._builder,
        arguments = [args],
        inputs = urls_files + meta_files,
        outputs = [out],
        mnemonic = "SitemapBuild",
        progress_message = "Building sitemap %{label}",
    )

    entry = SiteFileEntry(
        file = out,
        url_path = ctx.attr.url_path,
        metadata = {},
        kind = "static",
    )

    return [
        SiteOutputInfo(entries = [entry]),
        DefaultInfo(files = depset([out])),
    ]

site_sitemap = rule(
    implementation = _site_sitemap_impl,
    doc = "Generate a sitemap.xml listing every URL produced by the supplied targets.",
    attrs = {
        "pages": attr.label_list(
            providers = [SiteOutputInfo],
            doc = "Single-file page producers (site_markdown, site_page). Entries with kind=\"static\" (site_static assets) are skipped — sitemaps list pages, not images/CSS. URLs come from SiteOutputInfo.entries; when a dep also exposes an OutputGroupInfo 'metadata' group, its .meta.json sidecars feed <lastmod>.",
        ),
        "collections": attr.label_list(
            doc = "Aggregator targets (site_collection, site_tags) exposing a 'urls' OutputGroupInfo with a urls.json sidecar.",
        ),
        "base_url": attr.string(
            mandatory = True,
            doc = "Site base URL (e.g. 'https://example.com'). Each emitted <loc> is base_url + url_path.",
        ),
        "url_path": attr.string(
            mandatory = True,
            doc = "URL path of the sitemap in the final site (e.g. '/sitemap.xml').",
        ),
        "_builder": attr.label(
            default = "//renderers/sitemap",
            executable = True,
            cfg = "exec",
        ),
    },
)
