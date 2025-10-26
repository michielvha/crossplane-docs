# crossplane-xrd-docs

This is a Go CLI tool that generates terraform-docs style markdown documentation from Crossplane XRD files.

## Project Goals
- Parse Crossplane XRD YAML files
- Extract OpenAPI v3 schema information
- Generate formatted markdown documentation tables
- Match the style and structure of terraform-docs
- Easy integration into the official crossplane CLI later

## Tech Stack
- Go 1.25+
- Cobra CLI framework (same as crossplane CLI)
- YAML parsing libraries
- Template-based markdown generation
