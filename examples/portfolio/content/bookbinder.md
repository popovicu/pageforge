---
title: "Bookbinder, a print-on-demand workflow"
date: 2023-06-12
description: "An InDesign-replacing pipeline that turns a folder of Markdown into press-ready PDF signatures."
year: 2023
role: "Pipeline design"
client: "Vellum & Vine Press"
status: "shipped"
tags: [python, pandoc, pdf]
---

A small press wanted to stop opening InDesign every time an author submitted
revisions. We built a pipeline that takes a directory of Markdown chapters
and emits press-ready PDF signatures with bleed and crop marks.

## The pipeline

1. Pandoc renders each chapter to LaTeX with a custom template.
2. A small Python script imposes the pages into 8-up signatures.
3. Ghostscript stamps the bleed marks and produces the final PDF.

The whole thing is a Makefile. Build it once, run it every time the author
pushes a commit. The press operator only sees the final PDF.

## What surprised us

The hardest part was not the typesetting — Pandoc and LaTeX handle that
beautifully — but the imposition. Getting the page order right for a
saddle-stitched signature is a small algorithm with no margin for error: one
off-by-one and the book reads from the back to the front.

We wrote 40 lines of imposition code and 200 lines of tests for it. The
ratio felt right.
