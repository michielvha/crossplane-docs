---
status: planned
status_description: "Independent spec review passed and owner authorized implementation."
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

- [ ] AC-01: Users can document v1/v2 XRD metadata and select a named version or all versions; an unknown version fails explicitly.
- [ ] AC-02: Nested objects, array items, maps and referenced schemas produce deterministic paths, correct parent-relative required markers and lossless defaults; missing references fail and recursive schemas terminate with an explicit reference marker.
- [ ] AC-03: Documentation includes numeric/string/list constraints, enums, formats, Kubernetes list/map metadata, CEL rules/messages and schema alternatives without claiming unexecuted validation.
- [ ] AC-04: Composition references expand patch sets and document direction, combine inputs, readable transform chains, policies, readiness, connection details and environment inputs; invalid patch-set references fail.
- [ ] AC-05: Every pipeline step and source is documented, including unknown function inputs, without presenting an incomplete declared inventory as complete runtime output.
- [ ] AC-06: Explicit rendering and saved rendered YAML produce scenario-labelled resource/status documentation; observed/required resources, schemas and context reach the real renderer, while failures/timeouts preserve existing output.
- [ ] AC-07: Every advertised Go-templating source and behavior in the scenario matrix has a runnable example and passing real-render-to-documentation test, including failure cases.
- [ ] AC-08: Every advertised KCL source and target has a runnable example and passing real-render-to-documentation test, including failure cases.
- [ ] AC-09: CUE, Python and a custom function unknown to source adapters have passing real-render-to-documentation tests with successful and failing scenarios.
- [ ] AC-10: Mixed-pipeline create/update examples prove shared context, conditional/repeated resources, mocked observed state and readiness/status behavior through real rendering and resulting documentation.
- [ ] AC-11: Markdown, offline HTML, AsciiDoc, JSON, YAML and custom templates render the same shared model correctly, with format-appropriate escaping and a tested structured-data contract.
- [ ] AC-12: Configuration controls sections, columns, layout, sorting, nesting, descriptions, anchors, styling, headers/footers and output; explicit flags win and invalid configuration fails.
- [ ] AC-13: File/directory/bundle generation supports deterministic discovery and resource matching, catches collisions and malformed supported inputs, and retains existing single-resource commands.
- [ ] AC-14: Replace and marker-injection output preserve unrelated content and resist partial writes; repeated injection is identical and check mode distinguishes current, drift and failure without writing.
- [ ] AC-15: Every advertised capability maps to a tested runnable example, and mandatory CI executes the full renderer matrix and formatter checks before any release tag can be created.
- [ ] AC-16: A local six-target release snapshot has verified archive contents, checksums and version metadata; local checksum signing verifies successfully, and release workflow contract checks require the exact checked tag and protect existing assets.
- [ ] AC-17: Optional upstream schema/CEL validation accepts valid supplied/rendered examples, rejects invalid fields, constraints and CEL rules, fails for missing required schemas, and includes real validator tests in mandatory CI.

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

- [ ] T1 `tests/render/cases.yaml`, `examples/` - resolve and pin actual runtime versions/variants and add complete shared scenarios; fixture validation succeeds (AC-07, AC-08, AC-09, AC-10, AC-15).
- [ ] T2 `pkg/generator/generator_test.go`, `cmd/root_test.go` - add failing version/schema/default/reference scenarios; failures identify the missing behavior (AC-01, AC-02, AC-03).
- [ ] T3 `pkg/generator/generator.go`, `internal/model/`, `cmd/generate.go` - implement shared schema extraction and version selection; T2 passes (AC-01, AC-02, AC-03).
- [ ] T4 `pkg/composition/composition_test.go`, `tests/render/render_test.go` - add failing patch/reference and pipeline scenario assertions (AC-04, AC-05, AC-06).
- [ ] T5 `pkg/composition/composition.go`, `internal/model/`, `internal/render/` - implement source references, patch expansion and the generic renderer; T4 passes (AC-04, AC-05, AC-06).
- [ ] T6 `tests/render/render_test.go`, `examples/` - execute the complete matrix with real runtimes and assert generated documentation; every required scenario passes without skips (AC-07, AC-08, AC-09, AC-10).
- [ ] T7 `internal/format/`, `internal/config/`, `cmd/root_test.go` - add failing tests for every formatter, customization and configuration rule (AC-11, AC-12).
- [ ] T8 `internal/format/`, `internal/config/`, `cmd/` - implement the formatters, configuration and template contracts; T7 passes (AC-11, AC-12).
- [ ] T9 `cmd/root_test.go`, `internal/document/` - add failing discovery, bundle, injection, output-preservation and drift tests (AC-13, AC-14).
- [ ] T10 `cmd/`, `internal/document/` - implement batch generation and output modes; T9 passes (AC-13, AC-14).
- [ ] T13 `tests/render/render_test.go`, `internal/render/` - add failing validator/CLI scenarios for valid, invalid and missing-schema inputs (AC-17).
- [ ] T14 `internal/render/`, `cmd/` - implement the explicit upstream validation mode; T13 passes against the real validator (AC-17).
- [ ] T11 `.github/workflows/ci.yml`, `scripts/`, `Makefile`, `renovate.json5` - require the real-render, validator and formatter suites before versioning; the intentional-failure CI check blocks tagging (AC-15, AC-17).
- [ ] T12 `.goreleaser.yml`, `README.md`, `examples/` - verify six-target snapshots, document only tested support and exact release behavior, and verify local signing and release workflow contracts (AC-15, AC-16).

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

### R-1

FINDING: F-01
KIND: factual
SEVERITY: blocking
WHERE: Test Plan AC-16 and Tasks T12
EVIDENCE: At d9a09ca, docs/specs/documentation-engine.spec.md:233 requires separate shipping authorization for live publication and remote tag changes, but line 271 requires an exact-tag signed-release smoke test and line 289 makes that check part of implementation completion. The authorized deliverable is a local feature branch ready for review, so these checks require an action outside that deliverable.
PROPOSAL: Make AC-16 completion depend on local snapshot, checksum, version, signing and workflow contract checks. Describe the live exact-tag publication smoke test as a shipping procedure that runs only after separate authorization, without making it a prerequisite for local implementation completion.

### R-2

NO FINDINGS

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
