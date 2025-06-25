# Branch Naming & Labelling Conventions

All new work (features, bug fixes, experiments) **must** occur on a branch.  
This document defines the canonical prefixes so tooling and CI can attach the correct workflows automatically.

| Purpose | Pattern | Example |
|---------|---------|---------|
| **Bug fix** | `bug/<id>-<kebab-title>` | `bug/03-race-condition-websocket-handler` |
| Feature  | `feat/<short-title>` | `feat/prometheus-metrics` |
| Experiment / spike | `exp/<short>` | `exp/websocket-benchmarks` |
| Release prep | `release/vX.Y.Z` | `release/v1.2.0` |

### Bug Branches
* `<id>` is the two-part bug id from the markdown file (`03-bug-03`).  
  Keep only the first segment (`03`) for brevity.
* `<kebab-title>` ‑ lower-case, dash-separated summary.
* CI automatically runs the **Bug Fix** workflow for these branches.

### Labels
> GitHub labels applied automatically by the Issue / PR templates.

| Label | When | Why |
|-------|------|-----|
| `bug` | Issues & PRs that fix a bug | enables bug board filters |
| `critical` / `high` / `medium` / `low` | from bug markdown | severity-specific dashboards |
| `phase-01`, `phase-02`, … | bug discovery phase | release notes grouping |

### Commit Messages
Follow Conventional Commits with bug id reference:
```
fix(websocket): race in broadcast map (#03)
```

### PR Titles
```
fix: 03-bug-03 – race in websocket broadcast
```

Consistent naming keeps automation simple and makes the history searchable. 