---
status: complete
status_description: "All v0.2.0 criteria passed independent review; the local feature branch is ready for owner review."
branch: feat/v0.2.0
issue: null
goal: "Generate configurable, tested documentation for Crossplane APIs and composition pipelines in one v0.2.0 release."
covers:
  - cmd/
  - pkg/
  - internal/
  - examples/
  - tests/
  - scripts/
  - main.go
  - go.mod
  - go.sum
  - Makefile
  - README.md
  - AGENTS.md
  - CLAUDE.md
  - .github/
  - .goreleaser.yml
  - .golangci.yml
  - .gitignore
  - renovate.json5
---

# Crossplane documentation engine, v0.2.0

## Context

The current XRD model handles basic properties but omits map schemas, schema
composition and string constraints (`pkg/generator/generator.go:67`). Generation
chooses the first served version (`pkg/generator/generator.go:124`). Composition
inspection reads static `input.resources` entries
(`pkg/composition/composition.go:175`); that cannot describe arbitrary generated
resources. The command tree currently exposes only `xrd` and `composition`
(`cmd/root.go:29`).

Crossplane functions can use templates, programming languages and arbitrary
custom implementations. Upstream rendering executes the pipeline with a sample
XR and can accept mocked state and context. Documentation therefore needs both
an API/source reference and concrete rendered examples. Each advertised function
type needs real end-to-end proof, including the resulting documentation.

One feature branch, `feat/v0.2.0`, carries this release. The existing maintenance
changes are retained there. This is the single specification and implementation
plan; completed maintenance reports and separate progress documents are omitted.
The branch placement follows the owner's explicit request to keep this revamp
together rather than publishing the specification separately on main.

## Scope

Deliver all of the following in v0.2.0:

- Complete XRD references with version selection, nested collections, validation
  metadata and schema references.
- Optional schema/CEL validation of supplied and rendered examples using the
  upstream validator, with real valid/invalid fixture tests.
- Composition references for legacy resources and modern function pipelines,
  including patch sets, policies, connection details, readiness and environment
  configuration.
- Go templating, KCL, CUE, Python, patch-and-transform, mixed pipelines and custom
  functions through a generic render integration, with runnable examples and real
  render tests for every advertised type.
- Markdown, standalone HTML, AsciiDoc, JSON and YAML output, with custom templates
  and configuration for presentation and file management.
- Directory discovery, YAML bundles, README injection and a drift check for CI.
- Reproducible integration tests before versioning, then signed binary releases.

Preserve the existing `xrd` and `composition` commands and their flags. Add
capabilities to the tested implementation rather than replacing working behavior
with a new scaffold. Function execution requires an explicit render request.
Documentation generation without rendering continues to work without Docker or a
cluster. Rendering must not require a live provider or cloud credentials.

## Design

### Shared documentation model

Parse manifests into one model containing APIs, versions, field paths, constraints,
pipeline steps, function input/source references, resource definitions, mappings,
policies and render scenarios. Every formatter consumes this model. JSON and YAML
expose a versioned document schema, separate from the application version.

Preserve provenance: schema facts, declared resources and scenario-rendered
resources remain identifiable. Rendered output describes the supplied scenario;
it must not present one conditional branch as every resource the pipeline can
ever create. Resource and field identities include their parent paths. Repeated
generation with identical inputs produces identical bytes.

### XRD schema coverage

Support XRD v1 and v2 metadata, scope, claims where applicable, served and
referenceable versions, printer columns, and spec/status schemas. Keep the
current default version selection; add `--api-version` and `--all-versions`.
Reject an unknown selection rather than falling back silently.

Expand object properties, array items and additionalProperties with unambiguous
paths such as `spec.networks[].subnets[].cidr` and `spec.tags.*`. Preserve required
semantics at each parent, types, nullable values and defaults, including false,
zero and empty collections. Render enums, numeric and string bounds, patterns,
formats, list/map semantics, CEL rules and their messages. Show oneOf, anyOf,
allOf and not as schema alternatives rather than flattening them into false
unconditional requirements. Resolve local and explicitly supplied file references;
detect cycles and missing references. Do not fetch remote schemas automatically.
Displaying CEL or a pattern does not mean executing that validation.

An explicit validation mode delegates to the pinned upstream
`crossplane resource validate` command with supplied XRDs/CRDs or explicitly
selected extension packages. Evaluate schema constraints, unknown fields and CEL
against the scenario resources. Missing required schemas fail; invalid examples
cannot be presented as validated. The test suite supplies local schemas and runs
real valid/invalid cases. Rendering alone does not imply validation passed.

### Composition references and execution

Dispatch source adapters by input API version/kind and resolved Function metadata,
not a hardcoded functionRef name. Show every step in pipeline order. For known
inputs, document their source and relevant settings; for unknown functions,
preserve the input reference and use the same generic render path.

Patch-and-transform references expand patch sets, preserve patch direction and
combine variables, and explain transform chains in readable text with original
values available in structured output. Include policies, readiness checks,
connection details and environment inputs. Detect missing and cyclic patch-set
references. Real rendering, rather than a second local interpreter, establishes
transform results.

For Go templating, document inline templates, template lists, filesystem sources
and environment sources. For KCL, document inline, file and external package
sources and their targets. External or image-contained sources are identified
without pretending they were read from the caller's filesystem. Controlled test
images, mounted assets or a local test registry supply those sources in CI.
KCL filesystem loading needs explicit runtime setup; it is not equivalent to
passing an arbitrary host path to the standard renderer.

An explicit `--render` operation runs a verified Crossplane CLI with the sample
XR, Composition and Functions. Provide scenario inputs for observed resources,
required resources/schemas and context. Select the invocation for the pinned CLI
version; current upstream documentation uses `crossplane composition render`.
Pass arguments without shell interpolation, bound execution with a timeout, and
report subprocess failures. A failed render cannot become a successful static
fallback. Keep existing output files intact on generation failure.

Also accept saved rendered YAML so users can document results produced by their
existing pipeline. Separate composite/status output, composed resources, context
and function results. Include scenario identity and inputs used, while excluding
credentials from generated documentation. Match resources by composition resource
identity, including resources whose Kubernetes names are generated.

### Output and configuration

Built-in formatters:

| Format | Contract |
| --- | --- |
| Markdown | Table and document layouts with anchors and readable nested paths |
| HTML | A complete responsive document with embedded styles, navigation, accessible tables and escaped input; opens offline without a CDN |
| AsciiDoc | Native headings, tables and code blocks suitable for Asciidoctor |
| JSON | A parseable, versioned representation of the shared model |
| YAML | The same data contract as JSON, without lossy string conversion |

Custom templates use the shared model. Use text/template for text output and
html/template for HTML; templates do not invoke shell commands. Support optional
header/footer files, section ordering and visibility, field columns, sorting,
required-first ordering, nesting, anchors, descriptions, examples and HTML styling.
Defaults remain useful without configuration. Structured output retains the data
contract when presentation sections are hidden.

Use `.crossplane-docs.yml` with a configuration schema version. `--config` selects
an explicit file; otherwise look only in the input root, then the invocation
directory. Explicit flags override configuration, which overrides defaults,
including explicit false values. Resolve config-owned paths relative to its
directory. Unknown keys, invalid types and incompatible format/output choices
fail with a useful error. Do not copy terraform-docs options that only make sense
for Terraform; borrow its familiar formatter/sections/settings/sort/output model.

