"""site_chroma_css rule.

Generates a Chroma syntax-highlighting stylesheet matching the class names
emitted by the markdown renderer's Chroma formatter. The stylesheet lands at
a given URL path in the built site, so consumers can add a <link> tag to it
from their templates.
"""

load(":providers.bzl", "SiteFileEntry", "SiteOutputInfo")

def _site_chroma_css_impl(ctx):
    out = ctx.actions.declare_file(ctx.label.name + ".css")

    args = ctx.actions.args()
    args.add("--style", ctx.attr.style)
    args.add("--output", out)
    args.add("--with-all-classes=" + ("true" if ctx.attr.with_all_classes else "false"))

    ctx.actions.run(
        executable = ctx.executable._renderer,
        arguments = [args],
        inputs = [],
        outputs = [out],
        mnemonic = "ChromaCSS",
        progress_message = "Generating chroma stylesheet %{label}",
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

site_chroma_css = rule(
    implementation = _site_chroma_css_impl,
    doc = "Generate a Chroma syntax-highlighting stylesheet and place it at url_path in the built site.",
    attrs = {
        "style": attr.string(
            mandatory = True,
            doc = "Chroma style name (e.g. 'monokai', 'dracula', 'github-dark', 'gruvbox'). Validated at build time.",
        ),
        "url_path": attr.string(
            mandatory = True,
            doc = "URL path for the stylesheet in the final site (e.g. '/static/css/chroma.css').",
        ),
        "with_all_classes": attr.bool(
            default = True,
            doc = "Emit CSS for every Chroma class even if redundant. Keeps the output stable across Chroma upgrades.",
        ),
        "_renderer": attr.label(
            default = "//renderers/chromacss",
            executable = True,
            cfg = "exec",
        ),
    },
)
