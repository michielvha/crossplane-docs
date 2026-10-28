---
status: in-progress
status_description: "Implementing and testing the reviewed reading controls on the existing feature branch."
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

- [x] AC-01: All three layouts filter fields by All, Required or Optional together with text search, using parent-relative requirement semantics and showing correct groups, explorer links, counts and empty states.
      Evidence: Chromium verified all three layouts using a config that omits the required column: six required fields, including array-item and schema-alternative properties; optional/text combinations, counts and empty states passed.
- [ ] AC-02: Requirement and text filters survive API/version changes and reloads, invalid requirement URLs fall back to All, and explicit field links reveal their targets despite conflicting filters.
      Evidence: R-2 returned not met: a conflicting requirement filter replaced an explicit explorer field query.
- [x] AC-03: Every HTML document offers a keyboard-accessible theme toggle with an accurate accessible name and state, follows the OS preference before an explicit choice, remembers choices on reload, and remains usable with denied storage, offline or as a local file.
      Evidence: Chromium verified OS changes, keyboard activation, explicit light/dark persistence, denied storage without page errors, offline operation and a composition-only local file.
- [x] AC-04: Dark mode uses the Postbode gradient with legible text, controls, code and focus states, while light mode retains the existing colours; all three layouts remain usable at 360px, 390px, 768px and desktop widths.
      Evidence: Computed Postbode gradient, native control colours, text/control/focus contrast and four-width geometry checks passed in all layouts; inspected actual private desktop/mobile dark screenshots.
- [x] AC-05: Without JavaScript all documentation stays readable under the OS colour preference with inactive controls hidden, and printing uses a light palette with documentation content available.
      Evidence: JavaScript-disabled checks passed for both OS themes and all three layouts. Print emulation after zero-match filtering showed all API panels/fields, light text colours and no empty-filter message.
- [x] AC-06: Existing format and source-copy behavior passes regression checks, and generated private HTTPS previews demonstrate the requirement filter and both themes.
      Evidence: make check passed and 16 browser cases passed with 302 assertions. Each private HTTPS file matched real CLI bytes; actual filtering, toggling, reload and exact 64-line copying passed.

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

- [x] T1 `tests/format/format.test.ts` - write filter, theme and fallback checks and observe failures for missing behavior (AC-01 through AC-06).
      Evidence: New Chromium checks failed for the absent requirement selector, missing theme toggle and unchanged dark-system colours; eight existing checks passed (/tmp/crossplane-reading-red.log).
- [x] T2 `internal/engine/html.go`, `internal/engine/html.js`, `internal/engine/html.css` - implement controls and colours so the new and existing browser checks pass (AC-01 through AC-06).
      Evidence: All 16 browser checks passed with 302 assertions, including filters, URL/deep links, native dark controls, persistence, denied storage, file opening, mobile geometry and print; make check passed.
- [x] T3 `README.md` - describe controls and observe make check plus generated private HTTPS previews in both themes (AC-06).
      Evidence: make check passed; real CLI generation matched all three private HTTPS files byte-for-byte. Both themes and requirement/search/reload interactions passed at desktop/mobile widths, and all three 64-line sources copied exactly. README describes the controls.
- [ ] T5 `internal/engine/html.js`, `internal/engine/html.go`, `tests/format/format.test.ts` - reproduce and fix conflicting query-field links and no-JavaScript browser metadata, then pass regressions (AC-02, AC-03, AC-05).
- [ ] T4 `docs/specs/html-reading-controls.spec.md` - store an independent implementation review and pass the complete gate (AC-01 through AC-06).

## Open Questions

None. The owner requested the filter, toggled dark variant and shared style guidance, and
previously authorized end-to-end implementation once the specification is ready.

## Review Log

| Date | Review | Kind | Commit | Model | Isolation | Outcome |
| --- | --- | --- | --- | --- | --- | --- |
| 2026-10-29 | R-1 | spec | `c21e362` | unknown | procedural | No findings; model identity not attested. |
| 2026-10-29 | R-1 | spec | `c21e362` | n/a | n/a | accepted by owner |
| 2026-10-29 | R-2 | implementation | `41ddfb0` | unknown | procedural | Query field links and no-JavaScript browser colour need fixes; identity not attested. |

### R-1

NO FINDINGS

### R-2

CRITERION: AC-01 All three layouts filter fields by All, Required or Optional together with text search, using parent-relative requirement semantics and showing correct groups, explorer links, counts and empty states.
VERDICT: met
EVIDENCE: At 41ddfb08cf921e3983cf0701d537df47d97c09b9, fresh race tests passed. Inspected the passing 16-case browser run with hidden-required-column and nested/alternative field assertions. Independently selected Required in all three private previews and observed six matching fields; search, groups, explorer links and empty-state handling use the model-derived data-required attribute.

