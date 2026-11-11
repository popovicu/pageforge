"""site_tags rule.

Groups page metadata by tag, renders a listing page per tag and a tag index
page. Produces a .tar archive containing the rendered pages at their final
URL paths.
"""

load(":common.bzl", "metadata_files", "resolve_template_chain", "template_chain_arg")
load(":providers.bzl", "TemplateInfo")

def _site_tags_impl(ctx):
    out_tar = ctx.actions.declare_file(ctx.label.name + ".tar")
    out_urls = ctx.actions.declare_file(ctx.label.name + ".urls.json")

    tag_chain = resolve_template_chain(ctx.attr.tag_template[TemplateInfo])
    index_chain = resolve_template_chain(ctx.attr.index_template[TemplateInfo])

    meta_files = metadata_files(ctx.attr.pages)

    extra_metadata = json.encode(ctx.attr.extra_metadata) if ctx.attr.extra_metadata else "{}"

    args = ctx.actions.args()
    args.add("--meta-files", ",".join([f.path for f in meta_files]))
    args.add("--tag-templates", template_chain_arg(tag_chain))
    args.add("--index-templates", template_chain_arg(index_chain))
    args.add("--url-prefix", ctx.attr.url_prefix)
    args.add("--output", out_tar)
    args.add("--urls-output", out_urls)
    args.add("--sort-by", ctx.attr.sort_by)
    args.add("--sort-order", ctx.attr.sort_order)
    args.add("--extra-metadata", extra_metadata)

    inputs = meta_files + tag_chain + index_chain
    if ctx.file.site_config:
        args.add("--site-config", ctx.file.site_config)
        inputs.append(ctx.file.site_config)

    ctx.actions.run(
        executable = ctx.executable._builder,
        arguments = [args],
        inputs = inputs,
        outputs = [out_tar, out_urls],
        mnemonic = "TagBuild",
        progress_message = "Building tags %{label}",
    )

    return [
        DefaultInfo(files = depset([out_tar])),
        OutputGroupInfo(urls = depset([out_urls])),
    ]

site_tags = rule(
    implementation = _site_tags_impl,
    doc = "Group pages by tag, render per-tag listing pages and a tag index into a tar archive.",
    attrs = {
        "pages": attr.label_list(
            mandatory = True,
            doc = "Page targets (site_markdown) whose metadata to aggregate. Must provide OutputGroupInfo with a 'metadata' group.",
        ),
        "tag_template": attr.label(
            providers = [TemplateInfo],
            mandatory = True,
            doc = "Template target for rendering each per-tag listing page.",
        ),
        "index_template": attr.label(
            providers = [TemplateInfo],
            mandatory = True,
            doc = "Template target for rendering the tag index page.",
        ),
        "url_prefix": attr.string(
            mandatory = True,
            doc = "URL prefix for tag pages (e.g. '/tags'). Individual tags appear at /tags/<tag-name>.",
        ),
        "sort_by": attr.string(
            default = "date",
            doc = "Metadata field to sort pages within each tag by.",
        ),
        "sort_order": attr.string(
            default = "desc",
            doc = "Sort order within each tag: 'asc' or 'desc'.",
        ),
        "extra_metadata": attr.string_dict(
            doc = "Additional metadata passed to tag and index templates.",
        ),
        "site_config": attr.label(
            allow_single_file = [".yaml", ".yml"],
            doc = "Optional site-wide YAML merged into Metadata as a base. " +
                  "extra_metadata overlays on top — extra_metadata wins on " +
                  "key collisions. When unset, behavior is unchanged.",
        ),
        "_builder": attr.label(
            default = "//renderers/tags",
            executable = True,
            cfg = "exec",
        ),
    },
)