Add `generate PATH` for files or directories. Batch mode discovers YAML in stable
path/document order, supports includes/excludes and recursion, skips unrelated
Kubernetes kinds, and reports malformed supported inputs with their locations.
Single-resource commands retain strict kind checks. Match XRDs, Compositions and
example XRs by API identity. Avoid duplicate inputs and output-name collisions;
do not follow directory symlinks recursively.

Support stdout, replace-file output and marked-region injection for Markdown,
HTML and AsciiDoc. Injection preserves surrounding bytes, is idempotent, and
rejects missing, duplicate or reversed markers without modifying the file.
JSON/YAML use stdout or replace mode. `--check` returns 0 when current, 1 for
documentation drift and 2 for invalid input/configuration/rendering; it never
writes. Complete generation and validation before replacing output files.

### Examples, tests and release

Use one realistic platform API with networking, compute/storage, nested arrays,
maps, validations, defaults and status. Reuse its XRD and XR scenarios across
composition strategies. Keep source fixtures, expected results and necessary
runtime inputs; do not add parallel reports, duplicated walkthroughs or generated
documentation that no test checks.

Each matrix row must execute actual function packages using the pinned renderer.
Compare the rendered YAML semantically, then generate documentation and assert
the resource inventory and expected field/status values. A snapshot alone cannot
substitute for rendering. Test examples through the distributed CLI as well as
package APIs. Only unstable renderer-owned metadata may be normalized, with each
normalization named in the test.

| Required scenario | Behaviors proved |
| --- | --- |
| Legacy resources | Inventory, patch direction, patch sets and reference errors; compare with explicit expected manifests because upstream function rendering does not execute legacy mode |
| Patch-and-transform | All advertised patch/transform kinds, combine, policies, defaults, readiness, connection details and environment inputs; success and error paths |
| Go templating | Inline/template lists, image filesystem and environment sources, multiple documents, conditionals, loops, context, extra resources, helpers and composite/status updates |
| KCL | Inline, packaged file and local-registry package sources; advertised targets, custom parameters, loops, conditions and observed-state access |
| CUE | Pinned public function, generated resources, input changes and invalid-program failure |
| Python | Pinned public function, generated resources, observed/required inputs, status changes and exception failure |
| Mixed pipeline | Environment loading, patch/transform and templating with shared context, then readiness; create and update scenarios |
| Custom function | A protocol-compatible function unknown to source adapters still produces documented resources through the generic renderer |

Resolve exact supported CLI/function versions, input variants and image digests
before writing adapters. Record them in test configuration maintained by Renovate.
Every variant claimed in help, examples or the README must map to a real test.
New upstream variants are not silently treated as tested; keep the declared
compatibility matrix tied to the pinned packages. No test requires actual cloud
provisioning. All advertised render scenarios are mandatory CI checks, with no
pass-by-skip behavior when Docker, images or the renderer are unavailable.

Reuse the shared Go, version and release workflows. Go checks, formatter checks
and the complete real-render matrix pass before tagging. One publisher builds
and signs the exact tagged source. Keep retry behavior and asset immutability.
Before activation, inspect the latest existing release commit and add its
equivalent v-prefixed baseline tag without moving old tags. Request v0.2.0 through
the shared Conventional Commit policy; verify the calculated version rather than
hardcoding a second counter or replacing an existing tag. Live publication and
remote tag changes require explicit shipping authorization.

## Acceptance Criteria

- [x] AC-01: Users can document v1/v2 XRD metadata and select a named version or all versions; an unknown version fails explicitly.
      Evidence: At `7749946`, `make check` passed version-selection and v1/v2 schema regression tests, including unknown-version failures. Independently confirmed by R-4.
- [x] AC-02: Nested objects, array items, maps and referenced schemas produce deterministic paths, correct parent-relative required markers and lossless defaults; missing references fail and recursive schemas terminate with an explicit reference marker.
      Evidence: At `7749946`, `make check` passed nested path, parent-relative required, default, local/file-reference and cycle regression tests. Independently confirmed by R-4.
- [x] AC-03: Documentation includes numeric/string/list constraints, enums, formats, Kubernetes list/map metadata, CEL rules/messages and schema alternatives without claiming unexecuted validation.
      Evidence: At `7749946`, `make check` passed schema metadata and formatter tests for constraints, CEL and schema alternatives. Independently confirmed by R-4.
- [x] AC-04: Composition references expand patch sets and document direction, combine inputs, readable transform chains, policies, readiness, connection details and environment inputs; invalid patch-set references fail.
      Evidence: At `7749946`, `make check` and the real patch-and-transform case passed patch-set, mapping, transform, policy and connection assertions. Independently confirmed by R-4.
- [x] AC-05: Every pipeline step and source is documented, including unknown function inputs, without presenting an incomplete declared inventory as complete runtime output.
      Evidence: At `7749946`, `make check` and the real unknown-adapter case passed; custom resources stay opaque and pipeline order/provenance are preserved. Independently confirmed by R-4.
- [x] AC-06: Explicit rendering and saved rendered YAML produce scenario-labelled resource/status documentation; observed/required resources, schemas and context reach the real renderer, while failures/timeouts preserve existing output.
      Evidence: At `7749946`, `make check` and `go run ./scripts/verify-render` passed saved-output, renderer failure/timeout, consumed mock resource/schema and context checks. Independently confirmed by R-4.
- [x] AC-07: Every advertised Go-templating source and behavior in the scenario matrix has a runnable example and passing real-render-to-documentation test, including failure cases.
      Evidence: At `7749946`, `go run ./scripts/verify-render` passed all Go-template sources and behavior variants, all five documentation formats and invalid-program checks. Independently confirmed by R-4.
- [x] AC-08: Every advertised KCL source and target has a runnable example and passing real-render-to-documentation test, including failure cases.
      Evidence: At `7749946`, `go run ./scripts/verify-render` passed inline, filesystem, OCI, all five KCL targets, parameter/mock inputs and invalid programs through documentation generation. Independently confirmed by R-4.
- [x] AC-09: CUE, Python and a custom function unknown to source adapters have passing real-render-to-documentation tests with successful and failing scenarios.
      Evidence: At `7749946`, `go run ./scripts/verify-render` passed CUE, Python and custom-adapter success/failure scenarios and output-preservation assertions. Independently confirmed by R-4.
- [x] AC-10: Mixed-pipeline create/update examples prove shared context, conditional/repeated resources, mocked observed state and readiness/status behavior through real rendering and resulting documentation.
      Evidence: At `7749946`, `go run ./scripts/verify-render` passed mixed create/update resource, context, observed-patch, Pending/Ready and readiness-condition assertions in rendered documentation. Independently confirmed by R-4.
- [x] AC-11: Markdown, offline HTML, AsciiDoc, JSON, YAML and custom templates render the same shared model correctly, with format-appropriate escaping and a tested structured-data contract.
      Evidence: At `7749946`, `make check` passed shared-model and template regressions; `bun test --cwd tests/format` passed offline desktop/mobile HTML, Asciidoctor and HTML-escaping checks. Independently confirmed by R-4.
- [x] AC-12: Configuration controls sections, columns, layout, sorting, nesting, descriptions, anchors, styling, headers/footers and output; explicit flags win and invalid configuration fails.
      Evidence: At `7749946`, `make check` passed configuration, discovery, relative-path, presentation and explicit-flag override tests, including malformed configuration rejection. Independently confirmed by R-4.
- [x] AC-13: File/directory/bundle generation supports deterministic discovery and resource matching, catches collisions and malformed supported inputs, and retains existing single-resource commands.
      Evidence: At `7749946`, `make check` passed deterministic discovery, bundles, collisions, identity matching, malformed supported inputs and command compatibility checks. Independently confirmed by R-4.
