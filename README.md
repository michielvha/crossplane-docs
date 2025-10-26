# crossplane-xrd-docs

A terraform-docs style documentation generator for Crossplane XRDs (CompositeResourceDefinitions).

## Features

- Parse Crossplane XRD YAML files
- Extract OpenAPI v3 schema information
- Generate formatted markdown documentation tables (terraform-docs style)

## Installation

```bash
go install github.com/michielvha/crossplane-xrd-docs@latest
```

Or build from source:

```bash
git clone https://github.com/michielvha/crossplane-xrd-docs.git
cd crossplane-xrd-docs
go build -o crossplane-xrd-docs .
```

## Usage

### Generate documentation to stdout

```bash
crossplane-xrd-docs generate xrd.yaml
```

### Generate documentation to a file

```bash
crossplane-xrd-docs generate xrd.yaml -o README.md

# Flatten nested structures (simpler view)
crossplane-xrd-docs generate xrd.yaml --show-nested=false
```

### Example Output

The tool generates markdown tables with:
- **Spec Fields**: All input parameters with required/optional status, defaults, constraints
- **Status Fields**: All output/status fields (always included!)
- Field names (with indentation for nested objects)
- Types (string, integer, list(string), object, etc.)
- Descriptions from your XRD
- Required/optional indicators (✅/❌)
- Default values
- Validation constraints (enums, min/max, minItems, etc.)

## Tech Stack

- **Language:** Go 1.24+
- **CLI Framework:** [Cobra](https://github.com/spf13/cobra)
- **YAML Parsing:** gopkg.in/yaml.v3
- **Kubernetes Types:** k8s.io/apimachinery

## Roadmap

- [x] Basic XRD parsing
- [x] Markdown table generation
- [x] CLI with Cobra
- [ ] Support for nested objects
- [ ] Status field documentation
- [ ] Multiple output formats (JSON, YAML)

## Acknowledgments

Inspired by [terraform-docs](https://github.com/terraform-docs/terraform-docs) and designed to integrate with the [Crossplane](https://crossplane.io) ecosystem.
