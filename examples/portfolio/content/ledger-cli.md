---
title: "Ledger CLI for a small co-op"
date: 2026-01-15
description: "A double-entry bookkeeping tool the bookkeeper can actually use from the kitchen table."
year: 2024
role: "Tools & training"
client: "Riverbend Worker Co-op"
status: "shipped"
live_url: "https://example.invalid/ledger"
tags: [go, sqlite, cli]
---

The bookkeeper at Riverbend was keeping the books in a spreadsheet, which
worked until it didn't. We built a small command-line tool around a
single-file SQLite database that does double-entry bookkeeping in the open and
prints a balance sheet on demand.

## Constraints

- One person was going to maintain it. Whatever shipped had to read like
  English at the prompt.
- The data file had to be checkable into a private git repo, no binary blobs.
- The auditor wanted CSV export, on demand, no setup.

## What we built

A single Go binary (`ledger`) backed by a single SQLite file. Subcommands map
to verbs the bookkeeper already used: `ledger post`, `ledger reconcile`,
`ledger close`. Reports are plain text by default, CSV when piped.

The whole thing fits in about 1,200 lines of Go. Most of those lines are the
column-formatting logic for the balance-sheet output — the accounting rules
themselves are surprisingly small once you let SQLite do the heavy lifting.