- [x] AC-14: Replace and marker-injection output preserve unrelated content and resist partial writes; repeated injection is identical and check mode distinguishes current, drift and failure without writing.
      Evidence: At `7749946`, `make check` and the actual CLI integration tests passed injection, atomic output, idempotence and all accepted boolean check forms with failure exit 2. Independently confirmed by R-4.
- [x] AC-15: Every advertised capability maps to a tested runnable example, and mandatory CI executes the full renderer matrix and formatter checks before any release tag can be created.
      Evidence: At `7749946`, The full mandatory renderer matrix passed in 362.400s without skips. R-4 verified formatter/render prerequisites and the pinned exact-commit version workflow. Independently confirmed by R-4.
- [x] AC-16: A local six-target release snapshot has verified archive contents, checksums and version metadata; local checksum signing verifies successfully, and release workflow contract checks require the exact checked tag and protect existing assets.
      Evidence: At `7749946`, `goreleaser check` and a signed local six-target snapshot passed. R-4 verified SHA-256, archive contents, version 7749946, checksum signature and pinned release contracts; no publication ran. Independently confirmed by R-4.
- [x] AC-17: Optional upstream schema/CEL validation accepts valid supplied/rendered examples, rejects invalid fields, constraints and CEL rules, fails for missing required schemas, and includes real validator tests in mandatory CI.
      Evidence: At `7749946`, `go run ./scripts/verify-render` passed real supplied/rendered validation plus invalid field type, CEL, pattern, unknown-field and missing-schema rejection. Independently confirmed by R-4.

## Test Plan

| Criterion | Proof |
| --- | --- |
| AC-01, AC-02, AC-03 | Expanded XRD fixtures and generator/CLI tests compare version selections, full field paths, constraints and reference failure behavior. |
| AC-04, AC-05 | Composition unit tests and real patch-and-transform fixtures compare references, patch results, source classification and unknown-function handling. |
| AC-06 | Runner contract tests prove argument forwarding, timeout/error behavior and output preservation; actual render scenarios prove parsed resources/status and saved-output equivalence. |
| AC-07 | Go-templating matrix rows run pinned functions for all advertised source forms, branch/loop/context cases and malformed/missing-input cases, then assert generated docs. |
| AC-08 | KCL matrix rows run pinned functions and controlled package/file sources for all advertised targets and both valid and failing programs, then assert generated docs. |
| AC-09 | CUE, Python and custom-function rows run real pinned function runtimes for success and failure and verify the docs through the generic renderer. |
| AC-10 | Mixed create/update cases render with different mock inputs and compare resource identities, values, readiness and status in YAML and docs. |
| AC-11 | Formatter goldens plus HTML parsing and an offline browser check, Asciidoctor conversion, JSON/YAML round trips and custom-template tests. |
| AC-12 | Config/CLI table tests prove defaults, discovery, path resolution, explicit false overrides, styling/sections/columns and rejection of bad keys/types/combinations. |
| AC-13 | Temporary directory and bundle tests cover discovery, excludes, symlinks, identity matching, collisions, malformed inputs and command compatibility. |
| AC-14 | Byte-preservation and idempotence tests cover replace/inject/error behavior; binary tests assert check-mode exit codes and unchanged files. |
| AC-15 | CI runs the complete pinned integration matrix and formatter checks; force one render failure to prove the version job cannot run; reconcile advertised support with tested cases. |
| AC-16 | GoReleaser check/snapshot, archive checksum and version tests; verify local checksum signing with a temporary test key and inspect exact-tag and asset-immutability contracts in the selected shared workflows. A live publication smoke test is a separately authorized shipping step. |
| AC-17 | Invoke the pinned upstream validator on valid/invalid XR and rendered-resource fixtures with local XRD/CRD schemas; assert field, constraint, CEL and missing-schema failures and resulting CLI exit behavior. |

## Tasks

- [x] T1 `tests/render/cases.yaml`, `examples/` - resolve and pin actual runtime versions/variants and add complete shared scenarios; fixture validation succeeds (AC-07, AC-08, AC-09, AC-10, AC-15).
- [x] T2 `pkg/generator/generator_test.go`, `cmd/root_test.go` - add failing version/schema/default/reference scenarios; failures identify the missing behavior (AC-01, AC-02, AC-03).
- [x] T3 `pkg/generator/generator.go`, `internal/engine/`, `cmd/generate.go` - implement shared schema extraction and version selection; T2 passes (AC-01, AC-02, AC-03).
- [x] T4 `pkg/composition/composition_test.go`, `tests/render/render_test.go` - add failing patch/reference and pipeline scenario assertions (AC-04, AC-05, AC-06).
- [x] T5 `pkg/composition/composition.go`, `internal/engine/`, `internal/engine/render.go` - implement source references, patch expansion and the generic renderer; T4 passes (AC-04, AC-05, AC-06).
- [x] T6 `tests/render/render_test.go`, `examples/` - execute the complete matrix with real runtimes and assert generated documentation; every required scenario passes without skips (AC-07, AC-08, AC-09, AC-10).
- [x] T7 `internal/engine/format.go`, `internal/engine/config.go`, `cmd/root_test.go` - add failing tests for every formatter, customization and configuration rule (AC-11, AC-12).
- [x] T8 `internal/engine/format.go`, `internal/engine/config.go`, `cmd/` - implement the formatters, configuration and template contracts; T7 passes (AC-11, AC-12).
- [x] T9 `cmd/root_test.go`, `internal/engine/output.go` - add failing discovery, bundle, injection, output-preservation and drift tests (AC-13, AC-14).
- [x] T10 `cmd/`, `internal/engine/output.go` - implement batch generation and output modes; T9 passes (AC-13, AC-14).
- [x] T13 `tests/render/render_test.go`, `internal/engine/render.go` - add failing validator/CLI scenarios for valid, invalid and missing-schema inputs (AC-17).
- [x] T14 `internal/engine/render.go`, `cmd/` - implement the explicit upstream validation mode; T13 passes against the real validator (AC-17).
- [x] T11 `.github/workflows/ci.yml`, `scripts/`, `Makefile`, `renovate.json5` - require the real-render, validator and formatter suites before versioning; the intentional-failure CI check blocks tagging (AC-15, AC-17).
- [x] T12 `.goreleaser.yml`, `README.md`, `examples/` - verify six-target snapshots, document only tested support and exact release behavior, and verify local signing and release workflow contracts (AC-15, AC-16).

## Open Questions

None. The render test matrix and the additional output/configuration capabilities
are included in the proposed release scope. Acceptance of the complete normative
baseline is separate from agreement on these individual requirements.

## Review Log

| Date | Review | Kind | Commit | Model | Isolation | Outcome |
| --- | --- | --- | --- | --- | --- | --- |
| 2026-10-27 | R-1 | spec | `d9a09ca` | unknown | procedural | Blocking publication prerequisite corrected in the next revision. |
| 2026-10-27 | R-2 | spec | `b435db6` | unknown | procedural | No findings; model identity not attested. |
| 2026-10-27 | R-2 | spec | `b435db6` | n/a | n/a | accepted by owner |
| 2026-10-27 | R-3 | implementation | `1a22ab9` | unknown | procedural | Unknown inputs, check exit codes and scenario evidence need fixes. |
| 2026-10-27 | R-4 | implementation | `7749946` | unknown | procedural | All 17 criteria met; no findings. Model identity not attested. |
| 2026-10-29 | R-5 | implementation | `95124b3` | unknown | procedural | All 17 criteria met; F-01 requires source locations in semantic manifest errors. |
| 2026-10-29 | R-6 | implementation | `4aa59e2` | unknown | procedural | All 17 criteria met; diagnostic location repair verified, no findings. |

