# portfolio

A project portfolio. The interesting bit is a three-level template chain that
lets most pages stay flat while project pages opt into a two-column layout.

```sh
bazel run //examples/portfolio:dev    # → http://localhost:8083
```

## What it shows

- **Three-level `site_template` inheritance: `base` → `layout` → `project`.**
  - `base.html` defines a `shell` block whose default just renders `content`,
    so leaves that extend `base` directly (home, page, lists, tag indexes)
    don't need to know the layout exists.
  - `layout.html` overrides `shell` with a two-column wrapper exposing
    `primary` and `sidebar` blocks.
  - `project.html` fills in `primary` and `sidebar`.
- **Status badges from front matter.** Each project has `status: shipped|wip|archived`;
  the template emits a `.badge-{status}` class styled in `portfolio.css`.
- **Tags-as-stack.** The same `tags:` field that drives `site_tags` also
  doubles as the tech stack shown on each project card.
- **Sort by year, descending** — newest projects first.
- **Chroma highlighting** with the `github` style.

## Layout

```
content/      4 projects + home.yaml + about.yaml + BUILD
templates/    base, layout, project, project_list, home, page, tag_list, tag_index + BUILD
static/       portfolio.css (with .badge-* classes) + BUILD
```

## Tests

`tar_modes_test`.
