---
status: complete
status_description: "All three layouts and source readability passed independent re-review and the complete gate."
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

- [x] AC-01: Configuration and CLI select reference, explorer or compact HTML, default to reference, reject unknown values, and preserve custom templates and other output formats.
      Evidence: At `d0f1fdb`, Config/CLI regressions passed for default/reference, explorer, compact, invalid values, overrides, templates and unchanged other formats. Independently confirmed in R-4.
- [x] AC-02: Every layout retains enabled API, field, constraint, reference, printer-column, composition, patch, scenario and example documentation, including multiple API identities and versions, while respecting presentation settings.
      Evidence: At `d0f1fdb`, Rich-model and filter tests passed for all layouts, including multiple identities/versions and complete enabled model sections. Independently confirmed in R-4.
- [x] AC-03: All layouts work offline with JavaScript and remain readable without it; resource/version selection, field search, empty states, explorer inspection and reference example switching work through keyboard-accessible controls.
      Evidence: At `d0f1fdb`, Eight browser cases with 106 assertions passed offline/no-JavaScript, selection, search, empty-state, reload, inspection, copy and mobile views. Independently confirmed in R-4.
- [x] AC-04: Multiline templates and programs display and copy with real line breaks; readable manifest/settings output preserves scalar types and intentional literal escapes, and malicious text cannot become executable HTML.
      Evidence: At `d0f1fdb`, Source/clipboard tests passed for LF, CRLF, tabs, Unicode, literal backslashes and HTML injection; actual 64-line template copied exactly over HTTPS. Independently confirmed in R-4.
- [x] AC-05: At 360px, 390px, 768px and desktop widths, the layouts remain usable without page overflow; phones present readable field definitions, usable controls and accessible field/example navigation.
      Evidence: At `d0f1fdb`, Chromium checks passed at 360px, 390px, 768px and desktop; R-4 independently exercised all three private phone previews. Independently confirmed in R-4.
- [x] AC-06: Generated output is deterministic, field/navigation anchors are unique and correct across API identities and versions, and selection/search/deep-link state cannot strand hidden content.
      Evidence: At `d0f1fdb`, The reproduced heading/field collision now passes; repeated bytes match and all generated fragment targets resolve, with reload/selection regressions passing. Independently confirmed in R-4.
- [x] AC-07: The CLI-generated examples pass formatter, regression, lint and real-render checks, and all three layouts have verified private tailnet previews.
      Evidence: At `d0f1fdb`, make check, integration-tag lint, formatter/browser suite and full pinned real-render matrix passed; matrix completed without skips in 342.904s and private HTTPS files matched generated bytes. Independently confirmed in R-4.

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
- [x] T2 `internal/engine/`, `cmd/` - implement the shared renderer, configuration, readable sources and interactions; the new regressions pass (AC-01 through AC-06).
      Evidence: New configuration/source/model regressions and seven Chromium/AsciiDoc/template cases passed, including 98 assertions across all three layouts, four widths, plain HTML and reload/copy behavior.
- [x] T5 `internal/engine/`, `tests/format/` - add regressions and fix R-3 ID namespaces, mobile feedback/labels and CRLF source copying (AC-03 through AC-06).
      Evidence: Watched the ID-collision, CRLF source and mobile empty-state regressions fail, then pass after namespace, escaped-code and pane-state fixes; eight browser cases with 106 assertions now pass.
- [x] T3 `examples/platform/config.yml`, `README.md`, `tests/format/` - document usage, generate all three examples and execute required quality checks (AC-07).
      Evidence: README/config examples match the CLI; all three config-only outputs match their HTTPS previews. Required quality checks and the final real-render matrix passed.
- [x] T4 `docs/specs/html-layouts.spec.md` - record a fresh independent implementation review and completion evidence, then pass the complete gate (AC-07).

      Evidence: R-4 exact records passed review check; all seven criteria met and vh-adlc gate complete returned pass.

## Open Questions

None. The owner requested all three concepts, configuration selection, readable source and a
fresh review. Existing branch placement and local-review publication boundaries continue.

## Review Log

| Date | Review | Kind | Commit | Model | Isolation | Outcome |
| --- | --- | --- | --- | --- | --- | --- |
| 2026-10-28 | R-1 | spec | `cda8c93` | unknown | procedural | Circular review prerequisite needs correction. |
| 2026-10-28 | R-2 | spec | `bb9a5cc` | unknown | procedural | No blocking findings; move historical review records under Review Log. |
| 2026-10-28 | R-2 | spec | `bb9a5cc` | n/a | n/a | accepted by owner |
| 2026-10-28 | R-3 | implementation | `0a4d5e3` | unknown | procedural | Anchor, mobile empty-state and source-fidelity fixes required; model identity not attested. |
| 2026-10-28 | R-4 | implementation | `d0f1fdb` | unknown | procedural | All seven criteria met; no findings. Model identity not attested. |
| 2026-10-29 | R-5 | implementation | `95124b3` | unknown | procedural | All seven criteria met at the preceding candidate; final committed repairs require re-review. |

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