### R-1

FINDING: F-01
KIND: factual
SEVERITY: blocking
WHERE: Test Plan AC-16 and Tasks T12
EVIDENCE: At d9a09ca, docs/specs/documentation-engine.spec.md:233 requires separate shipping authorization for live publication and remote tag changes, but line 271 requires an exact-tag signed-release smoke test and line 289 makes that check part of implementation completion. The authorized deliverable is a local feature branch ready for review, so these checks require an action outside that deliverable.
PROPOSAL: Make AC-16 completion depend on local snapshot, checksum, version, signing and workflow contract checks. Describe the live exact-tag publication smoke test as a shipping procedure that runs only after separate authorization, without making it a prerequisite for local implementation completion.

### R-2

NO FINDINGS

### R-3

CRITERION: AC-01 Users can document v1/v2 XRD metadata and select a named version or all versions; an unknown version fails explicitly.
VERDICT: met
EVIDENCE: At 1a22ab9e6c8e5d26af16fcc701f9ed93111ccab8, fresh race tests for cmd, pkg and internal packages passed. CLI generation of examples/platform.yaml with --all-versions produced v1alpha1 and v1beta1 with scope, served and referenceable metadata; version-selection and unknown-version tests passed.

CRITERION: AC-02 Nested objects, array items, maps and referenced schemas produce deterministic paths, correct parent-relative required markers and lossless defaults; missing references fail and recursive schemas terminate with an explicit reference marker.
VERDICT: met
EVIDENCE: At 1a22ab9e6c8e5d26af16fcc701f9ed93111ccab8, fresh reference, file-reference, default and generator tests passed. CLI JSON contained spec.networks[].name with required=true and spec.tags.*. Inspection of schema traversal confirmed sorted property traversal, parent-relative required lookup and reference-cycle markers.

CRITERION: AC-03 Documentation includes numeric/string/list constraints, enums, formats, Kubernetes list/map metadata, CEL rules/messages and schema alternatives without claiming unexecuted validation.
VERDICT: met
EVIDENCE: At 1a22ab9e6c8e5d26af16fcc701f9ed93111ccab8, fresh formatter and schema tests passed. CLI JSON retained minimum, minLength, maxLength, pattern and oneOf metadata; schema.go preserves the other named constraint categories. Scenario validation defaults to false.

CRITERION: AC-04 Composition references expand patch sets and document direction, combine inputs, readable transform chains, policies, readiness, connection details and environment inputs; invalid patch-set references fail.
VERDICT: met
EVIDENCE: At 1a22ab9e6c8e5d26af16fcc701f9ed93111ccab8, fresh composition and patch-reference tests passed. CLI generation for examples/platform/patch-and-transform.yaml expanded the region patch set and emitted conversion, math, combine, policy, ToCompositeFieldPath, readiness and connection details. Composition extraction and formatting preserve environment configuration.

CRITERION: AC-05 Every pipeline step and source is documented, including unknown function inputs, without presenting an incomplete declared inventory as complete runtime output.
VERDICT: not met
EVIDENCE: At 1a22ab9e6c8e5d26af16fcc701f9ed93111ccab8, a valid custom function input with apiVersion custom.example.org/v1, kind Input and resources: [alpha, beta] failed CLI generation with exit 1 and "step custom: resource entry must be an object". Unknown inputs are interpreted as patch-and-transform resources whenever they contain a resources key.

CRITERION: AC-06 Explicit rendering and saved rendered YAML produce scenario-labelled resource/status documentation; observed/required resources, schemas and context reach the real renderer, while failures/timeouts preserve existing output.
VERDICT: not verified
EVIDENCE: At 1a22ab9e6c8e5d26af16fcc701f9ed93111ccab8, fresh saved-output classification, renderer failure and cancellation tests passed. Inspection confirmed argument forwarding and generation before file replacement. The complete real-render run was still progressing during this review, and no observed test established the required-schemas path through the real renderer.

CRITERION: AC-07 Every advertised Go-templating source and behavior in the scenario matrix has a runnable example and passing real-render-to-documentation test, including failure cases.
VERDICT: untested
EVIDENCE: At 1a22ab9e6c8e5d26af16fcc701f9ed93111ccab8, TestGoTemplateSourceVariants checks parsed resource counts for template-list, environment, condition and extra-resources variants; TestFileSystemSources checks parsed counts. These variants do not generate or inspect documentation. The extra-resource variant requests resources without asserting a consumed resource value.

CRITERION: AC-08 Every advertised KCL source and target has a runnable example and passing real-render-to-documentation test, including failure cases.
VERDICT: untested
EVIDENCE: At 1a22ab9e6c8e5d26af16fcc701f9ed93111ccab8, KCL filesystem, OCI-package and target tests inspect parsed scenarios without generating resulting documentation. The base matrix formats documentation, but it does not cover the other source forms and targets through that path.

CRITERION: AC-09 CUE, Python and a custom function unknown to source adapters have passing real-render-to-documentation tests with successful and failing scenarios.
VERDICT: untested
EVIDENCE: At 1a22ab9e6c8e5d26af16fcc701f9ed93111ccab8, TestUnknownFunctionInput exercises a successful Python runtime with an unknown input API and verifies source classification. It does not format resulting documentation or exercise a failing custom-function scenario. CUE and Python error tests call Render directly without checking CLI failure/output preservation.

CRITERION: AC-10 Mixed-pipeline create/update examples prove shared context, conditional/repeated resources, mocked observed state and readiness/status behavior through real rendering and resulting documentation.
VERDICT: untested
EVIDENCE: At 1a22ab9e6c8e5d26af16fcc701f9ed93111ccab8, TestMixedPipelineAndObservedState asserts resource counts, environment propagation and observedRegion. It does not generate documentation or assert readiness behavior and the other expected status values.

CRITERION: AC-11 Markdown, offline HTML, AsciiDoc, JSON, YAML and custom templates render the same shared model correctly, with format-appropriate escaping and a tested structured-data contract.
VERDICT: not verified
EVIDENCE: At 1a22ab9e6c8e5d26af16fcc701f9ed93111ccab8, fresh formatter tests for all five formats and JSON/YAML round trips passed. Independent execution of bun test --cwd tests/format failed before the tests because Chromium could not launch in the restricted environment: sandbox_host_linux.cc reported "shutdown: Operation not permitted". The offline-browser and Asciidoctor checks were not established by that execution.

CRITERION: AC-12 Configuration controls sections, columns, layout, sorting, nesting, descriptions, anchors, styling, headers/footers and output; explicit flags win and invalid configuration fails.
VERDICT: met
EVIDENCE: At 1a22ab9e6c8e5d26af16fcc701f9ed93111ccab8, fresh config and CLI tests passed, including explicit false overrides and unknown-key rejection. Inspection confirmed configuration-owned path resolution, section/column selection, layout, sorting, nesting, descriptions, anchors, styles and header/footer handling.

CRITERION: AC-13 File/directory/bundle generation supports deterministic discovery and resource matching, catches collisions and malformed supported inputs, and retains existing single-resource commands.
VERDICT: met
EVIDENCE: At 1a22ab9e6c8e5d26af16fcc701f9ed93111ccab8, fresh discovery, duplicate-identity, command compatibility and document-decoding tests passed. Inspection confirmed sorted file traversal, bundle ordering, API-identity matching of example XRs, symlink exclusion and strict single-resource kind checks.

