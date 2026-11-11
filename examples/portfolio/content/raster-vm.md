---
title: "A teaching VM for raster graphics"
date: 2025-09-04
description: "A 32-instruction stack machine that draws to a framebuffer. Built for a one-week summer workshop."
year: 2025
role: "Curriculum & implementation"
client: "Coast Polytechnic"
status: "shipped"
tags: [c, sdl, education]
---

A summer workshop on graphics needed something the students could read end to
end in a week. So we built a tiny VM with a 32-opcode instruction set, a
framebuffer for output, and an assembler simple enough to fit on a postcard.

## The shape of it

The VM is roughly 600 lines of C. Programs are bytecode files of two- and
three-byte instructions: arithmetic, stack ops, conditional jumps, and four
draw primitives that write into a 320×200 framebuffer. SDL2 takes that
framebuffer to a window.

The assembler is another 200 lines. It does not do macros or labels-as-data —
just symbolic labels and a handful of pseudo-ops. The grammar is "one
instruction per line, comments start with `;`".

## Why it worked

Students could trace any program by hand on paper. That sounds like a low
bar, but it is the bar — once a teaching VM grows past it, the exercise of
single-stepping through the bytecode stops being instructive and starts
being a chore.

The opcode count was the design constraint that kept everything else honest.
