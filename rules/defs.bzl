# Public API for Pageforge rules.
# Usage: load("@pageforge//rules:defs.bzl", "site_template", ...)

load(":build.bzl", _site_build = "site_build", _site_zip = "site_zip")
load(":chromacss.bzl", _site_chroma_css = "site_chroma_css")
load(":collection.bzl", _site_collection = "site_collection")
load(":feed.bzl", _site_feed = "site_feed")
load(":markdown.bzl", _site_markdown = "site_markdown")
load(":metadata_bundle.bzl", _site_metadata_bundle = "site_metadata_bundle")
load(":page.bzl", _site_page = "site_page")
load(":providers.bzl", _SiteFileEntry = "SiteFileEntry", _SiteOutputInfo = "SiteOutputInfo", _TemplateInfo = "TemplateInfo")
load(":search_index.bzl", _site_search_index = "site_search_index")
load(":serve.bzl", _site_serve = "site_serve")
load(":sitemap.bzl", _site_sitemap = "site_sitemap")
load(":static.bzl", _site_static = "site_static")
load(":tags.bzl", _site_tags = "site_tags")
load(":templates.bzl", _site_template = "site_template")

SiteFileEntry = _SiteFileEntry
SiteOutputInfo = _SiteOutputInfo
TemplateInfo = _TemplateInfo
site_build = _site_build
site_chroma_css = _site_chroma_css
site_collection = _site_collection
site_feed = _site_feed
site_markdown = _site_markdown
site_metadata_bundle = _site_metadata_bundle
site_page = _site_page
site_search_index = _site_search_index
site_serve = _site_serve
site_sitemap = _site_sitemap
site_static = _site_static
site_tags = _site_tags
site_template = _site_template
site_zip = _site_zip
