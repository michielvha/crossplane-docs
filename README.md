# crossplane-docs

Generate API references and rendered examples from Crossplane XRDs and
Compositions. Write Markdown, a standalone HTML reference, AsciiDoc, JSON or YAML;
keep the presentation and file-output settings in a configuration file.

![Generated XPlatform HTML reference with field requirements, defaults and a matching example manifest.](docs/assets/output-preview-light.svg#gh-light-mode-only)
![Generated XPlatform HTML reference with field requirements, defaults and a matching example manifest.](docs/assets/output-preview-dark.svg#gh-dark-mode-only)

## Install

```bash
go install github.com/michielvha/crossplane-docs@latest
```

Source builds require the Go version in [go.mod](go.mod). Run `make build` for a
local binary with version metadata. Published binaries live in
[GitHub Releases](https://github.com/michielvha/crossplane-docs/releases).
Release archives contain standalone Linux, macOS and Windows binaries for
amd64 and arm64.

## Generate a reference

```bash
# Inspect one XRD, including nested objects, arrays, maps and validations
crossplane-docs xrd examples/platform.yaml --all-versions

# Select one API version
crossplane-docs xrd examples/platform.yaml --api-version v1beta1

# Generate a complete reference from a directory or a multi-document YAML file
crossplane-docs generate examples --include platform.yaml --include 'platform/*.yaml' \
  --all-versions --format html -o platform.html

# Produce machine-readable documentation, rather than copying the input YAML
crossplane-docs generate examples/platform.yaml --all-versions --format json
crossplane-docs generate examples/platform.yaml --all-versions --format yaml

# A native AsciiDoc document
crossplane-docs xrd examples/platform.yaml --format asciidoc -o API.adoc
```

The reference documents scope, versions, claims, printer columns, spec/status
fields, parent-relative required markers, defaults, validation constraints, CEL
rules and schema alternatives. Local references are resolved; explicitly allow
external schema files with `--schema types.yaml`. Recursive references receive a
reference marker. Default values retain false, zero, empty strings and empty
collections.

Without a version flag, the first served XRD version is selected, falling back to
the first declared version when none is served. An unknown named version fails.
Use `--show-nested=false` for immediate properties in document formats. JSON and
YAML retain the complete nested model.

Directory generation discovers YAML in stable order. Use `--include`, `--exclude`
and `--recursive=false` to control discovery. Unrelated Kubernetes kinds are
skipped; malformed input and duplicate resource identities fail. The `xrd` and
`composition` commands each require exactly one document of the requested kind.

## Composition references and rendered examples

```bash
# Reference pipeline steps, sources, resource definitions, mappings and policies
crossplane-docs composition examples/platform/patch-and-transform.yaml

# Execute a pipeline with a sample XR, then document its actual output
crossplane-docs composition examples/platform/go-templating.yaml \
  --render --xr examples/platform/xr.yaml \
  --functions examples/platform/functions.yaml \
  --scenario team-atlas --format html -o composition.html

# Include mock observed state and required resources
crossplane-docs composition examples/platform/mixed.yaml \
  --render --xr examples/platform/xr.yaml \
  --functions examples/platform/functions.yaml \
  --observed-resources examples/platform/observed.yaml \
  --required-resources examples/platform/required.yaml \
  --scenario update --format markdown -o COMPOSITION.md

# Document output saved by an existing render pipeline
crossplane-docs composition examples/platform/go-templating.yaml \
  --rendered rendered.yaml --xr examples/platform/xr.yaml
```

Static references describe declared definitions, patch sets, combine inputs,
transform chains, policies, readiness, connection details, environment inputs and
function source settings. Named render scenarios show the resources produced for
that XR and its supplied inputs, including composite status, context and function
results. A different input can produce a different inventory. Secret contents and
credential fields are redacted in generated documentation. Scenario sections also
record the supplied XR, mock-file and schema references and safe context values
so readers can distinguish create and update inputs.

Rendering requires the Crossplane CLI and Docker. The tested CLI and engine pins
are in [the render matrix](tests/render/cases.yaml); the engine image and function
packages are pinned by digest. Set `--crossplane /path/to/crossplane` to use an
explicit CLI, `--crossplane-image` or `--crossplane-version` to choose another
engine, and `--timeout` to bound execution. Pipeline execution is opt-in and needs
no live Crossplane cluster, provider or cloud credentials for these examples.

Pass context with repeatable `--context-values KEY=JSON`, required resources with
`--required-resources`, and function-requested schemas with `--required-schemas`.
Only referenced Function runtimes are started. Rendering errors propagate;
existing output remains intact when generation fails.

The examples share one realistic [platform XRD](examples/platform.yaml) and
[sample XR](examples/platform/xr.yaml):

| Strategy | Runnable composition | Real integration coverage |
| --- | --- | --- |
| Application stack | [Composition](examples/platform/application-stack.yaml) | VPCs, subnets, compute replicas and PostgreSQL storage with branching and references |
| Patch and transform | [Composition](examples/platform/patch-and-transform.yaml) | Patch sets, combine, conversion, math, policies, readiness and connection details |
| Go templating | [Composition](examples/platform/go-templating.yaml) | Inline, template lists, image filesystem and environment sources; loops, conditions, required resources and status |
| KCL | [Composition](examples/platform/kcl.yaml) | Inline, image filesystem and local OCI package sources; all five targets |
| CUE | [Composition](examples/platform/cue.yaml) | Real resource generation and invalid programs |
| Python | [Composition](examples/platform/python.yaml) | Loops, resources, status and exceptions |
| Mixed pipeline | [Composition](examples/platform/mixed.yaml) | Environment configuration, patching, templating, observed state and readiness |

A function with an unknown input type can still use the generic render path;
its declared input remains visible in the reference. These are tested package
versions, not a claim that every future function version is compatible. The
[tests](tests/render/render_test.go) verify real rendered values and the resulting
documentation, including failures and a function input unknown to the adapters.

## Validate examples

```bash
crossplane-docs xrd examples/platform.yaml --all-versions \
  --validate --xr examples/platform/xr.yaml --extensions examples/platform.yaml
```

`--validate` uses Crossplane's schema and CEL validator. When combined with
`--render`, it validates the rendered resources before output. Supply the XRDs,
CRDs or extension packages through `--extensions`; missing schemas and invalid
examples fail. Reading and displaying a schema alone does not execute validation.

## Configure output

Start with the complete [configuration example](examples/platform/config.yml):

```bash
crossplane-docs generate examples --include platform.yaml --include 'platform/*.yaml' \
  --all-versions --config examples/platform/config.yml
```

By default, configuration discovery checks `.crossplane-docs.yml` in the input
root, then the invocation directory. Explicit flags override configuration,
including false values. Paths in a configuration file are relative to that file.
Unknown keys and invalid option combinations fail.

Select `html-layout: reference`, `explorer` or `compact` in the configuration, or
override it with `--html-layout`. The default is `reference`: field definitions
next to a matching example manifest. `explorer` provides a field browser and
inspector; `compact` uses comparison columns on desktop and stacked field
definitions on phones. All three are standalone files with embedded styles and
interactions, and remain readable with JavaScript disabled. HTML search and API
selection operate on the versions included by `--api-version` or `--all-versions`.
The explorer initially selects the first matching field in configured order;
the reference opens its first expandable field group. Its Examples inventory
links to paired manifests so each appears once. Examples without a paired API
reference remain inline.

Use Field requirement to show all, required or optional fields alongside text
search. Required means required within the field's parent object or schema
alternative. Filters stay in the URL across API/version changes and reloads;
opening a direct field link reveals it even if a filter would hide it.

The sun/moon toggle switches between the existing light palette and VH Midnight,
using Postbode's blue gradient with teal accents. The reference follows the system
theme until you choose a mode, then remembers that choice when browser storage
is available. Both themes work offline and with local files. Printing uses a
light palette and includes fields hidden by filters.

Embedded templates and programs use separate code blocks with real line breaks.
Copy controls preserve the program text, including intentional literal escapes.
Manifest and settings blocks use YAML rather than JSON-escaped source strings.

Configure section order and visibility, field columns, table/document text layouts,
sorting, nesting, descriptions, anchors, examples, HTML styles, indentation and
output location. Optional `template`, `header-from` and `footer-from` files
customize presentation. Custom templates receive the shared documentation model;
HTML templates escape manifest values. JSON and YAML retain the full versioned
model independently of presentation settings.

## Keep a README current

Add these markers to the target Markdown file:

```html
<!-- BEGIN_CROSSPLANE_DOCS -->
<!-- END_CROSSPLANE_DOCS -->
```

```bash
crossplane-docs xrd examples/platform.yaml --all-versions \
  -o API.md --output-mode inject

# CI drift check: no writes
crossplane-docs xrd examples/platform.yaml --all-versions \
  -o API.md --output-mode inject --check
```

Injection preserves surrounding content and rejects missing, duplicate or
reversed markers. It supports Markdown/HTML comments and AsciiDoc markers
`// BEGIN_CROSSPLANE_DOCS` and `// END_CROSSPLANE_DOCS`. Configurable `output.begin`
and `output.end` change the markers. JSON/YAML use stdout or replace-file output.
Check mode returns 0 for current documentation, 1 for drift and 2 for invalid
input, configuration or rendering. File writes use a temporary file and rename.

## Development and releases

```bash
make check        # Go tests with race detection, lint, vet, tidy and formatting
make integration  # Pinned CLI, real Docker function matrix and schema/CEL tests
make formats      # Offline Chromium checks and real Asciidoctor conversion
```

`make check` uses golangci-lint v2.15.0. Before `make formats`, run `bun install
--cwd tests/format` and, from that directory, `bunx playwright install chromium`.
`make integration` downloads and verifies the pinned Linux amd64 Crossplane CLI;
on another platform, set `CROSSPLANE_BIN` to the matching verified CLI. Fixture
images and the temporary OCI registry are built/started by the tests.
The [runtime Dockerfile](examples/platform/runtime/Dockerfile) builds composition
function fixtures with template/KCL source files for those tests.

The [CI workflow](.github/workflows/ci.yml) requires the Go, formatter and real
render suites before shared versioning. A new exact tag is passed to the shared
Go release workflow in the same run, using `GITHUB_TOKEN` and `PGP_PRIVATE_KEY`.
Linux/macOS/Windows amd64/arm64 archives receive checksums, a detached signature
and version metadata. Existing assets cannot be overwritten. Retry a failed
publication with **Re-run failed jobs**; manual tag pushes have no separate
publisher. The spec records the baseline-tag migration required before activation.

Contributor instructions live in [AGENTS.md](AGENTS.md); acceptance criteria and
implementation evidence live in the [v0.2.0 spec](docs/specs/documentation-engine.spec.md).
The small expected Markdown files in `examples/` are regression fixtures checked
by tests. Use `make examples` only for a deliberate, reviewed output change.

Inspired by [terraform-docs](https://github.com/terraform-docs/terraform-docs).
