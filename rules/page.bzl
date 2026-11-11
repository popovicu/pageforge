"""site_page rule.

Renders a Go template chain with data from a YAML file. Use this for pages
that are not Markdown-based, such as landing pages or about pages.
"""

load(":common.bzl", "get_template_chain", "template_chain_arg")
load(":providers.bzl", "SiteFileEntry", "SiteOutputInfo", "TemplateInfo")

def _site_page_impl(ctx):
    out_html = ctx.actions.declare_file(ctx.label.name + ".html")

    template_chain = get_template_chain(ctx)

    args = ctx.actions.args()
    args.add("--data", ctx.file.data)
    args.add("--templates", template_chain_arg(template_chain))
    args.add("--url-path", ctx.attr.url_path)
    args.add("--output", out_html)

    inputs = [ctx.file.data] + template_chain
    if ctx.file.site_config:
        args.add("--site-config", ctx.file.site_config)
        inputs.append(ctx.file.site_config)
    if ctx.files.extra_data:
        args.add("--extra-data", ",".join([f.path for f in ctx.files.extra_data]))
        inputs.extend(ctx.files.extra_data)

    ctx.actions.run(
        executable = ctx.executable._renderer,
        arguments = [args],
        inputs = inputs,
        outputs = [out_html],
        mnemonic = "TemplateRender",
        progress_message = "Rendering page %{label}",
    )

    entry = SiteFileEntry(
        file = out_html,
        url_path = ctx.attr.url_path,
        metadata = {},
        kind = "page",
    )

    return [
        SiteOutputInfo(entries = [entry]),
        DefaultInfo(files = depset([out_html])),
    ]

site_page = rule(
    implementation = _site_page_impl,
    doc = "Render a template chain with YAML data to produce an HTML page.",
    attrs = {
        "data": attr.label(
            allow_single_file = [".yaml", ".yml"],
            mandatory = True,
            doc = "YAML data file. All keys become template variables.",
        ),
        "template": attr.label(
            providers = [TemplateInfo],
            mandatory = True,
            doc = "Template target (site_template) to render through.",
        ),
        "url_path": attr.string(
            mandatory = True,
            doc = "URL path for this page in the final site (e.g. '/about').",
        ),
        "site_config": attr.label(
            allow_single_file = [".yaml", ".yml"],
            doc = "Optional site-wide YAML merged into Metadata as a base. " +
                  "The data file overlays on top — data wins on key " +
                  "collisions. When unset, behavior is unchanged.",
        ),
        "extra_data": attr.label_list(
            allow_files = [".yaml", ".yml"],
            doc = "Optional additional YAML files merged into Metadata between " +
                  "site_config (base) and data (top overlay). Files are merged " +
                  "in list order, each overlaying the previous. Use this to " +
                  "feed in build-derived data such as site_metadata_bundle " +
                  "outputs. The primary data file always wins on key collisions.",
        ),
        "_renderer": attr.label(
            default = "//renderers/template",
            executable = True,
            cfg = "exec",
        ),
    },
)