### R-3

CRITERION: AC-01 Configuration and CLI select reference, explorer or compact HTML, default to reference, reject unknown values, and preserve custom templates and other output formats.
VERDICT: met
EVIDENCE: At 0a4d5e3535374bb01253e1880d06f4013c6a0390, fresh race tests for engine, cmd and public generator packages passed. Inspected defaults, enum validation, CLI override, custom-template precedence and structured-format compatibility tests.

CRITERION: AC-02 Every layout retains enabled API, field, constraint, reference, printer-column, composition, patch, scenario and example documentation, including multiple API identities and versions, while respecting presentation settings.
VERDICT: met
EVIDENCE: At 0a4d5e3535374bb01253e1880d06f4013c6a0390, fresh rich-model and presentation-filter tests passed for all three layouts. Inspected server-rendered API/field/constraint/reference/printer-column content and reuse of composition, patch, scenario and example sections; multiple API identities and versions are covered.

CRITERION: AC-03 All layouts work offline with JavaScript and remain readable without it; resource/version selection, field search, empty states, explorer inspection and reference example switching work through keyboard-accessible controls.
VERDICT: not met
EVIDENCE: At 0a4d5e3535374bb01253e1880d06f4013c6a0390, inspected the passing seven-case browser log and exercised the private explorer preview directly. At 390px, Browse fields followed by a query with no matches leaves zero visible field links and zero visible empty-state messages because the message is inside the hidden inspector pane.

CRITERION: AC-04 Multiline templates and programs display and copy with real line breaks; readable manifest/settings output preserves scalar types and intentional literal escapes, and malicious text cannot become executable HTML.
VERDICT: not met
EVIDENCE: At 0a4d5e3535374bb01253e1880d06f4013c6a0390, fresh source-readability/security tests passed for LF and literal backslash sequences. A network-blocked local Chromium check on a synthetic CRLF program observed DOM textContent line1\nline2\n instead of the original line1\r\nline2\r\n. Copy uses textContent, so exact source contents change.

CRITERION: AC-05 At 360px, 390px, 768px and desktop widths, the layouts remain usable without page overflow; phones present readable field definitions, usable controls and accessible field/example navigation.
VERDICT: not met
EVIDENCE: At 0a4d5e3535374bb01253e1880d06f4013c6a0390, inspected passing four-width overflow tests and exercised the mobile explorer preview. After selecting a field from Browse fields, the inspector is visible and browser hidden, but the control still says Return to field. The mobile control no longer describes its action.

CRITERION: AC-06 Generated output is deterministic, field/navigation anchors are unique and correct across API identities and versions, and selection/search/deep-link state cannot strand hidden content.
VERDICT: not met
EVIDENCE: At 0a4d5e3535374bb01253e1880d06f4013c6a0390, independently generated all three layouts from a minimal two-field XRD plus a valid Composition named api-0-field. Each output contains duplicate id api-0-field-1 because API field IDs collide with the shared section-heading allocator. Existing determinism and reload tests do not cover this collision.

CRITERION: AC-07 The CLI-generated examples pass formatter, regression, lint and real-render checks, and all three layouts have verified private tailnet previews.
VERDICT: not verified
EVIDENCE: At 0a4d5e3535374bb01253e1880d06f4013c6a0390, inspected successful make-check, browser and integration-lint evidence and directly reached the private integrated explorer preview. The complete real-render run was still progressing in /tmp/crossplane-html-matrix.log when this review closed, so the composite verification requirement remains pending.

FINDING: F-01
KIND: factual
SEVERITY: blocking
WHERE: internal/engine/html.go field IDs and internal/engine/format.go heading IDs
EVIDENCE: A minimal XRD with spec.alpha and spec.beta plus a Composition named api-0-field generates duplicate id api-0-field-1 in reference, explorer and compact. The field and composition heading use the same ID namespace.
PROPOSAL: Give shared section headings and API fields distinct ID namespaces, or use one document-wide allocator. Add a generated multi-section collision regression and verify navigation targets remain unique.

FINDING: F-02
KIND: factual
SEVERITY: blocking
WHERE: internal/engine/html.js filterFields and mobile explorer markup
EVIDENCE: At 390px in the integrated explorer preview, Browse fields followed by no-matching-field leaves zero visible field links and zero visible empty-state messages. The empty message lives in field-list, which browsing mode hides.
PROPOSAL: Present empty-state feedback in whichever explorer pane is visible. Add a mobile browse-mode no-results regression, including clearing the query and returning to a field.