CRITERION: AC-14 Replace and marker-injection output preserve unrelated content and resist partial writes; repeated injection is identical and check mode distinguishes current, drift and failure without writing.
VERDICT: not met
EVIDENCE: At 1a22ab9e6c8e5d26af16fcc701f9ed93111ccab8, fresh injection and output-safety tests passed. A built CLI returned exit 2 for a missing input with --check and --check=true, but returned exit 1 for the same invalid input with accepted boolean forms --check=TRUE and --check=1.

CRITERION: AC-15 Every advertised capability maps to a tested runnable example, and mandatory CI executes the full renderer matrix and formatter checks before any release tag can be created.
VERDICT: untested
EVIDENCE: At 1a22ab9e6c8e5d26af16fcc701f9ed93111ccab8, TestVersioningRequiresRealChecks passed and CI inspection confirmed mandatory formatter/render commands before the version job. The per-variant documentation, custom failure and readiness coverage gaps identified for AC-07 through AC-10 prevent establishing the advertised-capability mapping. The complete integration execution remained pending during this review.

CRITERION: AC-16 A local six-target release snapshot has verified archive contents, checksums and version metadata; local checksum signing verifies successfully, and release workflow contract checks require the exact checked tag and protect existing assets.
VERDICT: not verified
EVIDENCE: At 1a22ab9e6c8e5d26af16fcc701f9ed93111ccab8, independently inspected six snapshot archives, verified their SHA-256 checksums and confirmed README, LICENSE and executable contents. The Linux binary reported the exact target commit and snapshot version. gpg --verify reported a good signature from the temporary snapshot-test key. Local release configuration prevents replacing existing artifacts, and CI passes the version output tag to the publisher. The pinned external shared workflow contract could not be independently retrieved in this review.

CRITERION: AC-17 Optional upstream schema/CEL validation accepts valid supplied/rendered examples, rejects invalid fields, constraints and CEL rules, fails for missing required schemas, and includes real validator tests in mandatory CI.
VERDICT: not verified
EVIDENCE: At 1a22ab9e6c8e5d26af16fcc701f9ed93111ccab8, inspected TestRealSchemaValidation, which covers supplied and rendered valid resources plus field-type, CEL, pattern, unknown-field and missing-schema failures. CI invokes the integration suite without optional skips. That real-validator execution had not completed during this review.

FINDING: F-01
KIND: factual
SEVERITY: blocking
WHERE: internal/engine/composition.go:62
EVIDENCE: A custom.example.org/v1 Input containing resources: [alpha, beta] exits with "step custom: resource entry must be an object". The resourcesFrom condition applies patch-and-transform semantics to every unknown input with a resources key.
PROPOSAL: Parse resource templates only after positively identifying the patch-and-transform input contract. Preserve unknown inputs unchanged for reference generation and generic rendering, and add a regression test with a differently shaped resources field.

FINDING: F-02
KIND: factual
SEVERITY: blocking
WHERE: tests/render/render_test.go and tests/render/oci_test.go
EVIDENCE: Source variants, KCL targets, mixed scenarios and the unknown-adapter case mostly stop after parsed-scenario assertions. They do not run the required resulting-documentation checks; the custom-adapter case has no failure scenario, and readiness is not asserted in the mixed test.
PROPOSAL: Route every advertised scenario variant through documentation generation and assert its expected inventory and distinguishing values. Add custom failure, consumed required-resource and readiness assertions.

FINDING: F-03
KIND: factual
SEVERITY: should-fix
WHERE: internal/engine/model.go Scenario and internal/engine/format.go scenario output
EVIDENCE: The design requires scenario identity and inputs used. Scenario records retain only the name and XR path, and the built-in formatter does not display the XR path; observed resources, required resources, supplied schemas and context input provenance are absent.
PROPOSAL: Record and display the scenario's supplied inputs with credential-safe values or references. Test that two scenarios with different mock inputs remain distinguishable in generated documentation.

FINDING: F-04
KIND: factual
SEVERITY: blocking
WHERE: cmd/root.go:39
EVIDENCE: pflag accepts --check=TRUE and --check=1, but Execute detects check mode only by scanning for --check or --check=true. A missing input therefore returns drift code 1 for these accepted forms instead of failure code 2.
PROPOSAL: Determine check mode from the parsed command flag or use a typed exit classification. Add binary-level tests for accepted boolean forms and invalid input.

### R-4

CRITERION: AC-01 Users can document v1/v2 XRD metadata and select a named version or all versions; an unknown version fails explicitly.
VERDICT: met
EVIDENCE: At 7749946c817914edd7d8ee687b8363b13143c22f, fresh race tests for cmd, generator and engine passed. Version-selection tests cover named, all and unknown versions; CLI JSON inspection confirmed version metadata and scope.

CRITERION: AC-02 Nested objects, array items, maps and referenced schemas produce deterministic paths, correct parent-relative required markers and lossless defaults; missing references fail and recursive schemas terminate with an explicit reference marker.
VERDICT: met
EVIDENCE: At 7749946c817914edd7d8ee687b8363b13143c22f, fresh reference, file-reference, default and generator tests passed. Inspected sorted traversal, parent-relative required lookup and cycle markers; generated JSON includes nested array and map paths.

CRITERION: AC-03 Documentation includes numeric/string/list constraints, enums, formats, Kubernetes list/map metadata, CEL rules/messages and schema alternatives without claiming unexecuted validation.
VERDICT: met
EVIDENCE: At 7749946c817914edd7d8ee687b8363b13143c22f, fresh schema/formatter tests passed. Inspected preservation and formatting of numeric, string, list, Kubernetes, CEL and alternative-schema metadata; generation does not mark scenarios validated by default.

CRITERION: AC-04 Composition references expand patch sets and document direction, combine inputs, readable transform chains, policies, readiness, connection details and environment inputs; invalid patch-set references fail.
VERDICT: met
EVIDENCE: At 7749946c817914edd7d8ee687b8363b13143c22f, fresh composition tests passed. Generated patch-and-transform JSON expands patch sets and retains direction, combine, transform chains, policies, readiness and connection details; inspected environment formatting and reference-error handling.

CRITERION: AC-05 Every pipeline step and source is documented, including unknown function inputs, without presenting an incomplete declared inventory as complete runtime output.
VERDICT: met
EVIDENCE: At 7749946c817914edd7d8ee687b8363b13143c22f, fresh unknown-input regression passed. Independently ran the candidate binary with custom resources: [alpha, beta]; generation succeeded and retained the opaque input. Inspected ordered steps and declared-inventory provenance.

CRITERION: AC-06 Explicit rendering and saved rendered YAML produce scenario-labelled resource/status documentation; observed/required resources, schemas and context reach the real renderer, while failures/timeouts preserve existing output.
VERDICT: met
EVIDENCE: At 7749946c817914edd7d8ee687b8363b13143c22f, observed complete passing real-render log, including distributed CLI, mock inputs, required-resource consumption and required-schema consumption. Fresh saved-output/failure tests passed; inspected timeout handling and generation before file replacement.

CRITERION: AC-07 Every advertised Go-templating source and behavior in the scenario matrix has a runnable example and passing real-render-to-documentation test, including failure cases.
VERDICT: met
EVIDENCE: At 7749946c817914edd7d8ee687b8363b13143c22f, observed passing real-render tests for inline/template-list, environment, filesystem, condition, loop, required-resource, status and invalid-program cases. The common render helper checks scenario inventory and values through all five formatters and exact JSON scenario round trips.

