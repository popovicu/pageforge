"""site_search_index rule.

Reads .meta.json sidecars from a list of site_markdown targets and emits a
JSON search index. When include_content = True, the rule additionally reads
each page's rendered body (the 'content' OutputGroupInfo group, falling back
to 'html' for targets without one) and includes a stripped-text 'body' field.

The consumer ships their own JS to fetch the index and feed it into Fuse.js
/ MiniSearch / whatever. Pageforge does not ship JS bundling.
"""

load(":common.bzl", "metadata_files")
load(":providers.bzl", "SiteFileEntry", "SiteOutputInfo")

def _site_search_index_impl(ctx):
    out = ctx.actions.declare_file(ctx.label.name + ".json")

    meta_files = []
    html_files = []
    for dep in ctx.attr.pages:
        page_meta = metadata_files([dep])
        for f in page_meta:
            meta_files.append(f)
        if ctx.attr.include_content:
            # 'content' is the Markdown body alone; indexing it keeps the
            # template's nav, footer and scripts out of search results.
            groups = dep[OutputGroupInfo]
            page_html = (groups.content if hasattr(groups, "content") else groups.html).to_list()
            if len(page_html) != len(page_meta):
                fail("page %s does not expose matched 'content' (or 'html') and 'metadata' output groups" % dep.label)
            for f in page_html:
                html_files.append(f)

    args = ctx.actions.args()
    args.add("--meta-files", ",".join([f.path for f in meta_files]))
    if ctx.attr.fields:
        args.add("--fields", ",".join(ctx.attr.fields))
    if ctx.attr.include_content:
        args.add("--html-files", ",".join([f.path for f in html_files]))
        args.add("--include-content")
    args.add("--output", out)

    inputs = list(meta_files)
    if ctx.attr.include_content:
        inputs += html_files

    ctx.actions.run(
        executable = ctx.executable._builder,
        arguments = [args],
        inputs = inputs,
        outputs = [out],
        mnemonic = "SearchIndexBuild",
        progress_message = "Building search index %{label}",
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

site_search_index = rule(
    implementation = _site_search_index_impl,
    doc = "Generate a JSON search index from a list of site_markdown targets.",
    attrs = {
        "pages": attr.label_list(
            mandatory = True,
            doc = "site_markdown targets to index. Must provide OutputGroupInfo with 'metadata' (and 'html' when include_content = True).",
        ),
        "fields": attr.string_list(
            doc = "Metadata field names to copy from each page into the index. URL is always included.",
        ),
        "include_content": attr.bool(
            default = False,
            doc = "When True, additionally include each page's rendered body (the Markdown content without the surrounding template, stripped of tags, scripts and styles) as a 'body' field.",
        ),
        "url_path": attr.string(
            mandatory = True,
            doc = "URL path of the index in the final site (e.g. '/search.json').",
        ),
        "_builder": attr.label(
            default = "//renderers/search_index",
            executable = True,
            cfg = "exec",
        ),
    },
)
