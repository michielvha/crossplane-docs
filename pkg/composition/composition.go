package composition

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/template"

	"github.com/michielvha/crossplane-docs/internal/document"
	"gopkg.in/yaml.v3"
)

// Options contains generation options
type Options struct {
	ShowPatches bool // show patch details
}

// Generator handles composition documentation generation
type Generator struct{}

// New creates a new Generator instance
func New() *Generator {
	return &Generator{}
}

// Composition represents a Crossplane Composition
type Composition struct {
	APIVersion string                 `yaml:"apiVersion"`
	Kind       string                 `yaml:"kind"`
	Metadata   map[string]interface{} `yaml:"metadata"`
	Spec       CompositionSpec        `yaml:"spec"`
}

// CompositionSpec contains the composition specification
type CompositionSpec struct {
	CompositeTypeRef CompositeTypeRef `yaml:"compositeTypeRef"`
	Mode             string           `yaml:"mode,omitempty"`
	Resources        []Resource       `yaml:"resources,omitempty"`
	Pipeline         []PipelineStep   `yaml:"pipeline,omitempty"`
}

// CompositeTypeRef references the XR type
type CompositeTypeRef struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
}

// Resource represents a managed resource in the composition
type Resource struct {
	Name              string                 `yaml:"name"`
	Base              map[string]interface{} `yaml:"base"`
	Patches           []Patch                `yaml:"patches,omitempty"`
	ConnectionDetails []ConnectionDetail     `yaml:"connectionDetails,omitempty"`
}

// PipelineStep represents a function in the pipeline
type PipelineStep struct {
	Step        string                 `yaml:"step"`
	FunctionRef FunctionRef            `yaml:"functionRef"`
	Input       map[string]interface{} `yaml:"input,omitempty"`
}

// FunctionRef references a composition function
type FunctionRef struct {
	Name string `yaml:"name"`
}

// Patch represents a field patch
type Patch struct {
	Type          string                   `yaml:"type"`
	FromFieldPath string                   `yaml:"fromFieldPath,omitempty"`
	ToFieldPath   string                   `yaml:"toFieldPath,omitempty"`
	Combine       *Combine                 `yaml:"combine,omitempty"`
	Policy        map[string]interface{}   `yaml:"policy,omitempty"`
	Transforms    []map[string]interface{} `yaml:"transforms,omitempty"`
	PatchSetName  string                   `yaml:"patchSetName,omitempty"`
}

// Combine represents a field combination
type Combine struct {
	Variables []Variable `yaml:"variables"`
	Strategy  string     `yaml:"strategy"`
	String    *StringFmt `yaml:"string,omitempty"`
}

// Variable represents a combine variable
type Variable struct {
	FromFieldPath string `yaml:"fromFieldPath"`
}

// StringFmt represents string formatting
type StringFmt struct {
	Fmt string `yaml:"fmt"`
}

// ConnectionDetail represents a connection secret detail
type ConnectionDetail struct {
	Name                    string `yaml:"name"`
	Type                    string `yaml:"type,omitempty"`
	FromConnectionSecretKey string `yaml:"fromConnectionSecretKey,omitempty"`
	FromFieldPath           string `yaml:"fromFieldPath,omitempty"`
}

// ManagedResource represents a documented managed resource
type ManagedResource struct {
	Name        string
	Kind        string
	APIVersion  string
	Description string
	Patches     []PatchInfo
}

// PatchInfo represents patch information
type PatchInfo struct {
	Type           string
	XRDField       string
	MappedTo       string
	Transformation string
}

// GenerateFromFile generates documentation from a composition file
func (g *Generator) GenerateFromFile(filename string, opts Options) (string, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	var comp Composition
	if err := document.Decode(data, &comp); err != nil {
		return "", fmt.Errorf("failed to parse Composition YAML: %w", err)
	}

	return g.Generate(&comp, opts)
}

// Generate generates documentation from a Composition struct
func (g *Generator) Generate(comp *Composition, opts Options) (string, error) {
	if comp == nil {
		return "", fmt.Errorf("Composition must not be nil")
	}
	if comp.Kind != "Composition" || !strings.HasPrefix(comp.APIVersion, "apiextensions.crossplane.io/") {
		return "", fmt.Errorf("expected an apiextensions.crossplane.io Composition")
	}
	if comp.Spec.CompositeTypeRef.APIVersion == "" || comp.Spec.CompositeTypeRef.Kind == "" {
		return "", fmt.Errorf("Composition requires spec.compositeTypeRef.apiVersion and kind")
	}
	if comp.Spec.Mode != "" && comp.Spec.Mode != "Resources" && comp.Spec.Mode != "Pipeline" {
		return "", fmt.Errorf("unsupported Composition mode %q", comp.Spec.Mode)
	}
	// Extract resources
	var resources []ManagedResource

	if comp.Spec.Mode == "Pipeline" {
		if len(comp.Spec.Pipeline) == 0 {
			return "", fmt.Errorf("pipeline mode requires spec.pipeline steps")
		}
		// Parse pipeline mode
		var err error
		resources, err = g.extractPipelineResources(comp, opts)
		if err != nil {
			return "", err
		}
	} else {
		// Parse resources mode
		resources = g.extractResources(comp.Spec.Resources, opts)
	}

	return g.generateMarkdown(comp, resources, opts)
}

