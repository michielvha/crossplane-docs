package generator

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Options contains generation options
type Options struct {
	ShowNested bool // show nested object structures
}

// Generator handles documentation generation
type Generator struct{}

// New creates a new Generator instance
func New() *Generator {
	return &Generator{}
}

// XRD represents a simplified Crossplane CompositeResourceDefinition
type XRD struct {
	APIVersion string        `yaml:"apiVersion"`
	Kind       string        `yaml:"kind"`
	Metadata   metav1.ObjectMeta `yaml:"metadata"`
	Spec       XRDSpec       `yaml:"spec"`
}

// XRDSpec contains the XRD specification
type XRDSpec struct {
	Group      string       `yaml:"group"`
	Names      XRDNames     `yaml:"names"`
	ClaimNames *XRDNames    `yaml:"claimNames,omitempty"`
	Versions   []XRDVersion `yaml:"versions"`
}

// XRDNames contains the resource names
type XRDNames struct {
	Kind     string `yaml:"kind"`
	Plural   string `yaml:"plural"`
	Singular string `yaml:"singular,omitempty"`
}

// XRDVersion represents a version in the XRD
type XRDVersion struct {
	Name              string                 `yaml:"name"`
	Served            bool                   `yaml:"served"`
	Referenceable     bool                   `yaml:"referenceable"`
	Schema            XRDVersionSchema       `yaml:"schema"`
	AdditionalPrinterColumns []map[string]interface{} `yaml:"additionalPrinterColumns,omitempty"`
}

// XRDVersionSchema contains the OpenAPI schema
type XRDVersionSchema struct {
	OpenAPIV3Schema OpenAPISchema `yaml:"openAPIV3Schema"`
}

// OpenAPISchema represents the OpenAPI v3 schema
type OpenAPISchema struct {
	Type        string                   `yaml:"type"`
	Description string                   `yaml:"description,omitempty"`
	Properties  map[string]OpenAPISchema `yaml:"properties,omitempty"`
	Items       *OpenAPISchema           `yaml:"items,omitempty"`
	Required    []string                 `yaml:"required,omitempty"`
	Default     interface{}              `yaml:"default,omitempty"`
	Enum        []interface{}            `yaml:"enum,omitempty"`
	Minimum     *float64                 `yaml:"minimum,omitempty"`
	Maximum     *float64                 `yaml:"maximum,omitempty"`
	MinItems    *int                     `yaml:"minItems,omitempty"`
	MaxItems    *int                     `yaml:"maxItems,omitempty"`
	XKubernetesValidations []map[string]interface{} `yaml:"x-kubernetes-validations,omitempty"`
}

// Field represents a documented field
type Field struct {
	Name        string
	Type        string
	Description string
	Required    bool
	Default     string
	Constraints string
	Nested      []Field // For nested object fields
	Level       int     // Nesting level for display
}

// GenerateFromFile generates documentation from an XRD file
func (g *Generator) GenerateFromFile(filename string, opts Options) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	var xrd XRD
	if err := yaml.Unmarshal(data, &xrd); err != nil {
		return "", fmt.Errorf("failed to parse XRD YAML: %w", err)
	}

	return g.Generate(&xrd, opts)
}

// Generate generates documentation from an XRD struct
func (g *Generator) Generate(xrd *XRD, opts Options) (string, error) {
	if len(xrd.Spec.Versions) == 0 {
		return "", fmt.Errorf("no versions found in XRD")
	}

	// Use the first served version
	var version *XRDVersion
	for i := range xrd.Spec.Versions {
		if xrd.Spec.Versions[i].Served {
			version = &xrd.Spec.Versions[i]
			break
		}
	}
	if version == nil {
		version = &xrd.Spec.Versions[0]
	}

	// Extract spec fields
	specFields := g.extractFields(version.Schema.OpenAPIV3Schema, "spec", []string{}, 0, opts.ShowNested)
	
	// Always include status fields (they're part of the API!)
	statusFields := g.extractFields(version.Schema.OpenAPIV3Schema, "status", []string{}, 0, opts.ShowNested)

	// Generate markdown
	return g.generateMarkdown(xrd, version, specFields, statusFields)
}

// extractFields recursively extracts fields from the schema
func (g *Generator) extractFields(schema OpenAPISchema, prefix string, required []string, level int, showNested bool) []Field {
	var fields []Field

	if schema.Properties == nil {
		return fields
	}

	// Get the target properties (spec or status)
	targetProp, hasTarget := schema.Properties[prefix]
	if !hasTarget || targetProp.Properties == nil {
		return fields
	}

	for name, prop := range targetProp.Properties {
		field := Field{
			Name:        name,
			Type:        g.formatType(prop),
			Description: prop.Description,
			Required:    contains(targetProp.Required, name),
			Default:     g.formatDefault(prop.Default),
			Constraints: g.formatConstraints(prop),
			Level:       level,
		}

		// If this is an object and we want to show nested fields
		if showNested && prop.Type == "object" && prop.Properties != nil {
			field.Nested = g.extractNestedFields(prop, level+1, showNested)
		}

		fields = append(fields, field)
	}

	return fields
}

