---
status: in-progress
status_description: "Reviewed and accepted; implementing the three configured HTML layouts."
branch: feat/v0.2.0
issue: null
goal: "Generate three readable, configurable HTML reference layouts with legible embedded source programs."
covers:
  - internal/engine/
  - cmd/
  - tests/format/
  - examples/platform/config.yml
  - README.md
---

# Configurable HTML references

## Context

The current HTML formatter opens with four metric cards and renders fields as a wide table
(`internal/engine/format.go:75`, `:162`). The owner approved three working design concepts:
a reference reader, a schema explorer and a compact reference. The concept source-settings
view uses JSON serialization, which escapes embedded template and program newlines
(`/tmp/crossplane-docs-design.py`, function `extra`). The generated source must be readable
as a program rather than as a quoted JSON string.

This extends the completed documentation-engine feature on its existing `feat/v0.2.0`
branch, as requested by the owner. The original feature's review remains historical evidence;
the changed formatter receives a fresh specification and implementation review here.

## Scope

Deliver all three concepts through the existing HTML formatter, preserving the teal/navy
palette and the shared documentation model. Include configuration, CLI override, readable
multiline source, mobile behavior, offline use, examples, tests and a fresh independent review.
Keep work on the existing local feature branch. Remote publication and live releases remain
separate actions. Shared frontend-design guidance is a documentation change in `vhco-pro/docs`.

## Design

Add `html-layout` with values `reference`, `explorer` and `compact`; default to `reference`.
Add `--html-layout` with explicit CLI values overriding configuration. Keep `formatter: html`
and the existing `layout: table|document` text-format setting. Custom templates retain their
existing precedence; JSON/YAML continue to serialize the unchanged documentation model.

Use one HTML renderer and shared assets, with view-specific structure where needed. Render
schema facts and all enabled composition/scenario/example sections on the server. Enhance
navigation, version/resource selection, search, field inspection and copying with embedded
JavaScript. The file needs no network requests. Without JavaScript, all enabled content stays
readable and navigation anchors work. IDs must remain unique across APIs, versions and fields.

The reference reader pairs field definitions and a matching example. The explorer provides a
field browser and inspector. Compact mode uses comparison columns on desktop and stacked
field definitions on phones. Mobile reference mode offers Fields and Example YAML views;
mobile explorer mode offers a field browser and inspector. Preserve full paths and parent-relative
required semantics. Respect enabled sections, columns, descriptions, examples, patches,
nesting and anchors; display cross-field validation, schema references and printer columns.

Use readable YAML for manifest/settings objects and separate preformatted blocks for embedded
program strings, including template, KCL, CUE and Python inputs. Preserve exact string contents,
including intentional literal backslash-n sequences. Never repair presentation by globally
replacing escapes. Escape HTML and retain the existing credential redaction boundary.

## Acceptance Criteria

- [ ] AC-01: Configuration and CLI select reference, explorer or compact HTML, default to reference, reject unknown values, and preserve custom templates and other output formats.
- [ ] AC-02: Every layout retains enabled API, field, constraint, reference, printer-column, composition, patch, scenario and example documentation, including multiple API identities and versions, while respecting presentation settings.
- [ ] AC-03: All layouts work offline with JavaScript and remain readable without it; resource/version selection, field search, empty states, explorer inspection and reference example switching work through keyboard-accessible controls.
- [ ] AC-04: Multiline templates and programs display and copy with real line breaks; readable manifest/settings output preserves scalar types and intentional literal escapes, and malicious text cannot become executable HTML.
- [ ] AC-05: At 360px, 390px, 768px and desktop widths, the layouts remain usable without page overflow; phones present readable field definitions, usable controls and accessible field/example navigation.
- [ ] AC-06: Generated output is deterministic, field/navigation anchors are unique and correct across API identities and versions, and selection/search/deep-link state cannot strand hidden content.
- [ ] AC-07: The CLI-generated examples pass formatter, regression, lint and real-render checks, and all three layouts have verified private tailnet previews.

## Test Plan

