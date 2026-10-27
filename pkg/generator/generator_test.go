package generator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExampleGolden(t *testing.T) {
	for _, nested := range []bool{true, false} {
		name := "xrd"
		if !nested {
			name = "xrd-shallow"
		}
		t.Run(name, func(t *testing.T) {
			want, err := os.ReadFile(filepath.Join("..", "..", "examples", name+".md"))
			if err != nil {
				t.Fatal(err)
			}
			for range 20 {
				got, err := New().GenerateFromFile("../../examples/xrd.yaml", Options{ShowNested: nested})
				if err != nil {
					t.Fatal(err)
				}
				if got != string(want) {
					t.Fatalf("output differs from examples/%s.md:\n%s", name, got)
				}
			}
		})
	}
}

func TestFormatTypePreservesEnumType(t *testing.T) {
	for _, typ := range []string{"integer", "number", "boolean", "string"} {
		t.Run(typ, func(t *testing.T) {
			got := New().formatType(OpenAPISchema{Type: typ, Enum: []interface{}{1, 2}})
			if got != typ {
				t.Fatalf("got %q, want %q", got, typ)
			}
		})
	}
}

func TestXRDValidation(t *testing.T) {
	for _, input := range []string{
		"", "kind: Composition", "kind: [broken", "apiVersion: v1\nkind: ConfigMap",
		"apiVersion: apiextensions.crossplane.io/v1\nkind: CompositeResourceDefinition\nspec: {}",
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
		t.Fatal("nil XRD should fail")
	}
	if _, err := New().GenerateFromFile("missing.yaml", Options{}); err == nil {
		t.Fatal("missing input should fail")
	}
}

func TestXRDRejectsMultipleDocuments(t *testing.T) {
	data, err := os.ReadFile("../../examples/xrd.yaml")
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(t.TempDir(), "bundle.yaml")
	if err := os.WriteFile(file, append(data, []byte("\n---\nkind: ConfigMap\n")...), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := New().GenerateFromFile(file, Options{}); err == nil {
		t.Fatal("multiple documents should fail instead of ignoring later documents")
	}
}

func TestServedVersionSelection(t *testing.T) {
	xrd := &XRD{APIVersion: "apiextensions.crossplane.io/v2", Kind: "CompositeResourceDefinition", Spec: XRDSpec{
		Group: "example.org", Names: XRDNames{Kind: "XDatabase"},
		Versions: []XRDVersion{{Name: "v1alpha1"}, {Name: "v1", Served: true}},
	}}
	got, err := New().Generate(xrd, Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "| API Version | v1 |") {
		t.Fatal("did not select first served version")
	}
	xrd.Spec.Versions[1].Served = false
	got, err = New().Generate(xrd, Options{})
	if err != nil || !strings.Contains(got, "| API Version | v1alpha1 |") {
		t.Fatalf("first version fallback failed: %s, %v", got, err)
	}
}