// extractNestedFields extracts nested object fields
func (g *Generator) extractNestedFields(schema OpenAPISchema, level int, showNested bool) []Field {
	var fields []Field

	if schema.Properties == nil {
		return fields
	}

	for name, prop := range schema.Properties {
		field := Field{
			Name:        name,
			Type:        g.formatType(prop),
			Description: prop.Description,
			Required:    contains(schema.Required, name),
			Default:     g.formatDefault(prop.Default),
			Constraints: g.formatConstraints(prop),
			Level:       level,
		}

		// Recursively extract if nested object
		if showNested && prop.Type == "object" && prop.Properties != nil {
			field.Nested = g.extractNestedFields(prop, level+1, showNested)
		}

		fields = append(fields, field)
	}

	return fields
}

// formatType formats the field type
func (g *Generator) formatType(schema OpenAPISchema) string {
	if schema.Type == "array" && schema.Items != nil {
		return fmt.Sprintf("list(%s)", g.formatType(*schema.Items))
	}
	if schema.Type == "object" {
		return "object"
	}
	if len(schema.Enum) > 0 {
		return "string"
	}
	return schema.Type
}

// formatDefault formats the default value
func (g *Generator) formatDefault(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprintf("%v", value)
}

// formatConstraints formats validation constraints
func (g *Generator) formatConstraints(schema OpenAPISchema) string {
	var constraints []string

	if len(schema.Enum) > 0 {
		enumVals := make([]string, len(schema.Enum))
		for i, v := range schema.Enum {
			enumVals[i] = fmt.Sprintf("`%v`", v)
		}
		constraints = append(constraints, fmt.Sprintf("Allowed: %s", strings.Join(enumVals, ", ")))
	}

	if schema.Minimum != nil {
		constraints = append(constraints, fmt.Sprintf("Min: %v", *schema.Minimum))
	}

	if schema.Maximum != nil {
		constraints = append(constraints, fmt.Sprintf("Max: %v", *schema.Maximum))
	}

	if schema.MinItems != nil {
		constraints = append(constraints, fmt.Sprintf("MinItems: %d", *schema.MinItems))
	}

	if schema.MaxItems != nil {
		constraints = append(constraints, fmt.Sprintf("MaxItems: %d", *schema.MaxItems))
	}

	return strings.Join(constraints, ", ")
}

// generateMarkdown generates the final markdown output
func (g *Generator) generateMarkdown(xrd *XRD, version *XRDVersion, specFields []Field, statusFields []Field) (string, error) {
	// Sort fields: required first, then alphabetically
	sort.Slice(specFields, func(i, j int) bool {
		if specFields[i].Required != specFields[j].Required {
			return specFields[i].Required
		}
		return specFields[i].Name < specFields[j].Name
	})

	if len(statusFields) > 0 {
		sort.Slice(statusFields, func(i, j int) bool {
			return statusFields[i].Name < statusFields[j].Name
		})
	}

	// Flatten nested fields for table display
	flatSpecFields := g.flattenFields(specFields)
	flatStatusFields := g.flattenFields(statusFields)

	tmpl := `# {{ .XRD.Spec.Names.Kind }}

{{ .Version.Schema.OpenAPIV3Schema.Description }}

**API Group:** {{ .XRD.Spec.Group }}  
**API Version:** {{ .Version.Name }}  
**Kind:** {{ .XRD.Spec.Names.Kind }}  
{{ if .XRD.Spec.ClaimNames }}**Claim Kind:** {{ .XRD.Spec.ClaimNames.Kind }}  {{ end }}

## Spec Fields

| Name | Type | Description | Required | Default | Constraints |
|------|------|-------------|----------|---------|-------------|
{{ range .SpecFields -}}
| {{ if gt .Level 0 }}{{ indent .Level }}{{ end }}{{ .Name }} | {{ .Type }} | {{ .Description }} | {{ if .Required }}✅{{ else }}❌{{ end }} | {{ if .Default }}` + "`{{ .Default }}`" + `{{ else }}-{{ end }} | {{ if .Constraints }}{{ .Constraints }}{{ else }}-{{ end }} |
{{ end }}
{{ if .StatusFields }}
## Status Fields

| Name | Type | Description |
|------|------|-------------|
{{ range .StatusFields -}}
| {{ if gt .Level 0 }}{{ indent .Level }}{{ end }}{{ .Name }} | {{ .Type }} | {{ .Description }} |
{{ end }}
{{ end }}
## Example

` + "```yaml" + `
apiVersion: {{ .XRD.Spec.Group }}/{{ .Version.Name }}
kind: {{ if .XRD.Spec.ClaimNames }}{{ .XRD.Spec.ClaimNames.Kind }}{{ else }}{{ .XRD.Spec.Names.Kind }}{{ end }}
metadata:
  name: example
spec:
  # Add your spec fields here
` + "```" + `
`

	funcMap := template.FuncMap{
		"indent": func(level int) string {
			return strings.Repeat("&nbsp;&nbsp;", level) + "↳ "
		},
	}

	t, err := template.New("markdown").Funcs(funcMap).Parse(tmpl)
	if err != nil {
		return "", err
	}

	data := struct {
		XRD          *XRD
		Version      *XRDVersion
		SpecFields   []Field
		StatusFields []Field
	}{
		XRD:          xrd,
		Version:      version,
		SpecFields:   flatSpecFields,
		StatusFields: flatStatusFields,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// flattenFields converts nested field structure to flat list for table display
func (g *Generator) flattenFields(fields []Field) []Field {
	var result []Field
	for _, field := range fields {
		result = append(result, field)
		if len(field.Nested) > 0 {
			result = append(result, g.flattenFields(field.Nested)...)
		}
	}
	return result
}

// Helper function
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