// extractPipelineResources extracts resources from pipeline mode
func (g *Generator) extractPipelineResources(comp *Composition, opts Options) ([]ManagedResource, error) {
	var resources []ManagedResource

	for _, step := range comp.Spec.Pipeline {
		input, exists := step.Input["resources"]
		if !exists {
			continue
		}
		// Decode both modes into the same types, preserving transforms and combine
		// variables instead of maintaining a second, incomplete patch parser.
		data, err := yaml.Marshal(input)
		if err != nil {
			return nil, fmt.Errorf("pipeline step %q resources: %w", step.Step, err)
		}
		var typed []Resource
		if err := yaml.Unmarshal(data, &typed); err != nil {
			return nil, fmt.Errorf("pipeline step %q resources: %w", step.Step, err)
		}
		resources = append(resources, g.extractResources(typed, opts)...)
	}

	return resources, nil
}

// extractResources extracts resources from resources mode
func (g *Generator) extractResources(resources []Resource, opts Options) []ManagedResource {
	var result []ManagedResource

	for _, res := range resources {
		mr := ManagedResource{
			Name:       res.Name,
			Kind:       getString(res.Base, "kind"),
			APIVersion: getString(res.Base, "apiVersion"),
		}

		if opts.ShowPatches {
			mr.Patches = g.extractPatches(res.Patches)
		}

		result = append(result, mr)
	}

	return result
}

// extractPatches extracts patch information
func (g *Generator) extractPatches(patches []Patch) []PatchInfo {
	var result []PatchInfo

	for _, p := range patches {
		if p.Type == "" {
			p.Type = "FromCompositeFieldPath"
		}
		info := PatchInfo{
			Type:           p.Type,
			XRDField:       p.FromFieldPath,
			MappedTo:       p.ToFieldPath,
			Transformation: g.formatTransformation(p),
		}
		if p.Combine != nil {
			var variables []string
			for _, variable := range p.Combine.Variables {
				variables = append(variables, variable.FromFieldPath)
			}
			info.XRDField = strings.Join(variables, ", ")
		}
		result = append(result, info)
	}

	return result
}

// formatTransformation formats the transformation description
func (g *Generator) formatTransformation(p Patch) string {
	var details []string
	if p.Combine != nil && p.Combine.String != nil {
		details = append(details, p.Combine.String.Fmt)
	}
	if len(p.Transforms) > 0 {
		data, err := json.Marshal(p.Transforms)
		if err == nil {
			details = append(details, string(data))
		} else {
			details = append(details, "Transforms defined (cannot display as JSON)")
		}
	}
	if len(details) > 0 {
		return strings.Join(details, "; ")
	}
	if p.Type == "PatchSet" {
		return "Patch set: " + p.PatchSetName + " (not expanded)"
	}
	if p.Type == "FromCompositeFieldPath" || p.Type == "ToCompositeFieldPath" {
		return "Direct copy"
	}
	return p.Type
}

// generateMarkdown generates the final markdown output
func (g *Generator) generateMarkdown(comp *Composition, resources []ManagedResource, opts Options) (string, error) {
	// Sort resources by name
	sort.Slice(resources, func(i, j int) bool {
		return resources[i].Name < resources[j].Name
	})

	tmpl := `# {{ .Composition.Spec.CompositeTypeRef.Kind }} Composition

**Composition Name:** {{ .Name }}  
**Composite Type:** {{ .Composition.Spec.CompositeTypeRef.APIVersion }}/{{ .Composition.Spec.CompositeTypeRef.Kind }}  
{{ if .Composition.Spec.Mode }}**Mode:** {{ .Composition.Spec.Mode }}{{ end }}
{{ if .Composition.Spec.Pipeline }}
## Pipeline Steps

| Step | Function |
|------|----------|
{{ range .Composition.Spec.Pipeline -}}
| {{ cell .Step }} | {{ cell .FunctionRef.Name }} |
{{ end }}
Pipeline functions are not executed. Only static input.resources entries are listed below; functions may create additional resources at runtime.
{{ end }}

## Managed Resources

Static resource definitions found: {{ len .Resources }}.

| Resource Name | Kind | API Version |
|---------------|------|-------------|
{{ range .Resources -}}
| {{ cell .Name }} | {{ cell .Kind }} | {{ cell .APIVersion }} |
{{ end }}
{{ if .ShowPatches }}
## Field Mappings
{{ range .Resources }}
### {{ .Name }} ({{ .Kind }})
{{ if .Patches }}
FromComposite patches copy from the composite to the resource; ToComposite patches copy from the resource to the composite.

| Patch Type | Source Field | Destination Field | Transformation |
|------------|--------------|-------------------|----------------|
{{ range .Patches -}}
| {{ cell .Type }} | {{ if .XRDField }}{{ cell .XRDField }}{{ else }}-{{ end }} | {{ if .MappedTo }}{{ cell .MappedTo }}{{ else }}-{{ end }} | {{ cell .Transformation }} |
{{ end }}
{{ else }}
No patches defined.
{{ end }}
{{ end }}
{{ end }}
`

	t, err := template.New("markdown").Funcs(template.FuncMap{"cell": document.Cell}).Parse(tmpl)
	if err != nil {
		return "", err
	}

	name := "unknown"
	if metadata, ok := comp.Metadata["name"].(string); ok {
		name = metadata
	}

	data := struct {
		Composition *Composition
		Name        string
		Resources   []ManagedResource
		ShowPatches bool
	}{
		Composition: comp,
		Name:        name,
		Resources:   resources,
		ShowPatches: opts.ShowPatches,
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Helper functions
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
