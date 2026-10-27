package composition

import (
	"fmt"

	"github.com/michielvha/crossplane-docs/internal/engine"
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
	PatchSets        []PatchSet       `yaml:"patchSets,omitempty"`
	Resources        []Resource       `yaml:"resources,omitempty"`
	Pipeline         []PipelineStep   `yaml:"pipeline,omitempty"`
}

type PatchSet struct {
	Name    string  `yaml:"name"`
	Patches []Patch `yaml:"patches"`
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

// GenerateFromFile generates Markdown through the shared documentation model.
func (g *Generator) GenerateFromFile(filename string, opts Options) (string, error) {
	config := engine.Defaults()
	config.Settings.Patches = opts.ShowPatches
	inputs, err := engine.Load(filename, config)
	if err != nil {
		return "", err
	}
	doc, err := engine.Build(inputs, engine.Options{Kind: "Composition", Config: config})
	if err != nil {
		return "", err
	}
	data, err := engine.Format(doc, config)
	return string(data), err
}

// Generate generates Markdown from the public resource type.
func (g *Generator) Generate(resource *Composition, opts Options) (string, error) {
	if resource == nil {
		return "", fmt.Errorf("resource must not be nil")
	}
	data, err := yaml.Marshal(resource)
	if err != nil {
		return "", err
	}
	inputs, err := engine.Read(data, "")
	if err != nil {
		return "", err
	}
	config := engine.Defaults()
	config.Settings.Patches = opts.ShowPatches
	doc, err := engine.Build(inputs, engine.Options{Kind: "Composition", Config: config})
	if err != nil {
		return "", err
	}
	data, err = engine.Format(doc, config)
	return string(data), err
}