CRITERION: AC-08 Every advertised KCL source and target has a runnable example and passing real-render-to-documentation test, including failure cases.
VERDICT: met
EVIDENCE: At 7749946c817914edd7d8ee687b8363b13143c22f, observed passing inline, filesystem and OCI-package render tests, all five KCL targets, mock/custom parameters and invalid-program cases. Inspected the common render-to-documentation assertions used by every successful variant.

CRITERION: AC-09 CUE, Python and a custom function unknown to source adapters have passing real-render-to-documentation tests with successful and failing scenarios.
VERDICT: met
EVIDENCE: At 7749946c817914edd7d8ee687b8363b13143c22f, observed passing CUE, Python and unknown-adapter success cases, plus invalid CUE/Python/custom programs. Successful cases run all five documentation formatters; failing cases execute the CLI and verify existing output remains intact.

CRITERION: AC-10 Mixed-pipeline create/update examples prove shared context, conditional/repeated resources, mocked observed state and readiness/status behavior through real rendering and resulting documentation.
VERDICT: met
EVIDENCE: At 7749946c817914edd7d8ee687b8363b13143c22f, observed passing mixed create/update tests. Assertions verify shared environment values, changing resource inventories, observed-region patches, Pending/Ready phase, resourceCount and Ready conditions; both scenarios also pass the common documentation assertions.

CRITERION: AC-11 Markdown, offline HTML, AsciiDoc, JSON, YAML and custom templates render the same shared model correctly, with format-appropriate escaping and a tested structured-data contract.
VERDICT: met
EVIDENCE: At 7749946c817914edd7d8ee687b8363b13143c22f, fresh five-format and structured round-trip tests passed. Inspected completed formatter log: offline desktop/mobile HTML, Asciidoctor conversion and custom HTML escaping all passed, with 16 assertions. Production HTML has embedded styles and accessible table headers.

CRITERION: AC-12 Configuration controls sections, columns, layout, sorting, nesting, descriptions, anchors, styling, headers/footers and output; explicit flags win and invalid configuration fails.
VERDICT: met
EVIDENCE: At 7749946c817914edd7d8ee687b8363b13143c22f, fresh config and CLI tests passed, including explicit false overrides, path handling and unknown-key rejection. Inspected section/column, layout, sorting, nesting, descriptions, anchors, styles and header/footer handling.

CRITERION: AC-13 File/directory/bundle generation supports deterministic discovery and resource matching, catches collisions and malformed supported inputs, and retains existing single-resource commands.
VERDICT: met
EVIDENCE: At 7749946c817914edd7d8ee687b8363b13143c22f, fresh discovery, duplicate-identity, document and command tests passed. Inspected stable discovery/document order, API identity matching, symlink exclusion and strict single-resource commands; final fixtures are semantically equal to the previously tested formatting.

CRITERION: AC-14 Replace and marker-injection output preserve unrelated content and resist partial writes; repeated injection is identical and check mode distinguishes current, drift and failure without writing.
VERDICT: met
EVIDENCE: At 7749946c817914edd7d8ee687b8363b13143c22f, fresh injection and output-safety tests passed. Independently executed the final candidate binary with --check, --check=true, --check=TRUE and --check=1 against missing input; each returned 2. Inspected temporary-file replacement and byte-preserving marker handling.

CRITERION: AC-15 Every advertised capability maps to a tested runnable example, and mandatory CI executes the full renderer matrix and formatter checks before any release tag can be created.
VERDICT: met
EVIDENCE: At 7749946c817914edd7d8ee687b8363b13143c22f, observed complete mandatory real-render matrix PASS, including source variants, targets, validation, invalid programs and versioning-contract test. Inspected CI: formatter and renderer commands must succeed in verify before version runs. Independently fetched the pinned shared version workflow and confirmed tagging uses the checked commit.

CRITERION: AC-16 A local six-target release snapshot has verified archive contents, checksums and version metadata; local checksum signing verifies successfully, and release workflow contract checks require the exact checked tag and protect existing assets.
VERDICT: met
EVIDENCE: At 7749946c817914edd7d8ee687b8363b13143c22f, independently verified six final snapshot archives, SHA-256 checksums, LICENSE/README/executable contents and candidate version metadata. GPG reported a good temporary-key checksum signature. Independently fetched pinned shared workflows and inspected exact-tag checkout, chained version output, signing and local asset-immutability settings. No live publication was performed.

CRITERION: AC-17 Optional upstream schema/CEL validation accepts valid supplied/rendered examples, rejects invalid fields, constraints and CEL rules, fails for missing required schemas, and includes real validator tests in mandatory CI.
VERDICT: met
EVIDENCE: At 7749946c817914edd7d8ee687b8363b13143c22f, observed TestRealSchemaValidation passing for supplied and rendered valid resources and invalid field type, CEL, pattern, unknown field and missing-schema cases. Inspected mandatory CI invocation and validator error propagation before documentation output.

### R-5

CRITERION: AC-01 Users can document v1/v2 XRD metadata and select a named version or all versions; an unknown version fails explicitly.
VERDICT: met
EVIDENCE: At 95124b34ce929b84464f4de9548c922f5b08aef2, fresh command, generator and engine race tests passed. Inspected v1/v2 metadata extraction, named/all-version selection and explicit unknown-version rejection.

CRITERION: AC-02 Nested objects, array items, maps and referenced schemas produce deterministic paths, correct parent-relative required markers and lossless defaults; missing references fail and recursive schemas terminate with an explicit reference marker.
VERDICT: met
EVIDENCE: Fresh reference, explicit-file-reference, defaults and schema tests passed at the target. Inspected sorted traversal, parent-relative required markers, nested paths, reference allowlisting, cycle markers and recursion bounds.

CRITERION: AC-03 Documentation includes numeric/string/list constraints, enums, formats, Kubernetes list/map metadata, CEL rules/messages and schema alternatives without claiming unexecuted validation.
VERDICT: met
EVIDENCE: Fresh schema and formatter tests passed at the target, including root constraints and container metadata. Inspected preservation of constraint families, CEL messages and schema alternatives; validation status is set only after successful explicit validation.

CRITERION: AC-04 Composition references expand patch sets and document direction, combine inputs, readable transform chains, policies, readiness, connection details and environment inputs; invalid patch-set references fail.
VERDICT: met
EVIDENCE: Fresh composition and engine race tests passed. Inspected patch-set expansion and missing/cyclic-reference rejection, patch direction, combine variables, transform summaries, policies, readiness, connection details and environment output. Observed the real patch-and-transform scenario passing.

CRITERION: AC-05 Every pipeline step and source is documented, including unknown function inputs, without presenting an incomplete declared inventory as complete runtime output.
VERDICT: met
EVIDENCE: Fresh unknown-input and source-rendering tests passed. Inspected ordered steps, opaque custom input preservation and declared-resource provenance. Observed the unknown-adapter integration case passing; document output explicitly distinguishes declared definitions from scenario results.

CRITERION: AC-06 Explicit rendering and saved rendered YAML produce scenario-labelled resource/status documentation; observed/required resources, schemas and context reach the real renderer, while failures/timeouts preserve existing output.
VERDICT: met
EVIDENCE: Fresh saved-output, input-provenance, renderer failure/cancellation and output-safety tests passed. Inspected argument forwarding, bounded execution and generation before replacement. Observed real mock-resource, context and required-schema consumption tests passing across the completed integration runs.

