---
status: planned
status_description: "Reviewed specification; continuing under the owner’s end-to-end implementation instruction."
branch: feat/v0.2.0
issue: null
goal: "Readers can narrow HTML fields by requirement and choose a readable Postbode-inspired dark theme."
covers:
  - internal/engine/html.go
  - internal/engine/html.js
  - internal/engine/html.css
  - tests/format/
  - README.md
---

# HTML reading controls

## Context

The three HTML layouts share field search and API/version controls, but search only matches
field text (`internal/engine/html.js:38-44`). Required state exists in the model and visible
markers (`internal/engine/html.go:222-224`). The stylesheet defines only a light palette
(`internal/engine/html.css:1`). Postbode defines a fixed 160-degree blue gradient, translucent
surfaces and pale text (`/home/m/code/postbode/internal/webui/templates/index.html:13-25`,
at `99dfd65b2cf44411d6ef8a920fe95e874b1adf0f`). The owner requested that dark visual treatment,
a requirement filter and a reusable name for the style in shared engineering docs.

## Scope

Extend reference, explorer and compact HTML with field requirement filtering and a theme
toggle. Preserve their layout, source fidelity and light colours. Keep implementation on
the existing local feature branch under the owner's continuing v0.2.0 instructions. Shared
style guidance is edited separately in `vhco-pro/docs`; publication remains separate.
No new output formats, dependency, schema validation or pipeline execution behavior is added.

## Design

Add a labelled Field requirement selector with All fields, Required and Optional options.
Use the model's parent-relative Required boolean, even when the required column is hidden.
Combine it with text search, synchronize explorer links and groups, show the matching count
and a useful empty state. Keep query/filter state across API/version changes and reloads in
the URL (`requirement=required|optional`, absent means all). Invalid values use all. Explicit
field deep links clear incompatible filters so the target is visible; filters never remove
server-rendered content. Required markers retain their existing parent-relative semantics.

Name the reusable style VH Midnight: Postbode's 160-degree gradient from #020814 through
#061229 at 48% to #0b1f45, quiet translucent surfaces and pale blue text. Crossplane's light
palette remains unchanged. Its dark variant retains teal accents at a contrast-safe lighter
value, without adding decorative cards or changing the information hierarchy. The shared guide
records reusable tokens, contrast and responsive review rules, and a pinned Postbode source.

Provide a labelled icon toggle in the app bar, available in every HTML document, including
composition-only output. Default to the operating-system preference until the reader chooses
light or dark. Remember that explicit preference in browser local storage; apply it before
first paint, and tolerate missing or denied storage without breaking the page. Changes to
the OS preference apply only before an explicit choice. Use the same behavior when opened
as a local file. Native controls, code, focus, selected states and theme-colour metadata follow
the active theme. Without JavaScript, use the OS theme via CSS and hide inactive controls.
Print uses a light background and dark text regardless of the reader theme.

## Acceptance Criteria

- [ ] AC-01: All three layouts filter fields by All, Required or Optional together with text search, using parent-relative requirement semantics and showing correct groups, explorer links, counts and empty states.
- [ ] AC-02: Requirement and text filters survive API/version changes and reloads, invalid requirement URLs fall back to All, and explicit field links reveal their targets despite conflicting filters.
- [ ] AC-03: Every HTML document offers a keyboard-accessible theme toggle with an accurate accessible name and state, follows the OS preference before an explicit choice, remembers choices on reload, and remains usable with denied storage, offline or as a local file.
- [ ] AC-04: Dark mode uses the Postbode gradient with legible text, controls, code and focus states, while light mode retains the existing colours; all three layouts remain usable at 360px, 390px, 768px and desktop widths.
- [ ] AC-05: Without JavaScript all documentation stays readable under the OS colour preference with inactive controls hidden, and printing uses a light palette with documentation content available.
- [ ] AC-06: Existing format and source-copy behavior passes regression checks, and generated private HTTPS previews demonstrate the requirement filter and both themes.

## Test Plan

| Criterion | Proof |
| --- | --- |
| AC-01 | Chromium exercises all three layouts with required, optional, text-combined and zero-result queries; verifies parent-relative nested fields, hidden required columns, counts, groups and explorer links. |
| AC-02 | Browser navigation tests API/version transitions, reloads, invalid parameters and conflicting explicit field links. |
| AC-03 | Browser checks system light/dark and changes, explicit override and reload persistence, denied storage, keyboard toggle, local-file opening and a composition-only document. |
| AC-04 | Browser checks computed palette, contrast of text/controls/focus against surfaces, and no page overflow at four widths; inspect desktop/mobile screenshots against Postbode. |
| AC-05 | JavaScript-disabled browser checks both OS themes and no inactive controls; emulate print after filtering to verify light colour and full content. |
| AC-06 | Run make check and the existing formatter/browser suite, including exact source-copy regressions; generate each layout from the real CLI and compare advertised HTTPS bytes, then exercise filtering and toggling there. |

## Tasks

None yet; populated after specification review.

## Open Questions

None. The owner requested the filter, toggled dark variant and shared style guidance, and
previously authorized end-to-end implementation once the specification is ready.

## Review Log

| Date | Review | Kind | Commit | Model | Isolation | Outcome |
| --- | --- | --- | --- | --- | --- | --- |
| 2026-10-29 | R-1 | spec | `c21e362` | unknown | procedural | No findings; model identity not attested. |
| 2026-10-29 | R-1 | spec | `c21e362` | n/a | n/a | accepted by owner |

### R-1

NO FINDINGS

## Sources

- Postbode `99dfd65b2cf44411d6ef8a920fe95e874b1adf0f`, `internal/webui/templates/index.html:13-25`: gradient, surfaces and text colours.
- `vhco-pro/docs@c98ce2f926b7e0e1c29388be95cf7d8133a298e2`, `conventions/feature-documents.md`: specification, review and verification procedure.
- [Frontend design](https://github.com/anthropics/skills/blob/dbd4588f9e1033efb41dad4bef2f7947c8993d44/skills/frontend-design/SKILL.md): preserve the owner's visual direction and task-specific hierarchy.
- [Interface rules](https://github.com/vercel-labs/web-interface-guidelines/blob/434b7f91364665f2f733b310ec54809bf8f37937/command.md): accessible controls, URL state, dark native controls and rendered review.
