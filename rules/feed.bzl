"""site_feed rule.

Reads .meta.json sidecars from a list of page targets and emits an RSS 2.0
feed XML file as a single SiteFileEntry.

The contract for entry into the feed is: the target must expose an
OutputGroupInfo("metadata") group whose files are JSON sidecars carrying the
parsed front matter (see SplitFrontMatter in renderers/core). Today
site_markdown is the only built-in rule that meets this contract; site_page
emits HTML only and is therefore not feed-eligible. Any future rule that
writes a .meta.json sidecar plugs in unchanged.

Items must carry a parseable `date` (RFC3339 or YYYY-MM-DD); undated entries
are silently skipped, since an undated entry has no place in a feed.

Channel-level lastBuildDate is the latest item date — never the wall clock —
so the action stays reproducible across builds and machines.
"""

load(":common.bzl", "metadata_files")
load(":providers.bzl", "SiteFileEntry", "SiteOutputInfo")

def _site_feed_impl(ctx):
    out = ctx.actions.declare_file(ctx.label.name + ".xml")

    meta_files = metadata_files(ctx.attr.pages)

    args = ctx.actions.args()
    args.add("--meta-files", ",".join([f.path for f in meta_files]))
    args.add("--title", ctx.attr.title)
    args.add("--description", ctx.attr.description)
    args.add("--base-url", ctx.attr.base_url)
    args.add("--url-path", ctx.attr.url_path)
    args.add("--output", out)
    args.add("--sort-by", ctx.attr.sort_by)
    args.add("--sort-order", ctx.attr.sort_order)
    if ctx.attr.limit > 0:
        args.add("--limit", str(ctx.attr.limit))
    if ctx.attr.author:
        args.add("--author", ctx.attr.author)

    ctx.actions.run(
        executable = ctx.executable._builder,
        arguments = [args],
        inputs = meta_files,
        outputs = [out],
        mnemonic = "FeedBuild",
        progress_message = "Building feed %{label}",
    )

    entry = SiteFileEntry(
        file = out,
        url_path = ctx.attr.url_path,
        metadata = {},
        kind = "feed",
    )

    return [
        SiteOutputInfo(entries = [entry]),
        DefaultInfo(files = depset([out])),
    ]

site_feed = rule(
    implementation = _site_feed_impl,
    doc = "Render an RSS 2.0 feed from a list of site_markdown targets.",
    attrs = {
        "pages": attr.label_list(
            mandatory = True,
            doc = "Targets that expose an OutputGroupInfo 'metadata' group of .meta.json sidecars. Today only site_markdown qualifies; any future rule that writes a sidecar plugs in unchanged. Items lacking a parseable 'date' field are skipped.",
        ),
        "title": attr.string(
            mandatory = True,
            doc = "Channel title.",
        ),
        "description": attr.string(
            doc = "Channel description.",
        ),
        "base_url": attr.string(
            mandatory = True,
            doc = "Site base URL (e.g. 'https://example.com'). Item links are base_url + url_path.",
        ),
        "url_path": attr.string(
            mandatory = True,
            doc = "URL path of the feed in the final site (e.g. '/rss.xml').",
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
            doc = "Maximum number of items in the feed. 0 = unlimited.",
        ),
        "author": attr.string(
            doc = "Channel-level author. Optional. An email address (e.g. 'me@example.com (Me)') goes in <managingEditor>, as RSS 2.0 requires; a plain name goes in <dc:creator>.",
        ),
        "_builder": attr.label(
            default = "//renderers/feed",
            executable = True,
            cfg = "exec",
        ),
    },
)
