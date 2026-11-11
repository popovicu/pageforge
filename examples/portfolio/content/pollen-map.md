---
title: "A regional pollen map"
date: 2026-03-20
description: "An interactive map of pollen counts across the watershed, updated hourly from a network of small sensors."
year: 2026
role: "Backend & data pipeline"
client: "Watershed Allergy Network"
status: "wip"
tags: [go, sqlite, leaflet]
---

A volunteer network of about 40 pollen sensors sends counts to a central
collector hourly. The site shows the current readings on a leaflet map and
keeps the last 90 days of history queryable.

## In progress

The data pipeline is solid: sensors POST counts, a Go service writes them to
SQLite, a small reader process aggregates and caches the map tiles. The
front-end map is the part that needs more design work — the current layer
toggle is intuitive only to people who already know the pollens by name.

## What we are testing

We are still calibrating the sensor network. About a third of the sensors
need a daily re-zero against a known sample, and we are working out whether
that should happen via a small mechanical actuator (expensive, maintenance-
free) or via a software baseline correction (cheap, fragile). The honest
answer is probably both, in different parts of the network.
