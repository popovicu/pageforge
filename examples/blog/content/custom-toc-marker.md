---
title: "Custom TOC marker"
date: 2026-04-27
description: "Override the marker heading text via front matter (or site_config)."
tags: [pageforge, demo]
toc_marker: "Contents"
toc_summary: "Jump to a section"
author: "The Editors"
reading_minutes: 1
---

The marker heading and the disclosure summary are both configurable. Setting
`toc_marker` and `toc_summary` in front matter overrides the defaults for this
post; the same keys work in `site_config` for a site-wide rename (e.g. for a
non-English blog).

## Contents

## First section

Body to give the TOC something to point at.

## Second section

The marker heading above renders verbatim and the generated `<details>` block
sits right under it, exactly as with the default marker.
