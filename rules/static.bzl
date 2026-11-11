"""site_static rule.

Pass-through rule for static files (CSS, images, fonts, etc.). Each file in
srcs becomes a SiteFileEntry with its URL path set to url_prefix + filename.
"""

load(":providers.bzl", "SiteFileEntry", "SiteOutputInfo")

def _site_static_impl(ctx):
    entries = []
    for src in ctx.files.srcs:
        url = ctx.attr.url_prefix.rstrip("/") + "/" + src.basename
        entries.append(SiteFileEntry(
            file = src,
            url_path = url,
            metadata = {},
            kind = "static",
        ))
    return [
        SiteOutputInfo(entries = entries),
        DefaultInfo(files = depset(ctx.files.srcs)),
    ]

site_static = rule(
    implementation = _site_static_impl,
    doc = "Include static files in the site at a given URL prefix.",
    attrs = {
        "srcs": attr.label_list(
            allow_files = True,
            mandatory = True,
            doc = "Static files to include (CSS, images, fonts, etc.).",
        ),
        "url_prefix": attr.string(
            mandatory = True,
            doc = "URL prefix for all files (e.g. '/css'). Each file appears at prefix/filename.",
        ),
    },
)