FINDING: F-03
KIND: factual
SEVERITY: blocking
WHERE: internal/engine/html.go writeHTMLSource and internal/engine/format.go HTML code blocks
EVIDENCE: A YAML quoted source string containing CRLF reaches the generator with CRLF but becomes LF in the browser DOM. The local Chromium check reported exactCRLF=false, and copy reads the changed textContent.
PROPOSAL: Preserve carriage returns with HTML character references in a dedicated escaped-code rendering helper. Test copied LF, CRLF, tabs, Unicode and intentional literal backslash sequences without globally replacing escapes.

FINDING: F-04
KIND: factual
SEVERITY: should-fix
WHERE: internal/engine/html.js field-link selection and selectPanel
EVIDENCE: After a mobile explorer field link is selected, browsing is removed and the inspector is visible, but the button retains Return to field. Its next action opens the browser, so its visible label describes the opposite direction.
PROPOSAL: Synchronize the browse button label and expanded state whenever field or API selection changes the explorer pane. Cover selection and API/version transitions in the mobile browser tests.

### R-4

CRITERION: AC-01 Configuration and CLI select reference, explorer or compact HTML, default to reference, reject unknown values, and preserve custom templates and other output formats.
VERDICT: met
EVIDENCE: At d0f1fdb4b22bacd03da0514d2aa81ce173386150, fresh race tests for engine, cmd and public generator packages passed. Inspected default/reference selection, all three configured layouts, unknown-value rejection, explicit CLI override, custom-template precedence and structured/text-format compatibility tests.

CRITERION: AC-02 Every layout retains enabled API, field, constraint, reference, printer-column, composition, patch, scenario and example documentation, including multiple API identities and versions, while respecting presentation settings.
VERDICT: met
EVIDENCE: At d0f1fdb4b22bacd03da0514d2aa81ce173386150, fresh rich-model and presentation-filter tests passed across all layouts. Inspected server-rendered API identities/versions, fields, constraints, references, printer columns and reuse of the complete composition, patch, scenario and example writer. Presentation filters and unchanged structured model are covered.

CRITERION: AC-03 All layouts work offline with JavaScript and remain readable without it; resource/version selection, field search, empty states, explorer inspection and reference example switching work through keyboard-accessible controls.
VERDICT: met
EVIDENCE: At d0f1fdb4b22bacd03da0514d2aa81ce173386150, inspected the completed eight-case browser run with 106 assertions, including offline and JavaScript-disabled behavior, selection, search, reload and source copying. Independently exercised the refreshed private explorer at 390px: a no-results search in Browse fields now shows visible feedback, clearing it restores selection, and field inspection works.

CRITERION: AC-04 Multiline templates and programs display and copy with real line breaks; readable manifest/settings output preserves scalar types and intentional literal escapes, and malicious text cannot become executable HTML.
VERDICT: met
EVIDENCE: At d0f1fdb4b22bacd03da0514d2aa81ce173386150, fresh source/security regressions passed for mixed LF, CRLF, tabs, Unicode, literal backslashes and closing-script text. Independently regenerated the synthetic CRLF fixture and checked it in network-blocked local Chromium; DOM textContent exactly preserved line1\r\nline2\r\n. Inspected type-preserving YAML settings and escaped source rendering.

CRITERION: AC-05 At 360px, 390px, 768px and desktop widths, the layouts remain usable without page overflow; phones present readable field definitions, usable controls and accessible field/example navigation.
VERDICT: met
EVIDENCE: At d0f1fdb4b22bacd03da0514d2aa81ce173386150, inspected passing browser checks at 360px, 390px, 768px and desktop widths. Independently used all three refreshed private previews at 390px: reference field/example switching worked, compact fields stacked without page overflow, and explorer returned to Browse fields with aria-expanded=false after field selection.

CRITERION: AC-06 Generated output is deterministic, field/navigation anchors are unique and correct across API identities and versions, and selection/search/deep-link state cannot strand hidden content.
VERDICT: met
EVIDENCE: At d0f1fdb4b22bacd03da0514d2aa81ce173386150, fresh heading/field namespace, determinism and reload regressions passed. Independently generated the original collision fixture twice in each layout: bytes matched, no duplicate IDs remained and every fragment target existed. Inspected field/API/version transitions and visible empty states after search.

CRITERION: AC-07 The CLI-generated examples pass formatter, regression, lint and real-render checks, and all three layouts have verified private tailnet previews.
VERDICT: met
EVIDENCE: At d0f1fdb4b22bacd03da0514d2aa81ce173386150, observed the final pinned real-render suite complete with PASS in 342.904 seconds and no skips. Inspected successful final make-check, integration-lint and eight-case formatter/browser evidence. Independently reached and exercised the reference, explorer and compact private HTTPS previews under /integrated/ on fedora-43.tail6948f.ts.net:8954. The reviewed feature tree is clean.