| Criterion | Proof |
| --- | --- |
| AC-01 | Config and CLI regression tests exercise defaults, enum rejection, explicit overrides and custom-template/structured-output compatibility. |
| AC-02 | Engine tests use a rich document with multiple API identities/versions and all model sections; browser tests verify presentation filters and source/scenario facts. |
| AC-03 | Offline and JavaScript-disabled Chromium checks exercise native content, keyboard navigation, selection, search, empty states, copying and mobile controls. |
| AC-04 | Unit and browser regressions compare actual multiline source text and copied content, preserve literal escapes and types, and reject script/HTML injection. |
| AC-05 | Browser checks at four widths verify field readability, touch targets, navigation and absence of page/field horizontal scrolling. |
| AC-06 | Repeated generation compares bytes; multi-identity tests inspect duplicate IDs, navigation targets, reloadable URLs and selection transitions. |
| AC-07 | Run `make check`, `bun test --cwd tests/format`, integration-tag lint and the complete real-render matrix; independently review the exact commit and verify generated previews over the private HTTPS route. |

## Tasks

- [x] T1 `internal/engine/`, `cmd/`, `tests/format/` - write and observe failing config, multiline-source and three-layout browser regressions (AC-01 through AC-06).
      Evidence: Targeted Go regressions failed on unknown html-layout configuration and --html-layout flag; all three new browser cases failed on the missing flag before implementation.
- [ ] T2 `internal/engine/`, `cmd/` - implement the shared renderer, configuration, readable sources and interactions; the new regressions pass (AC-01 through AC-06).
- [ ] T3 `examples/platform/config.yml`, `README.md`, `tests/format/` - document usage, generate all three examples and execute required quality checks (AC-07).
- [ ] T4 `docs/specs/html-layouts.spec.md` - record a fresh independent implementation review and completion evidence, then pass the complete gate (AC-07).

## Open Questions

None. The owner requested all three concepts, configuration selection, readable source and a
fresh review. Existing branch placement and local-review publication boundaries continue.

## Review Log

| Date | Review | Kind | Commit | Model | Isolation | Outcome |
| --- | --- | --- | --- | --- | --- | --- |
| 2026-10-28 | R-1 | spec | `cda8c93` | unknown | procedural | Circular review prerequisite needs correction. |
| 2026-10-28 | R-2 | spec | `bb9a5cc` | unknown | procedural | No blocking findings; move historical review records under Review Log. |
| 2026-10-28 | R-2 | spec | `bb9a5cc` | n/a | n/a | accepted by owner |

### R-1

FINDING: F-01
KIND: factual
SEVERITY: blocking
WHERE: Acceptance Criteria AC-07
EVIDENCE: AC-07 requires that "the final committed implementation receives a clean independent review". The pinned conventions/feature-documents.md requires an implementation review to establish every current criterion as met before that review is clean. The first review cannot observe an existing clean review of this criterion, so making that review an acceptance condition introduces a circular prerequisite.
PROPOSAL: Keep the independent review in the Test Plan, T4 and the standard complete gate. State AC-07 in terms of the executable quality checks and the three private previews, without requiring a clean review as a criterion that the same review must establish.

### R-2

FINDING: F-01
KIND: factual
SEVERITY: should-fix
WHERE: Review Log and Sources
EVIDENCE: At bb9a5cc4100654a70a829f323ed06e56e1b2af39, the recorded R-1 subsection follows the Sources heading and source links. The pinned feature-document standard places each returned review subsection under Review Log and Sources last. The corrected AC-07 removes the circular prerequisite; no blocking normative finding remains.
PROPOSAL: Move the complete R-1 subsection before Sources without changing its returned records. This is a review-storage correction and does not change the normative baseline.

## Sources

- [Anthropic frontend-design](https://github.com/anthropics/skills/blob/dbd4588f9e1033efb41dad4bef2f7947c8993d44/skills/frontend-design/SKILL.md): content-first hierarchy, deliberate layouts and visual verification.
- [Vercel interface guidelines](https://github.com/vercel-labs/web-interface-guidelines/blob/434b7f91364665f2f733b310ec54809bf8f37937/command.md): keyboard, focus, touch, overflow and content-handling checks.
