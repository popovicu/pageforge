"""site_collection rule.

Aggregates page metadata, sorts and paginates it, and renders listing pages
through a template chain. Produces a .tar archive containing the rendered
pages at their final URL paths.
"""

load(":common.bzl", "get_template_chain", "metadata_files", "template_chain_arg")
load(":providers.bzl", "TemplateInfo")

def _site_collection_impl(ctx):
    out_tar = ctx.actions.declare_file(ctx.label.name + ".tar")
    out_urls = ctx.actions.declare_file(ctx.label.name + ".urls.json")

    template_chain = get_template_chain(ctx)

    meta_files = metadata_files(ctx.attr.pages)

    extra_metadata = json.encode(ctx.attr.extra_metadata) if ctx.attr.extra_metadata else "{}"

    args = ctx.actions.args()
    args.add("--meta-files", ",".join([f.path for f in meta_files]))
    args.add("--templates", template_chain_arg(template_chain))
    args.add("--url-path", ctx.attr.url_path)
    args.add("--output", out_tar)
    args.add("--urls-output", out_urls)
    args.add("--sort-by", ctx.attr.sort_by)
    args.add("--sort-order", ctx.attr.sort_order)
    if ctx.attr.paginate > 0:
        args.add("--paginate", str(ctx.attr.paginate))
    args.add("--extra-metadata", extra_metadata)

    inputs = meta_files + template_chain
    if ctx.file.site_config:
        args.add("--site-config", ctx.file.site_config)
        inputs.append(ctx.file.site_config)

    ctx.actions.run(
        executable = ctx.executable._builder,
        arguments = [args],
        inputs = inputs,
        outputs = [out_tar, out_urls],
        mnemonic = "CollectionBuild",
        progress_message = "Building collection %{label}",
    )

    return [
        DefaultInfo(files = depset([out_tar])),
        OutputGroupInfo(urls = depset([out_urls])),
    ]

site_collection = rule(
    implementation = _site_collection_impl,
    doc = "Aggregate, sort, paginate, and render listing pages into a tar archive.",
    attrs = {
        "pages": attr.label_list(
            mandatory = True,
            doc = "Page targets (site_markdown) whose metadata to aggregate. Must provide OutputGroupInfo with a 'metadata' group.",
        ),
        "template": attr.label(
            providers = [TemplateInfo],
            mandatory = True,
            doc = "Template target (site_template) for rendering each listing page.",
        ),
        "url_path": attr.string(
            mandatory = True,
            doc = "Base URL path for the collection (e.g. '/posts'). Paginated pages appear at /posts/page/2, etc.",
        ),
        "paginate": attr.int(
            default = 0,
            doc = "Number of items per page. 0 means all items on one page.",
        ),
        "sort_by": attr.string(
            default = "date",
            doc = "Metadata field to sort by.",
        ),
        "sort_order": attr.string(
            default = "desc",
            doc = "Sort order: 'asc' or 'desc'.",
        ),
        "extra_metadata": attr.string_dict(
            doc = "Additional metadata passed to the collection template.",
        ),
        "site_config": attr.label(
            allow_single_file = [".yaml", ".yml"],
            doc = "Optional site-wide YAML merged into Metadata as a base. " +
                  "extra_metadata overlays on top — extra_metadata wins on " +
                  "key collisions. When unset, behavior is unchanged.",
        ),
        "_builder": attr.label(
            default = "//renderers/collection",
            executable = True,
            cfg = "exec",
        ),
    },
)
