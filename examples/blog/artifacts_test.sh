#!/usr/bin/env bash
# Asserts the blog site tar contains the rss.xml, sitemap.xml, and search.json
# artifacts produced by site_feed, site_sitemap, and site_search_index, and
# that each has the expected shape.

set -euo pipefail

TAR="${1:?usage: artifacts_test.sh <tar>}"

read_entry() {
  tar xf "$TAR" -O "$1"
}

assert_contains() {
  local label="$1" needle="$2" haystack="$3"
  if [[ "$haystack" != *"$needle"* ]]; then
    echo "FAIL: $label is missing expected content: $needle" >&2
    echo "got:" >&2
    printf '%s\n' "$haystack" >&2
    exit 1
  fi
}

rss=$(read_entry rss.xml)
[[ -n "$rss" ]] || { echo "FAIL: rss.xml is empty" >&2; exit 1; }
assert_contains "rss.xml" '<rss version="2.0"' "$rss"
assert_contains "rss.xml" '<channel>' "$rss"
assert_contains "rss.xml" '<lastBuildDate>' "$rss"
assert_contains "rss.xml" '<item>' "$rss"

sitemap=$(read_entry sitemap.xml)
[[ -n "$sitemap" ]] || { echo "FAIL: sitemap.xml is empty" >&2; exit 1; }
assert_contains "sitemap.xml" '<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">' "$sitemap"
assert_contains "sitemap.xml" '<loc>https://example.com/posts/welcome/</loc>' "$sitemap"
# Aggregator URLs (collection pagination + tag pages) must be in there too.
assert_contains "sitemap.xml" '<loc>https://example.com/posts/page/2/</loc>' "$sitemap"
assert_contains "sitemap.xml" '<loc>https://example.com/tags/' "$sitemap"

search=$(read_entry search.json)
[[ -n "$search" ]] || { echo "FAIL: search.json is empty" >&2; exit 1; }
assert_contains "search.json" '"url": "/posts/welcome/index.html"' "$search"
assert_contains "search.json" '"title": "Welcome to Forge Notes"' "$search"

# auto-toc post: the markdown renderer should preserve the literal heading
# AND inject a <details> block with anchor links to every subsequent H2/H3.
auto_toc=$(read_entry posts/auto-toc/index.html)
[[ -n "$auto_toc" ]] || { echo "FAIL: posts/auto-toc is empty" >&2; exit 1; }
assert_contains "posts/auto-toc" '<h2 id="table-of-contents">Table of contents</h2>' "$auto_toc"
assert_contains "posts/auto-toc" '<details class="pageforge-toc"><summary>Open Table of contents</summary>' "$auto_toc"
assert_contains "posts/auto-toc" '<a href="#first-section">First section</a>' "$auto_toc"
assert_contains "posts/auto-toc" '<a href="#a-nested-point">A nested point</a>' "$auto_toc"
assert_contains "posts/auto-toc" '<a href="#second-section">Second section</a>' "$auto_toc"
assert_contains "posts/auto-toc" '<a href="#third-section">Third section</a>' "$auto_toc"

# no-toc post: front-matter `auto_toc: false` keeps the literal heading and
# suppresses the <details> block.
no_toc=$(read_entry posts/no-toc/index.html)
[[ -n "$no_toc" ]] || { echo "FAIL: posts/no-toc is empty" >&2; exit 1; }
assert_contains "posts/no-toc" '<h2 id="table-of-contents">Table of contents</h2>' "$no_toc"
if [[ "$no_toc" == *'<details class="pageforge-toc">'* ]]; then
  echo "FAIL: posts/no-toc opted out via auto_toc:false but a TOC was still injected" >&2
  exit 1
fi

# welcome post: no `## Table of contents` heading at all → no TOC injected,
# no spurious markup.
welcome=$(read_entry posts/welcome/index.html)
if [[ "$welcome" == *'<details class="pageforge-toc">'* ]]; then
  echo "FAIL: posts/welcome has no TOC marker but a TOC block was injected" >&2
  exit 1
fi

# custom-toc-marker post: front-matter `toc_marker: "Contents"` retargets the
# detection, and `toc_summary: "Jump to a section"` overrides the disclosure
# label.
custom_toc=$(read_entry posts/custom-toc-marker/index.html)
[[ -n "$custom_toc" ]] || { echo "FAIL: posts/custom-toc-marker is empty" >&2; exit 1; }
assert_contains "posts/custom-toc-marker" '<h2 id="contents">Contents</h2>' "$custom_toc"
assert_contains "posts/custom-toc-marker" '<details class="pageforge-toc"><summary>Jump to a section</summary>' "$custom_toc"
assert_contains "posts/custom-toc-marker" '<a href="#first-section">First section</a>' "$custom_toc"
# Default marker text must NOT match when toc_marker is overridden — a stray
# default-keyed injection would mean the override didn't take effect.
if [[ "$custom_toc" == *'<summary>Open Table of contents</summary>'* ]]; then
  echo "FAIL: posts/custom-toc-marker still uses the default summary text" >&2
  exit 1
fi
