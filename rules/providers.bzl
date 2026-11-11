"""Pageforge providers."""

SiteFileEntry = provider(
    doc = "A single file destined for the final static site.",
    fields = {
        "file": "File. The generated output file.",
        "url_path": "String. URL path, e.g. '/posts/my-article'.",
        "metadata": "Dict. Arbitrary metadata (title, date, tags, etc.).",
        "kind": "String. Built-in rules emit 'page', 'static' or 'feed'; site_sitemap skips 'static'.",
    },
)

SiteOutputInfo = provider(
    doc = "One or more files destined for the final static site.",
    fields = {
        "entries": "List of SiteFileEntry structs.",
    },
)

TemplateInfo = provider(
    doc = "A resolved template with its full inheritance chain.",
    fields = {
        "src": "File. This template's source file.",
        "parent": "TemplateInfo. The parent template, or None.",
        "chain": "depset of File. All template files in the chain, leaf first (preorder).",
    },
)
