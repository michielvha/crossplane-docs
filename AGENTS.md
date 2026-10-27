# Repository instructions

Read the personal conventions at
https://github.com/vhco-pro/docs/blob/c98ce2f926b7e0e1c29388be95cf7d8133a298e2/AGENTS.md
and follow the task routes there before implementation.

For code, read conventions/coding.md at that revision. For Go, also read
conventions/go.md. For CI and releases, read conventions/delivery.md and
knowledge/shared-delivery.md. For documents, read conventions/documentation.md
and guides/humanize-prose.md. For onboarding, read guides/onboard-a-repository.md.
For new features, read conventions/feature-documents.md before specification or
implementation. If a required source cannot be read, report it and stop.

Feature documents: docs/specs/

## Local commands and constraints

- `make build` builds the CLI with version metadata.
- `make test` runs regression tests with the race detector and coverage.
- `make check` also checks lint, vet, module tidiness and formatting. Use
  golangci-lint v2.15.0, matching the shared Go CI action.
- `actionlint .github/workflows/ci.yml` checks the workflow.
- `goreleaser check` checks release configuration with GoReleaser v2.18.3.
- `make examples` deliberately refreshes the expected Markdown. Review those
  diffs before accepting an output change.

Keep the public `pkg/generator` and `pkg/composition` APIs and the root Go entry
point while maintaining existing behavior. Fixtures and expected outputs live in
`examples/` and require no network services. Input must not be silently truncated
to the first document in a YAML bundle.
