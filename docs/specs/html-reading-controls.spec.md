---
status: complete
status_description: "All reading controls passed regression checks, independent re-review and the complete gate."
branch: feat/v0.2.0
issue: 16
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
- [x] AC-02: Requirement and text filters survive API/version changes and reloads, invalid requirement URLs fall back to All, and explicit field links reveal their targets despite conflicting filters.
      Evidence: The R-2 query-link regression failed before correction. Conflicting requirement, text and combined query-form links now reveal the requested field; all 16 browser checks with 320 assertions pass.
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
- [x] T5 `internal/engine/html.js`, `internal/engine/html.go`, `tests/format/format.test.ts` - reproduce and fix conflicting query-field links and no-JavaScript browser metadata, then pass regressions (AC-02, AC-03, AC-05).
      Evidence: Both review regressions failed before correction (/tmp/crossplane-reading-review-red.log). After resolving explicit query targets before filtering and supplying media-qualified metadata, all 16 browser cases passed with 320 assertions; make check passed.
- [x] T4 `docs/specs/html-reading-controls.spec.md` - store an independent implementation review and pass the complete gate (AC-01 through AC-06).

      Evidence: R-3 independently reran 16 browser cases/320 assertions, package race tests and both HTTPS repros. All six criteria met with no findings; exact records verified with review check.

## Open Questions

None. The owner requested the filter, toggled dark variant and shared style guidance, and
previously authorized end-to-end implementation once the specification is ready.

## Review Log

| Date | Review | Kind | Commit | Model | Isolation | Outcome |
| --- | --- | --- | --- | --- | --- | --- |
| 2026-10-29 | R-1 | spec | `c21e362` | unknown | procedural | No findings; model identity not attested. |
| 2026-10-29 | R-1 | spec | `c21e362` | n/a | n/a | accepted by owner |
| 2026-10-29 | R-2 | implementation | `41ddfb0` | unknown | procedural | Query field links and no-JavaScript browser colour need fixes; identity not attested. |
| 2026-10-29 | R-3 | implementation | `f7d00a8` | unknown | procedural | All six criteria met; no findings. Model identity not attested. |
| 2026-10-29 | R-4 | implementation | `95124b3` | unknown | procedural | AC-02 and AC-06 not met; observe saved URL state before reload and rerun regression checks. |
| 2026-10-29 | R-5 | implementation | `4aa59e2` | unknown | procedural | All six criteria met; synchronized reload and full browser regressions passed, no findings. |

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

### R-3

CRITERION: AC-01 All three layouts filter fields by All, Required or Optional together with text search, using parent-relative requirement semantics and showing correct groups, explorer links, counts and empty states.
VERDICT: met
EVIDENCE: At f7d00a822f829fea6727e07ff524b184ea72ba01, independently ran the final browser suite: 16 cases and 320 assertions passed, including all layouts, parent-relative nested/alternative requirements, hidden required columns, combined search, counts, groups, explorer links and empty states. Fresh package race tests also passed.

CRITERION: AC-02 Requirement and text filters survive API/version changes and reloads, invalid requirement URLs fall back to All, and explicit field links reveal their targets despite conflicting filters.
VERDICT: met
EVIDENCE: At f7d00a822f829fea6727e07ff524b184ea72ba01, final browser regressions passed for API/version transitions, reload persistence, invalid URLs and hash/query deep links. Independently opened the original conflicting explorer query on HTTPS: spec.region remained selected, the requirement became All and conflicting text cleared. Compatible saved filters still pass reload tests.

CRITERION: AC-03 Every HTML document offers a keyboard-accessible theme toggle with an accurate accessible name and state, follows the OS preference before an explicit choice, remembers choices on reload, and remains usable with denied storage, offline or as a local file.
VERDICT: met
EVIDENCE: At f7d00a822f829fea6727e07ff524b184ea72ba01, independently executed passing system-change, keyboard-toggle, explicit-choice persistence, denied-storage, offline and composition-only local-file browser tests. Inspected synchronous pre-style initialization and OS-change handling; accessible toggle state and metadata follow explicit choices.

CRITERION: AC-04 Dark mode uses the Postbode gradient with legible text, controls, code and focus states, while light mode retains the existing colours; all three layouts remain usable at 360px, 390px, 768px and desktop widths.
VERDICT: met
EVIDENCE: At f7d00a822f829fea6727e07ff524b184ea72ba01, final browser checks passed for the exact Postbode gradient, composited text/control/focus contrast, original light colours and geometry at 360px, 390px, 768px and desktop widths. Independently observed the HTTPS compact preview switch to the pinned gradient and restore the original light ink.

