package composition

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCompositionExamples(t *testing.T) {
	for _, name := range []string{"composition", "pipeline"} {
		t.Run(name, func(t *testing.T) {
			got, err := New().GenerateFromFile("../../examples/"+name+".yaml", Options{ShowPatches: true})
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile("../../examples/" + name + ".md")
			if err != nil {
				t.Fatal(err)
			}
			if got != string(want) {
				t.Fatalf("output differs from examples/%s.md:\n%s", name, got)
			}
			got, err = New().GenerateFromFile("../../examples/"+name+".yaml", Options{})
			if err != nil || strings.Contains(got, "Field Mappings") || !strings.Contains(got, "Instance") {
				t.Fatalf("hidden patches should preserve resource inventory: %s, %v", got, err)
			}
		})
	}
}

func TestPatchDescriptions(t *testing.T) {
	for _, name := range []string{"composition", "pipeline"} {
		t.Run(name, func(t *testing.T) {
			got, err := New().GenerateFromFile("../../examples/"+name+".yaml", Options{ShowPatches: true})
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{"Multiply", "metadata.name, spec.parameters.region", "ToCompositeFieldPath", "FromCompositeFieldPath"} {
				if !strings.Contains(got, want) {
					t.Errorf("output missing %q", want)
				}
			}
		})
	}
}

func TestCompositionValidation(t *testing.T) {
	for _, input := range []string{
		"", "kind: CompositeResourceDefinition", "kind: [broken",
		"apiVersion: apiextensions.crossplane.io/v1\nkind: Composition\nspec: {}",
		"apiVersion: apiextensions.crossplane.io/v1\nkind: Composition\nspec:\n  compositeTypeRef: {apiVersion: example.org/v1, kind: XDatabase}\n  mode: Unknown",
		"apiVersion: apiextensions.crossplane.io/v1\nkind: Composition\nspec:\n  compositeTypeRef: {apiVersion: example.org/v1, kind: XDatabase}\n  mode: Pipeline\n  pipeline:\n    - step: broken\n      functionRef: {name: function-patch-and-transform}\n      input: {resources: invalid}",
	} {
		t.Run(input, func(t *testing.T) {
			file := filepath.Join(t.TempDir(), "input.yaml")
			if err := os.WriteFile(file, []byte(input), 0o600); err != nil {
				t.Fatal(err)
			}
			if got, err := New().GenerateFromFile(file, Options{}); err == nil || got != "" {
				t.Fatalf("expected failure, got %q, %v", got, err)
			}
		})
	}
	if _, err := New().Generate(nil, Options{}); err == nil {
		t.Fatal("nil composition should fail")
	}
	if _, err := New().GenerateFromFile("missing.yaml", Options{}); err == nil {
		t.Fatal("missing input should fail")
	}
}

func TestCompositionRejectsMultipleDocuments(t *testing.T) {
	data, err := os.ReadFile("../../examples/composition.yaml")
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "bundle.yaml")
	if err := os.WriteFile(file, append(data, []byte("\n---\nkind: ConfigMap\n")...), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := New().GenerateFromFile(file, Options{}); err == nil {
		t.Fatal("multiple documents should fail")
	}
}

func TestPatchSetAndCombinedTransforms(t *testing.T) {
	comp := &Composition{
		APIVersion: "apiextensions.crossplane.io/v1", Kind: "Composition",
		Spec: CompositionSpec{
			CompositeTypeRef: CompositeTypeRef{APIVersion: "example.org/v1", Kind: "XDatabase"},
			Resources: []Resource{{Name: "database", Patches: []Patch{
				{Type: "PatchSet", PatchSetName: "common"},
				{
					Type: "CombineFromComposite", Combine: &Combine{String: &StringFmt{Fmt: "%s-name"}},
					Transforms: []map[string]interface{}{{"type": "string", "string": map[string]interface{}{"type": "Convert", "convert": "ToLower"}}},
				},
			}}},
		},
	}
	got, err := New().Generate(comp, Options{ShowPatches: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Patch set: common (not expanded)", "%s-name;", "ToLower"} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q: %s", want, got)
		}
	}
}

func TestRuntimeOnlyPipeline(t *testing.T) {
	comp := &Composition{
		APIVersion: "apiextensions.crossplane.io/v1", Kind: "Composition",
		Spec: CompositionSpec{
			Mode: "Pipeline", CompositeTypeRef: CompositeTypeRef{APIVersion: "example.org/v1", Kind: "XDatabase"},
			Pipeline: []PipelineStep{{Step: "render", FunctionRef: FunctionRef{Name: "function-go-templating"}}},
		},
	}
	got, err := New().Generate(comp, Options{})
	if err != nil || !strings.Contains(got, "function-go-templating") || !strings.Contains(got, "functions may create additional resources at runtime") {
		t.Fatalf("runtime-only pipelines need an explicit limitation: %s, %v", got, err)
	}
	comp.Spec.Pipeline = nil
	if _, err := New().Generate(comp, Options{}); err == nil {
		t.Fatal("empty pipeline should fail")
	}
}
