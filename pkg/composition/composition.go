package composition

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"
	"text/template"

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
	Type          string                 `yaml:"type"`
	FromFieldPath string                 `yaml:"fromFieldPath,omitempty"`
	ToFieldPath   string                 `yaml:"toFieldPath,omitempty"`
	Combine       *Combine               `yaml:"combine,omitempty"`
	Policy        map[string]interface{} `yaml:"policy,omitempty"`
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
	if err := yaml.Unmarshal(data, &comp); err != nil {
		return "", fmt.Errorf("failed to parse Composition YAML: %w", err)
	}

	return g.Generate(&comp, opts)
}

// Generate generates documentation from a Composition struct
func (g *Generator) Generate(comp *Composition, opts Options) (string, error) {
	// Extract resources
	var resources []ManagedResource

	if comp.Spec.Mode == "Pipeline" && len(comp.Spec.Pipeline) > 0 {
		// Parse pipeline mode
		resources = g.extractPipelineResources(comp, opts)
	} else {
		// Parse resources mode
		resources = g.extractResources(comp.Spec.Resources, opts)
	}

	return g.generateMarkdown(comp, resources, opts)
}

// extractPipelineResources extracts resources from pipeline mode
func (g *Generator) extractPipelineResources(comp *Composition, opts Options) []ManagedResource {
	var resources []ManagedResource

	for _, step := range comp.Spec.Pipeline {
		if input, ok := step.Input["resources"].([]interface{}); ok {
			for _, r := range input {
				if resMap, ok := r.(map[string]interface{}); ok {
					resource := g.parseResource(resMap, opts)
					resources = append(resources, resource)
				}
			}
		}
	}

	return resources
}

// extractResources extracts resources from resources mode
func (g *Generator) extractResources(resources []Resource, opts Options) []ManagedResource {
	var result []ManagedResource

	for _, res := range resources {
		mr := ManagedResource{
			Name:       res.Name,
			Kind:       getStringFromMap(res.Base, "kind"),
			APIVersion: getStringFromMap(res.Base, "apiVersion"),
		}

		if opts.ShowPatches {
			mr.Patches = g.extractPatches(res.Patches)
		}

		result = append(result, mr)
	}

	return result
}

// parseResource parses a resource from map
func (g *Generator) parseResource(resMap map[string]interface{}, opts Options) ManagedResource {
	resource := ManagedResource{
		Name: getString(resMap, "name"),
	}

	if base, ok := resMap["base"].(map[string]interface{}); ok {
		resource.Kind = getString(base, "kind")
		resource.APIVersion = getString(base, "apiVersion")
	}

	if opts.ShowPatches {
		if patches, ok := resMap["patches"].([]interface{}); ok {
			resource.Patches = g.parsePatchesFromInterface(patches)
		}
	}

	return resource
}

// extractPatches extracts patch information
func (g *Generator) extractPatches(patches []Patch) []PatchInfo {
	var result []PatchInfo

	for _, p := range patches {
		info := PatchInfo{
			XRDField:       p.FromFieldPath,
			MappedTo:       p.ToFieldPath,
			Transformation: g.formatTransformation(p),
		}
		result = append(result, info)
	}

	return result
}

// parsePatchesFromInterface parses patches from interface
func (g *Generator) parsePatchesFromInterface(patches []interface{}) []PatchInfo {
	var result []PatchInfo

	for _, p := range patches {
		if patchMap, ok := p.(map[string]interface{}); ok {
			info := PatchInfo{
				XRDField: getString(patchMap, "fromFieldPath"),
				MappedTo: getString(patchMap, "toFieldPath"),
			}

			// Handle combine transformations
			if combine, ok := patchMap["combine"].(map[string]interface{}); ok {
				if str, ok := combine["string"].(map[string]interface{}); ok {
					if fmt, ok := str["fmt"].(string); ok {
						info.Transformation = fmt
					}
				}
			} else if info.XRDField != "" {
				info.Transformation = "Direct copy"
			}

			if info.XRDField != "" || info.MappedTo != "" {
				result = append(result, info)
			}
		}
	}

	return result
}

// formatTransformation formats the transformation description
func (g *Generator) formatTransformation(p Patch) string {
	if p.Combine != nil && p.Combine.String != nil {
		return p.Combine.String.Fmt
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

## Managed Resources

This composition creates {{ len .Resources }} managed resource(s):

| Resource Name | Kind | API Version |
|---------------|------|-------------|
{{ range .Resources -}}
| {{ .Name }} | {{ .Kind }} | {{ .APIVersion }} |
{{ end }}
{{ if .ShowPatches }}
## Field Mappings
{{ range .Resources }}
### {{ .Name }} ({{ .Kind }})
{{ if .Patches }}
| XRD Field | Mapped To | Transformation |
|-----------|-----------|----------------|
{{ range .Patches -}}
| {{ if .XRDField }}{{ .XRDField }}{{ else }}-{{ end }} | {{ .MappedTo }} | {{ .Transformation }} |
{{ end }}
{{ else }}
No patches defined.
{{ end }}
{{ end }}
{{ end }}
`

	t, err := template.New("markdown").Parse(tmpl)
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

func getStringFromMap(m map[string]interface{}, key string) string {
	keys := strings.Split(key, ".")
	current := m

	for i, k := range keys {
		if i == len(keys)-1 {
			if v, ok := current[k].(string); ok {
				return v
			}
			return ""
		}
		if next, ok := current[k].(map[string]interface{}); ok {
			current = next
		} else {
			return ""
		}
	}
	return ""
}