CRITERION: AC-05 Without JavaScript all documentation stays readable under the OS colour preference with inactive controls hidden, and printing uses a light palette with documentation content available.
VERDICT: met
EVIDENCE: At f7d00a822f829fea6727e07ff524b184ea72ba01, final browser tests passed for JavaScript-disabled light/dark output and print after filtering in every layout. Independently checked both OS themes on the HTTPS reference with JavaScript disabled: readable palette, hidden toggle and exactly one correctly matching media-qualified theme-colour metadata value were observed.

CRITERION: AC-06 Existing format and source-copy behavior passes regression checks, and generated private HTTPS previews demonstrate the requirement filter and both themes.
VERDICT: met
EVIDENCE: At f7d00a822f829fea6727e07ff524b184ea72ba01, fresh package race tests and the complete 16-case formatter/browser suite passed, including exact multiline source copying and AsciiDoc/template compatibility. Inspected successful final make-check output. Independently reached the refreshed reference, explorer and compact private HTTPS outputs and verified the reported regressions, filtering and theme behavior. The feature tree is clean and its spec review is fresh.

### R-4

CRITERION: AC-01 All three layouts filter fields by All, Required or Optional together with text search, using parent-relative requirement semantics and showing correct groups, explorer links, counts and empty states.
VERDICT: met
EVIDENCE: At 95124b34ce929b84464f4de9548c922f5b08aef2, independently executed requirement/search browser cases for all three layouts with the required column omitted. Six parent-relative required fields, nested/alternative properties, optional/text combinations, groups, explorer links, counts and empty states pass. Actual private HTTPS interactions independently confirm six required matches and the combined optional/empty results.

CRITERION: AC-02 Requirement and text filters survive API/version changes and reloads, invalid requirement URLs fall back to All, and explicit field links reveal their targets despite conflicting filters.
VERDICT: not met
EVIDENCE: At 95124b3, the independent final formatter run fails format.test.ts:177: search is empty after reload rather than region. API/version, invalid requirement and explicit field-link cases pass, as do actual private HTTPS filter persistence and deep-link reveal. Forty isolated executions of the same failing flow do not reproduce the loss. The recorded intermittent failure remains unresolved; test input/URL synchronization needs clarification before this composite criterion can be treated as verified passing.

CRITERION: AC-03 Every HTML document offers a keyboard-accessible theme toggle with an accurate accessible name and state, follows the OS preference before an explicit choice, remembers choices on reload, and remains usable with denied storage, offline or as a local file.
VERDICT: met
EVIDENCE: At 95124b3, independently executed system-theme changes, keyboard activation, explicit-choice reload persistence, denied storage, offline output and composition-only local-file cases successfully. Inspected synchronous pre-style initialization and accessible name/state handling. On all three real private HTTPS layouts, keyboard dark activation, persisted reload and restoration of light mode pass with no page errors.

CRITERION: AC-04 Dark mode uses the Postbode gradient with legible text, controls, code and focus states, while light mode retains the existing colours; all three layouts remain usable at 360px, 390px, 768px and desktop widths.
VERDICT: met
EVIDENCE: At 95124b3, independently executed computed gradient, native-control, composited text/control/focus contrast and geometry cases for all three layouts at360,390,768 and1280px. The exact Postbode160-degree gradient and original light ink pass. Actual private HTTPS geometry and theme interactions also pass; inspected newly captured dark-mobile screenshots for all layouts and the desktop reference. Shared VH Midnight guidance records the same reusable palette and behavior.

CRITERION: AC-05 Without JavaScript all documentation stays readable under the OS colour preference with inactive controls hidden, and printing uses a light palette with documentation content available.
VERDICT: met
EVIDENCE: At 95124b3, independently executed all JavaScript-disabled OS-light/OS-dark cases and print-after-zero-match cases successfully. All server-rendered APIs remain readable, inactive controls are hidden, effective media-qualified theme-color metadata matches the OS, and print exposes all fields with the light palette. Inspected the supporting CSS and print detail expansion.

CRITERION: AC-06 Existing format and source-copy behavior passes regression checks, and generated private HTTPS previews demonstrate the requirement filter and both themes.
VERDICT: not met
EVIDENCE: At 95124b3, private HTTPS reference/explorer/compact files match current clean-build CLI output byte-for-byte; actual filters, reload/deep links, themes, four widths and exact source copying pass. Inspected the clean readiness make-check and18-case/336-assertion log. Independent reruns exposed an unresolved search/reload assertion failure in the final run:17pass,1fail,335assertions. Earlier runs passed all18 cases but hit browser timing/teardown failures during quota pressure. The fresh regression execution therefore is not clean despite successful source-copy and preview checks.