CRITERION: AC-02 Requirement and text filters survive API/version changes and reloads, invalid requirement URLs fall back to All, and explicit field links reveal their targets despite conflicting filters.
VERDICT: not met
EVIDENCE: At 41ddfb08cf921e3983cf0701d537df47d97c09b9, the actual private explorer URL with api=api-1&field=spec.region&requirement=optional selected spec.networks[] and rewrote the requested field in the URL. Hash links clear incompatible filters, but the generated explorer field query parameter is resolved only after filtering and silently replaced.

CRITERION: AC-03 Every HTML document offers a keyboard-accessible theme toggle with an accurate accessible name and state, follows the OS preference before an explicit choice, remembers choices on reload, and remains usable with denied storage, offline or as a local file.
VERDICT: met
EVIDENCE: At 41ddfb08cf921e3983cf0701d537df47d97c09b9, inspected system-preference, keyboard activation, explicit-choice reload, denied-storage, offline and composition-only local-file browser checks. Theme initialization runs before stylesheet/body, catches denied storage and keeps OS changes separate from explicit preference. Independently exercised the accessible theme toggle in all three private previews.

CRITERION: AC-04 Dark mode uses the Postbode gradient with legible text, controls, code and focus states, while light mode retains the existing colours; all three layouts remain usable at 360px, 390px, 768px and desktop widths.
VERDICT: met
EVIDENCE: At 41ddfb08cf921e3983cf0701d537df47d97c09b9, independently observed the exact 160-degree Postbode gradient and pale text in all three private previews, then restored the original light ink. Inspected passing composited text/control/focus contrast and geometry checks at 360px, 390px, 768px and desktop widths; palette values match the pinned Postbode source.

CRITERION: AC-05 Without JavaScript all documentation stays readable under the OS colour preference with inactive controls hidden, and printing uses a light palette with documentation content available.
VERDICT: met
EVIDENCE: At 41ddfb08cf921e3983cf0701d537df47d97c09b9, inspected passing JavaScript-disabled light/dark and print-after-filter checks for every layout. Independently opened the reference with JavaScript disabled and OS dark preference: documentation used the dark gradient and pale text, and the inactive toggle stayed hidden. Print rules and detail expansion expose documentation in a light palette.

CRITERION: AC-06 Existing format and source-copy behavior passes regression checks, and generated private HTTPS previews demonstrate the requirement filter and both themes.
VERDICT: met
EVIDENCE: At 41ddfb08cf921e3983cf0701d537df47d97c09b9, fresh package race tests passed; inspected successful make-check and the 16-case, 302-assertion formatter/browser log, including exact source-copy regressions. Independently reached all three generated private HTTPS previews and exercised requirement filtering plus dark/light toggling.

FINDING: F-01
KIND: factual
SEVERITY: blocking
WHERE: internal/engine/html.js initial selectedPath and filterFields
EVIDENCE: Opening explorer.html?api=api-1&field=spec.region&requirement=optional on the actual preview selected spec.networks[] and rewrote field=spec.networks[]. Explorer-generated deep links use the field query parameter, but only hash-based field links clear conflicting filters.
PROPOSAL: Resolve an explicit initial field query before fallback selection and clear filters that would hide that target. Add query-form deep-link regressions with conflicting requirement and text filters, alongside the existing hash-link tests.

FINDING: F-02
KIND: factual
SEVERITY: should-fix
WHERE: internal/engine/html.go theme-colour metadata
EVIDENCE: With JavaScript disabled and OS dark preference, the actual reference preview has dark ink and the Postbode gradient but retains the single theme-color meta value #f5f8f8. The Design requires theme-colour metadata to follow the active theme.
PROPOSAL: Provide light/dark media-qualified theme-colour metadata for the no-JavaScript case and keep it synchronized with explicit JavaScript choices. Test the effective metadata under both OS themes with JavaScript disabled.

## Sources

- Postbode `99dfd65b2cf44411d6ef8a920fe95e874b1adf0f`, `internal/webui/templates/index.html:13-25`: gradient, surfaces and text colours.
- `vhco-pro/docs@c98ce2f926b7e0e1c29388be95cf7d8133a298e2`, `conventions/feature-documents.md`: specification, review and verification procedure.
- [Frontend design](https://github.com/anthropics/skills/blob/dbd4588f9e1033efb41dad4bef2f7947c8993d44/skills/frontend-design/SKILL.md): preserve the owner's visual direction and task-specific hierarchy.
- [Interface rules](https://github.com/vercel-labs/web-interface-guidelines/blob/434b7f91364665f2f733b310ec54809bf8f37937/command.md): accessible controls, URL state, dark native controls and rendered review.
