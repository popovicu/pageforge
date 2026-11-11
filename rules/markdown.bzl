"""site_markdown rule.

Parses a Markdown file with YAML front matter, renders it through a template
chain, and produces an HTML output file plus a .meta.json sidecar containing
the extracted metadata.
"""

load(":common.bzl", "get_template_chain", "template_chain_arg")
load(":providers.bzl", "SiteFileEntry", "SiteOutputInfo", "TemplateInfo")

def _site_markdown_impl(ctx):
    out_html = ctx.actions.declare_file(ctx.label.name + ".html")
    out_meta = ctx.actions.declare_file(ctx.label.name + ".meta.json")
    out_content = ctx.actions.declare_file(ctx.label.name + ".content.html")

    template_chain = get_template_chain(ctx)

    args = ctx.actions.args()
    args.add("--src", ctx.file.src)
    args.add("--templates", template_chain_arg(template_chain))
    args.add("--url-path", ctx.attr.url_path)
    args.add("--output", out_html)
    args.add("--meta-output", out_meta)
    args.add("--content-output", out_content)

    inputs = [ctx.file.src] + template_chain
    if ctx.file.site_config:
        args.add("--site-config", ctx.file.site_config)
        inputs.append(ctx.file.site_config)

    ctx.actions.run(
        executable = ctx.executable._renderer,
        arguments = [args],
        inputs = inputs,
        outputs = [out_html, out_meta, out_content],
        mnemonic = "MarkdownRender",
        progress_message = "Rendering markdown %{label}",
    )

    entry = SiteFileEntry(
        file = out_html,
        url_path = ctx.attr.url_path,
        metadata = {},
        kind = "page",
    )

    return [
        SiteOutputInfo(entries = [entry]),
        DefaultInfo(files = depset([out_html, out_meta])),
        OutputGroupInfo(
            html = depset([out_html]),
            metadata = depset([out_meta]),
            content = depset([out_content]),
        ),
    ]

site_markdown = rule(
    implementation = _site_markdown_impl,
    doc = "Render a Markdown file with YAML front matter through a template chain.",
    attrs = {
        "src": attr.label(
            allow_single_file = [".md"],
            mandatory = True,
            doc = "Markdown source file with optional YAML front matter.",
        ),
        "template": attr.label(
            providers = [TemplateInfo],
            mandatory = True,
            doc = "Template target (site_template) to render through.",
        ),
        "url_path": attr.string(
            mandatory = True,
            doc = "URL path for this page in the final site (e.g. '/posts/my-post').",
        ),
        "site_config": attr.label(
            allow_single_file = [".yaml", ".yml"],
            doc = "Optional site-wide YAML merged into Metadata as a base. " +
                  "Front matter overlays on top — front matter wins on key " +
                  "collisions. When unset, behavior is unchanged.",
        ),
        "_renderer": attr.label(
            default = "//renderers/markdown",
            executable = True,
            cfg = "exec",
        ),
    },
)