## Sources

- [Anthropic frontend-design](https://github.com/anthropics/skills/blob/dbd4588f9e1033efb41dad4bef2f7947c8993d44/skills/frontend-design/SKILL.md): content-first hierarchy, deliberate layouts and visual verification.
- [Vercel interface guidelines](https://github.com/vercel-labs/web-interface-guidelines/blob/434b7f91364665f2f733b310ec54809bf8f37937/command.md): keyboard, focus, touch, overflow and content-handling checks.


### R-5

CRITERION: AC-01 Configuration and CLI select reference, explorer or compact HTML, default to reference, reject unknown values, and preserve custom templates and other output formats.
VERDICT: met
EVIDENCE: At 95124b34ce929b84464f4de9548c922f5b08aef2, fresh race tests for cmd and internal/engine passed, including layout defaults, enum rejection, configuration and CLI overrides, custom-template precedence and structured-output compatibility. Inspected config validation, CLI flag handling and formatter dispatch.

CRITERION: AC-02 Every layout retains enabled API, field, constraint, reference, printer-column, composition, patch, scenario and example documentation, including multiple API identities and versions, while respecting presentation settings.
VERDICT: met
EVIDENCE: At 95124b34ce929b84464f4de9548c922f5b08aef2, fresh engine tests passed for rich-model coverage, all layouts, presentation controls, references, root/container constraints, printer columns and deterministic output. Inspected source and the example-inventory regression covering composite, claim and unmatched examples, section ordering and unchanged shared-model data.

CRITERION: AC-03 All layouts work offline with JavaScript and remain readable without it; resource/version selection, field search, empty states, explorer inspection and reference example switching work through keyboard-accessible controls.
VERDICT: met
EVIDENCE: At 95124b34ce929b84464f4de9548c922f5b08aef2, inspected the executed 18-case, 336-assertion browser log and its test implementation. Independently exercised all three private HTTPS layouts: requirement/search filtering, empty states, exact source copying and readable panels with JavaScript disabled passed. Keyboard activation of the reference example inventory revealed its manifest on mobile and after reload.

CRITERION: AC-04 Multiline templates and programs display and copy with real line breaks; readable manifest/settings output preserves scalar types and intentional literal escapes, and malicious text cannot become executable HTML.
VERDICT: met
EVIDENCE: At 95124b34ce929b84464f4de9548c922f5b08aef2, fresh source-program tests passed for LF, CRLF, tabs, Unicode, literal escapes and escaped hostile HTML. Inspected the passing browser clipboard and Asciidoctor checks, then independently copied the source program from each HTTPS layout and compared its exact text. Both repaired SVG assets parsed as native vectors without script, foreignObject, image or external references; browser character measurements preserved four-space YAML indentation.

CRITERION: AC-05 At 360px, 390px, 768px and desktop widths, the layouts remain usable without page overflow; phones present readable field definitions, usable controls and accessible field/example navigation.
VERDICT: met
EVIDENCE: At 95124b34ce929b84464f4de9548c922f5b08aef2, independently opened each private HTTPS layout at 360px, 390px, 768px and 1280px and observed no page overflow or browser errors. Inspected passing browser checks for field geometry and mobile field/example controls; independently verified the reference inventory link and manifest visibility at phone width.

CRITERION: AC-06 Generated output is deterministic, field/navigation anchors are unique and correct across API identities and versions, and selection/search/deep-link state cannot strand hidden content.
VERDICT: met
EVIDENCE: At 95124b34ce929b84464f4de9548c922f5b08aef2, fresh engine tests passed repeated-byte determinism and unique field/heading IDs across API identities. Inspected passing executed browser tests for selection, filtering, hash/query deep links, reload, arbitrary-schema defaults and example inventory navigation. Independently exercised the HTTPS reference inventory by keyboard and reloaded its target without losing visibility.

CRITERION: AC-07 The CLI-generated examples pass formatter, regression, lint and real-render checks, and all three layouts have verified private tailnet previews.
VERDICT: met
EVIDENCE: At 95124b34ce929b84464f4de9548c922f5b08aef2, fresh targeted race tests passed; inspected successful make-check output with zero lint issues and the 18-case, 336-assertion formatter log. Compared full-render and targeted-recheck logs with source: all 14 declared integration tests have passing records; the first run failed from host Docker address-pool exhaustion and the affected groups passed recovery, so default allocation on the exhausted host is not established. Independently built the current CLI and matched all three private HTTPS preview bodies byte for byte, then exercised their live behavior. The initial target was clean; the controller began follow-up edits after these checks, so completion requires review of the final committed state.