CRITERION: AC-07 Every advertised Go-templating source and behavior in the scenario matrix has a runnable example and passing real-render-to-documentation test, including failure cases.
VERDICT: met
EVIDENCE: Inspected fixtures and common documentation assertions. Observed passing inline, template-list, environment, filesystem, condition, loop, required-resource and status scenarios, plus invalid-program rejection. Successful variants assert identities and values through all five formatters and exact JSON scenario round trips.

CRITERION: AC-08 Every advertised KCL source and target has a runnable example and passing real-render-to-documentation test, including failure cases.
VERDICT: met
EVIDENCE: Observed passing inline, filesystem and local OCI-package cases, all five KCL targets, mock/parameter consumption and invalid programs. Inspected their common render-to-documentation assertions. Address-pool failures in the initial run were covered by the successful dedicated-network recovery run.

CRITERION: AC-09 CUE, Python and a custom function unknown to source adapters have passing real-render-to-documentation tests with successful and failing scenarios.
VERDICT: met
EVIDENCE: Observed passing CUE, Python and unknown-adapter success cases and CUE/Python/custom failure cases. Inspected five-format resource assertions and CLI failure checks that preserve an existing output file.

CRITERION: AC-10 Mixed-pipeline create/update examples prove shared context, conditional/repeated resources, mocked observed state and readiness/status behavior through real rendering and resulting documentation.
VERDICT: met
EVIDENCE: Observed both mixed-pipeline scenarios passing. Inspected assertions for environment propagation, changing resource inventories, observed-region patches, Pending/Ready phases, resourceCount and Ready conditions, followed by common documentation assertions.

CRITERION: AC-11 Markdown, offline HTML, AsciiDoc, JSON, YAML and custom templates render the same shared model correctly, with format-appropriate escaping and a tested structured-data contract.
VERDICT: met
EVIDENCE: Fresh shared-model and structured-output race tests passed. Independently reran offline desktop/mobile HTML, Asciidoctor conversion and custom HTML escaping: three tests and 16 assertions passed with a 30-second timeout after an initial Chromium page-creation timeout. Inspected the completed 18-case formatter log and confirmed relevant implementation files are unchanged from its tested source.

CRITERION: AC-12 Configuration controls sections, columns, layout, sorting, nesting, descriptions, anchors, styling, headers/footers and output; explicit flags win and invalid configuration fails.
VERDICT: met
EVIDENCE: Fresh configuration and command tests passed, including explicit boolean overrides, relative paths and malformed configuration rejection. Inspected section/column selection, layouts, sorting, nesting, description and anchor controls, styles, headers/footers and output settings.

CRITERION: AC-13 File/directory/bundle generation supports deterministic discovery and resource matching, catches collisions and malformed supported inputs, and retains existing single-resource commands.
VERDICT: met
EVIDENCE: Fresh discovery, duplicate-identity, bundle, malformed-input and compatibility tests passed. Inspected stable traversal/document order, symlink exclusion and API-identity matching. Independently confirmed that a malformed Composition fails batch generation; its missing diagnostic location is reported separately as F-01.

CRITERION: AC-14 Replace and marker-injection output preserve unrelated content and resist partial writes; repeated injection is identical and check mode distinguishes current, drift and failure without writing.
VERDICT: met
EVIDENCE: Fresh injection, idempotence, output-preservation and empty-template tests passed. Inspected temporary-file replacement and marker validation. Observed binary integration checks passing for --check, --check=true, --check=TRUE and --check=1 with invalid-input exit code 2.

CRITERION: AC-15 Every advertised capability maps to a tested runnable example, and mandatory CI executes the full renderer matrix and formatter checks before any release tag can be created.
VERDICT: met
EVIDENCE: Independently reconciled all 14 declared integration tests against passing records across the original and recovery runs; the recovery run had no failures or skips. Inspected runnable fixtures, mandatory formatter/render steps and version.needs=verify. Fresh workflow-contract tests passed, and the pinned shared version workflow was retrieved and inspected. This is source and local execution evidence, not a hosted CI claim.

CRITERION: AC-16 A local six-target release snapshot has verified archive contents, checksums and version metadata; local checksum signing verifies successfully, and release workflow contract checks require the exact checked tag and protect existing assets.
VERDICT: met
EVIDENCE: Independently verified all six fresh snapshot archives, their SHA-256 checksums, executable/README/LICENSE contents and exact target commit metadata. The native binary reports 0.1.5-SNAPSHOT-95124b3; GPG returned VALIDSIG for the receipt's ephemeral key. Confirmed the snapshot configuration differs only by its scratch dist path. Freshly retrieved pinned version/release workflows and publisher sources establish checked-commit tagging, exact-tag checkout and signing; committed GoReleaser configuration disables artifact replacement. Live release-version selection and publication are separate shipping checks.

CRITERION: AC-17 Optional upstream schema/CEL validation accepts valid supplied/rendered examples, rejects invalid fields, constraints and CEL rules, fails for missing required schemas, and includes real validator tests in mandatory CI.
VERDICT: met
EVIDENCE: Observed real validation passing for valid supplied/rendered resources and rejecting field-type, CEL, pattern, unknown-field and missing-schema cases. Inspected validator metadata filtering, error-on-missing-schemas, error propagation before output and mandatory CI inclusion.

FINDING: F-01
KIND: factual
SEVERITY: should-fix
WHERE: internal/engine/load.go:158 and internal/engine/load.go:164; documentation-engine.spec.md Design, Output and configuration
EVIDENCE: The design requires malformed supported inputs to be reported with their locations. At 95124b34ce929b84464f4de9548c922f5b08aef2, the fresh snapshot CLI generated a directory containing a valid XRD and b-broken-composition.yaml with spec: {}. It exited 1 with only “Composition requires compositeTypeRef apiVersion and kind”, omitting the filename and document number. Build returns extraction errors without adding the available Input.File and Input.Index.
PROPOSAL: Wrap XRD and Composition extraction errors with their source filename and document index. Add a batch CLI regression that verifies semantic manifest errors identify the offending file and document.

### R-6

CRITERION: AC-01 Users can document v1/v2 XRD metadata and select a named version or all versions; an unknown version fails explicitly.
VERDICT: met
EVIDENCE: At 4aa59e22b11a7c09f640a55c369a49483ae484b4, inspected unchanged version-selection logic and metadata extraction. The fresh full check log passes command, generator and engine tests; independently reran the new XRD semantic-error cases successfully.

CRITERION: AC-02 Nested objects, array items, maps and referenced schemas produce deterministic paths, correct parent-relative required markers and lossless defaults; missing references fail and recursive schemas terminate with an explicit reference marker.
VERDICT: met
EVIDENCE: At the target, independently reran reference, explicit-file-reference and lossless-default tests with race detection. All passed. Reviewed the error-only schema/load changes; deterministic paths, parent-relative requirements, reference cycles and successful schema extraction are preserved.

CRITERION: AC-03 Documentation includes numeric/string/list constraints, enums, formats, Kubernetes list/map metadata, CEL rules/messages and schema alternatives without claiming unexecuted validation.
VERDICT: met
EVIDENCE: Inspected the unchanged constraint collection and formatting paths and the fresh passing engine test log. Numeric/string/list constraints, enums, formats, Kubernetes metadata, CEL messages and schema alternatives remain present; validation status still requires explicit successful validation.

CRITERION: AC-04 Composition references expand patch sets and document direction, combine inputs, readable transform chains, policies, readiness, connection details and environment inputs; invalid patch-set references fail.
VERDICT: met
EVIDENCE: Composition extraction and formatting are unchanged from the prior reviewed source. Observed the fresh full composition/engine checks passing and the completed real patch-and-transform scenario. Patch-set errors, direction, combine inputs, transforms, policies, readiness, connections and environment output remain covered.

