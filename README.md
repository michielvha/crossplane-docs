# crossplane-docs

A terraform-docs style documentation generator for Crossplane resources.

## Features

- Parse Crossplane XRD and Composition YAML files
- Extract OpenAPI v3 schema information
- Generate formatted markdown documentation tables
- Show field mappings and transformations in compositions
- Support for nested objects with indentation
- Always documents both spec and status fields
- Show required/optional fields with visual indicators
- Display defaults, constraints, and validations

## Installation

```bash
go install github.com/michielvha/crossplane-docs@latest
```

Or build from source:

```bash
git clone https://github.com/michielvha/crossplane-docs.git
cd crossplane-docs
go build -o crossplane-docs .
```

## Usage

### XRD Documentation

Generate documentation for a CompositeResourceDefinition:

```bash
# Print to stdout
crossplane-docs xrd xrd.yaml

# Save to file
crossplane-docs xrd xrd.yaml -o README.md

# Flatten nested structures
crossplane-docs xrd xrd.yaml --show-nested=false
```

### Composition Documentation

Generate documentation for a Composition:

```bash
# Print to stdout  
crossplane-docs composition composition.yaml

# Save to file
crossplane-docs composition composition.yaml -o COMPOSITION.md

# Hide patch details
crossplane-docs composition composition.yaml --show-patches=false
```

## What It Generates

### XRD Documentation
- Spec fields table with types, descriptions, required/optional, defaults, constraints
- Status fields table
- Example YAML usage
- Nested object support with indentation

### Composition Documentation
- List of managed resources created
- Field mapping tables showing XRD field → managed resource field
- Transformation details (direct copy, string formatting, etc.)
- Resource inventory (what gets provisioned)

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

## Acknowledgments

Inspired by [terraform-docs](https://github.com/terraform-docs/terraform-docs) and designed to integrate with the [Crossplane](https://crossplane.io) ecosystem.
