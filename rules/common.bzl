"""Common utilities for Pageforge rules."""

load(":providers.bzl", "TemplateInfo")

def resolve_template_chain(template_info):
    """Resolves a TemplateInfo provider into an ordered list of template files.

    Returns the chain in base-first order, suitable for passing to Go's
    template.ParseFiles which executes the first file parsed.

    Args:
        template_info: A TemplateInfo provider.

    Returns:
        List of File objects, base template first, leaf template last.
    """
    chain = list(reversed(template_info.chain.to_list()))

    # Go names each parsed file after its basename, so a second file with the
    # same name silently replaces the first instead of extending it.
    seen = {}
    for f in chain:
        if f.basename in seen:
            fail("template chain has two files named %s (%s and %s); rename one, since Go templates are keyed by file name" % (f.basename, seen[f.basename].short_path, f.short_path))
        seen[f.basename] = f
    return chain

def get_template_chain(ctx):
    """Convenience wrapper: resolves ctx.attr.template's TemplateInfo.

    Args:
        ctx: Rule context. Must have a 'template' attr with TemplateInfo.

    Returns:
        List of File objects, base template first, leaf template last.
    """
    return resolve_template_chain(ctx.attr.template[TemplateInfo])

def metadata_files(deps):
    """Collects .meta.json sidecars from page targets.

    Args:
        deps: Targets exposing an OutputGroupInfo 'metadata' group.

    Returns:
        List of sidecar File objects, in dep order.
    """
    files = []
    for dep in deps:
        groups = dep[OutputGroupInfo]
        if not hasattr(groups, "metadata"):
            fail("%s has no 'metadata' output group of .meta.json sidecars; pass site_markdown targets" % dep.label)
        files.extend(groups.metadata.to_list())
    return files

def template_chain_arg(template_files):
    """Formats a template chain as a comma-separated string of paths.

    Args:
        template_files: List of File objects (base first).

    Returns:
        Comma-separated string of file paths.
    """
    return ",".join([f.path for f in template_files])