CRITERION: AC-05 Every pipeline step and source is documented, including unknown function inputs, without presenting an incomplete declared inventory as complete runtime output.
VERDICT: met
EVIDENCE: Inspected unchanged ordered pipeline extraction, opaque custom inputs and declared-inventory provenance. The fresh full checks and observed real unknown-adapter case pass; the source/runtime distinction remains explicit in generated documents.

CRITERION: AC-06 Explicit rendering and saved rendered YAML produce scenario-labelled resource/status documentation; observed/required resources, schemas and context reach the real renderer, while failures/timeouts preserve existing output.
VERDICT: met
EVIDENCE: Renderer, validator, scenario provenance and output-success paths are unchanged. Observed passing real saved/live scenario, mock-resource, required-schema and context coverage, plus fresh failure/cancellation tests in the full check log. The new error wrappers return before output replacement.

CRITERION: AC-07 Every advertised Go-templating source and behavior in the scenario matrix has a runnable example and passing real-render-to-documentation test, including failure cases.
VERDICT: met
EVIDENCE: Reconciled the observed passing Go-template inline, template-list, environment, filesystem, condition, loop, required-resource, status and invalid-program cases with unchanged fixtures and integration assertions. Successful variants retain five-format identity/value assertions and exact JSON scenario round trips.

CRITERION: AC-08 Every advertised KCL source and target has a runnable example and passing real-render-to-documentation test, including failure cases.
VERDICT: met
EVIDENCE: Observed passing inline, filesystem, OCI-package, all five KCL targets, mock/parameter and invalid-program coverage. Fixtures and render-to-documentation assertions are unchanged. The earlier Docker address-pool failures were covered by the successful dedicated-network recovery run.

CRITERION: AC-09 CUE, Python and a custom function unknown to source adapters have passing real-render-to-documentation tests with successful and failing scenarios.
VERDICT: met
EVIDENCE: Observed passing CUE, Python and unknown-adapter successes and CUE/Python/custom failures. Their unchanged tests verify resulting documentation through all five formats and preserve existing files when the CLI render fails.

CRITERION: AC-10 Mixed-pipeline create/update examples prove shared context, conditional/repeated resources, mocked observed state and readiness/status behavior through real rendering and resulting documentation.
VERDICT: met
EVIDENCE: Observed both mixed create/update cases passing with unchanged fixtures and assertions for shared environment, conditional/repeated inventories, observed-region patches, Pending/Ready phases, resourceCount and Ready conditions in rendered documentation.

CRITERION: AC-11 Markdown, offline HTML, AsciiDoc, JSON, YAML and custom templates render the same shared model correctly, with format-appropriate escaping and a tested structured-data contract.
VERDICT: met
EVIDENCE: Observed the final 18-case formatter run passing all 336 assertions. Independently reran the changed reload case successfully with five assertions. Earlier independently executed offline desktop/mobile HTML, Asciidoctor and custom-template escaping checks remain applicable because formatter/model output code is unchanged.

CRITERION: AC-12 Configuration controls sections, columns, layout, sorting, nesting, descriptions, anchors, styling, headers/footers and output; explicit flags win and invalid configuration fails.
VERDICT: met
EVIDENCE: Configuration and presentation code are unchanged. Inspected the final full passing command/engine check log and independently reran configuration/injection checks. Existing coverage establishes sections, columns, layouts, sorting, nesting, descriptions, anchors, styling, headers/footers, explicit overrides and rejection of invalid configuration.

CRITERION: AC-13 File/directory/bundle generation supports deterministic discovery and resource matching, catches collisions and malformed supported inputs, and retains existing single-resource commands.
VERDICT: met
EVIDENCE: Independently reran discovery and all four new semantic-error regression cases with race detection. They pass, identify the source file and document 2 exactly once, preserve the underlying cause and retain existing output. Stable discovery, API matching, duplicate handling and single-resource commands are unchanged. F-01 is closed.

CRITERION: AC-14 Replace and marker-injection output preserve unrelated content and resist partial writes; repeated injection is identical and check mode distinguishes current, drift and failure without writing.
VERDICT: met
EVIDENCE: Independently reran marker injection, output-safety and empty-template checks with race detection; all passed. The final full check log passes the remaining command tests. Previously observed binary checks establish current/drift/failure exits, including all accepted true boolean spellings; output implementation is unchanged.

CRITERION: AC-15 Every advertised capability maps to a tested runnable example, and mandatory CI executes the full renderer matrix and formatter checks before any release tag can be created.
VERDICT: met
EVIDENCE: All 14 declared real-render tests have observed passing coverage across the original and dedicated-network recovery logs. Render fixtures, runner and CI are unchanged. Inspected mandatory formatter/render verification before versioning and previously retrieved the exact pinned shared version workflow. No hosted CI result is inferred.

CRITERION: AC-16 A local six-target release snapshot has verified archive contents, checksums and version metadata; local checksum signing verifies successfully, and release workflow contract checks require the exact checked tag and protect existing assets.
VERDICT: met
EVIDENCE: Independently verified all six fresh 4aa59e2 snapshot archives, SHA-256 checksums, executable/README/LICENSE contents and exact target commit metadata. The native binary reports 0.1.5-SNAPSHOT-4aa59e2, and GPG returned VALIDSIG for the receipt fingerprint. Snapshot configuration differs only by its scratch dist path. Previously retrieved exact-tag/signing contracts and committed artifact-immutability settings are unchanged; live v0.2.0 version selection and publication remain shipping checks.

CRITERION: AC-17 Optional upstream schema/CEL validation accepts valid supplied/rendered examples, rejects invalid fields, constraints and CEL rules, fails for missing required schemas, and includes real validator tests in mandatory CI.
VERDICT: met
EVIDENCE: Observed real validation accepting supplied/rendered valid resources and rejecting field-type, CEL, pattern, unknown-field and missing-schema cases. Validator implementation, local schemas and mandatory CI invocation are unchanged; the diagnostic repair preserves validation failure propagation before output.


## Sources

- [Crossplane CLI reference](https://docs.crossplane.io/cli/v2.5/command-reference/): rendering interface, mock inputs, runtime controls and resource/schema/CEL validation.
- [Crossplane CLI v2.4 reference](https://docs.crossplane.io/cli/v2.4/command-reference/): version-specific command contract and renderer inputs.
- [Crossplane compositions](https://docs.crossplane.io/latest/composition/compositions/): function pipelines and required schemas.
- [Legacy composition rendering](https://docs.crossplane.io/master/concepts/compositions/): legacy render command and resource-mode limitation.
- [Go-templating function](https://github.com/crossplane-contrib/function-go-templating): source forms, dynamic inputs and helper behavior.
- [KCL function](https://github.com/crossplane-contrib/function-kcl): source forms, targets and filesystem runtime requirements.
- [CUE function](https://github.com/crossplane-contrib/function-cue): public function implementation for the CUE test scenario.
- [Python function](https://github.com/crossplane-contrib/function-python): public function implementation for the Python test scenario.
- [Crossplane v2.2 composition examples](https://github.com/crossplane/docs/blob/master/content/v2.2/composition/compositions.md): mixed functions and Python dynamic resource requests.
- [terraform-docs configuration](https://terraform-docs.io/user-guide/configuration/): declarative configuration and discovery precedent.
- [terraform-docs formatters](https://terraform-docs.io/user-guide/configuration/formatter/), [reference](https://terraform-docs.io/reference/): output and formatter precedent; native HTML is a crossplane-docs requirement.

