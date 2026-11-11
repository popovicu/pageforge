"""site_template rule.

Defines a template with optional inheritance. Templates can extend a parent
template, forming an inheritance chain resolved by Go's html/template package
via {{block}} / {{define}} directives.
"""

load(":providers.bzl", "TemplateInfo")

def _site_template_impl(ctx):
    parent = None
    transitive = []
    if ctx.attr.extends:
        parent = ctx.attr.extends[TemplateInfo]
        transitive = [parent.chain]

    return [TemplateInfo(
        src = ctx.file.src,
        parent = parent,
        chain = depset(
            direct = [ctx.file.src],
            transitive = transitive,
            order = "preorder",
        ),
    )]

site_template = rule(
    implementation = _site_template_impl,
    doc = "Define a template with optional inheritance via 'extends'.",
    attrs = {
        "src": attr.label(
            allow_single_file = True,
            mandatory = True,
            doc = "Go html/template source file.",
        ),
        "extends": attr.label(
            providers = [TemplateInfo],
            doc = "Parent template target. This template's {{define}} blocks fill the parent's {{block}} slots.",
        ),
    },
)
