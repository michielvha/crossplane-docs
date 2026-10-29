package generator

import (
	"fmt"

	"github.com/michielvha/crossplane-docs/internal/engine"
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
	APIVersion string            `yaml:"apiVersion"`
	Kind       string            `yaml:"kind"`
	Metadata   metav1.ObjectMeta `yaml:"metadata"`
	Spec       XRDSpec           `yaml:"spec"`
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
	Name                     string                   `yaml:"name"`
	Served                   bool                     `yaml:"served"`
	Referenceable            bool                     `yaml:"referenceable"`
	Schema                   XRDVersionSchema         `yaml:"schema"`
	AdditionalPrinterColumns []map[string]interface{} `yaml:"additionalPrinterColumns,omitempty"`
}

// XRDVersionSchema contains the OpenAPI schema
type XRDVersionSchema struct {
	OpenAPIV3Schema OpenAPISchema `yaml:"openAPIV3Schema"`
}

// OpenAPISchema represents the OpenAPI v3 schema
type OpenAPISchema struct {
	Type                   string                   `yaml:"type"`
	Description            string                   `yaml:"description,omitempty"`
	Properties             map[string]OpenAPISchema `yaml:"properties,omitempty"`
	Items                  *OpenAPISchema           `yaml:"items,omitempty"`
	Required               []string                 `yaml:"required,omitempty"`
	Default                interface{}              `yaml:"default,omitempty"`
	Enum                   []interface{}            `yaml:"enum,omitempty"`
	Minimum                *float64                 `yaml:"minimum,omitempty"`
	Maximum                *float64                 `yaml:"maximum,omitempty"`
	MinItems               *int                     `yaml:"minItems,omitempty"`
	MaxItems               *int                     `yaml:"maxItems,omitempty"`
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

// GenerateFromFile generates Markdown through the shared documentation model.
func (g *Generator) GenerateFromFile(filename string, opts Options) (string, error) {
	config := engine.Defaults()
	config.Settings.Nested = opts.ShowNested
	inputs, err := engine.Load(filename, config)
	if err != nil {
		return "", err
	}
	doc, err := engine.Build(inputs, engine.Options{Kind: "CompositeResourceDefinition", Config: config})
	if err != nil {
		return "", err
	}
	data, err := engine.Format(doc, config)
	return string(data), err
}

// Generate generates Markdown from the public resource type.
func (g *Generator) Generate(resource *XRD, opts Options) (string, error) {
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
	config.Settings.Nested = opts.ShowNested
	doc, err := engine.Build(inputs, engine.Options{Kind: "CompositeResourceDefinition", Config: config})
	if err != nil {
		return "", err
	}
	data, err = engine.Format(doc, config)
	return string(data), err
}