FINDING: F-01
KIND: factual
SEVERITY: should-fix
WHERE: tests/format/format.test.ts:175-177, search state before reload
EVIDENCE: /tmp/crossplane-reading-controls-release-independent-formats-final.log records Expected region, Received empty string at line177, while17 other cases pass. The exact flow passes40 isolated browser repetitions, and actual private HTTPS persistence passes, so a deterministic renderer defect is not established. The test reloads immediately after fill without observing the URL state that is expected to survive.
PROPOSAL: Observe the saved q=region URL state before reloading, then retain the reload assertion and rerun the full formatter suite. If that synchronized check fails, investigate it as a product persistence defect rather than dismissing the failure as host timing.

### R-5

CRITERION: AC-01 All three layouts filter fields by All, Required or Optional together with text search, using parent-relative requirement semantics and showing correct groups, explorer links, counts and empty states.
VERDICT: met
EVIDENCE: At 4aa59e22b11a7c09f640a55c369a49483ae484b4, independently executed the full formatter suite:18 cases and336 assertions pass. All three hidden-required-column filter cases establish six parent-relative required fields, nested/alternative properties, combined search, optional results, groups, explorer links, counts and empty states. Prior actual private HTTPS filter interactions remain applicable because the HTML generator, script and stylesheet are unchanged.

CRITERION: AC-02 Requirement and text filters survive API/version changes and reloads, invalid requirement URLs fall back to All, and explicit field links reveal their targets despite conflicting filters.
VERDICT: met
EVIDENCE: At 4aa59e2, independently observed the strengthened persistence test pass: it now waits for q=region in the URL before reloading, while retaining the post-reload input and field assertions. The full18-case/336-assertion run is clean, including API/version transitions, invalid requirement URLs and conflicting hash/query field links. The prior intermittent failure remains recorded in R-4; this new execution clears its unresolved synchronization finding without asserting a renderer defect.

CRITERION: AC-03 Every HTML document offers a keyboard-accessible theme toggle with an accurate accessible name and state, follows the OS preference before an explicit choice, remembers choices on reload, and remains usable with denied storage, offline or as a local file.
VERDICT: met
EVIDENCE: At 4aa59e2, independently executed passing OS-theme changes, keyboard activation, explicit-choice persistence, denied-storage, offline and composition-only local-file cases. Accessible name/state and synchronous pre-style initialization remain intact. Previously exercised keyboard toggling, persisted reload and light restoration on all three real private HTTPS layouts; the reviewed diff changes no theme implementation.

CRITERION: AC-04 Dark mode uses the Postbode gradient with legible text, controls, code and focus states, while light mode retains the existing colours; all three layouts remain usable at 360px, 390px, 768px and desktop widths.
VERDICT: met
EVIDENCE: At 4aa59e2, independently executed passing exact Postbode gradient, native-control, composited text/control/focus contrast and four-width geometry checks for all three layouts. Light ink remains unchanged. Previously inspected newly captured dark-mobile screenshots and the desktop reference, and verified the reusable VH Midnight guide. The HTML palette and layout files are unchanged from those actual HTTPS and screenshot checks.

CRITERION: AC-05 Without JavaScript all documentation stays readable under the OS colour preference with inactive controls hidden, and printing uses a light palette with documentation content available.
VERDICT: met
EVIDENCE: At 4aa59e2, independently executed passing JavaScript-disabled light/dark and print-after-zero-match cases for every layout. Server-rendered documentation remains available, inactive controls are hidden, effective theme metadata follows the OS without JavaScript, and print reveals all fields with light colours. No fallback or print implementation changed in the reviewed delta.

CRITERION: AC-06 Existing format and source-copy behavior passes regression checks, and generated private HTTPS previews demonstrate the requirement filter and both themes.
VERDICT: met
EVIDENCE: At 4aa59e2, the independent full formatter suite passes18 cases/336 assertions cleanly in8.38 seconds, including exact multiline source copying and AsciiDoc/template regressions. Inspected the successful final make-check log. The real private HTTPS files were independently compared byte-for-byte with clean95124b3 CLI generation and exercised for filtering, deep links, both themes, four widths and exact copying. Current changes only strengthen test synchronization and add semantic-error source context; successful-input rendering and all HTML files are unchanged, so that HTTPS proof remains applicable. Earlier timing/quota and intermittent-test failures remain preserved.


## Sources

- Postbode `99dfd65b2cf44411d6ef8a920fe95e874b1adf0f`, `internal/webui/templates/index.html:13-25`: gradient, surfaces and text colours.
- `vhco-pro/docs@c98ce2f926b7e0e1c29388be95cf7d8133a298e2`, `conventions/feature-documents.md`: specification, review and verification procedure.
- [Frontend design](https://github.com/anthropics/skills/blob/dbd4588f9e1033efb41dad4bef2f7947c8993d44/skills/frontend-design/SKILL.md): preserve the owner's visual direction and task-specific hierarchy.
- [Interface rules](https://github.com/vercel-labs/web-interface-guidelines/blob/434b7f91364665f2f733b310ec54809bf8f37937/command.md): accessible controls, URL state, dark native controls and rendered review.
