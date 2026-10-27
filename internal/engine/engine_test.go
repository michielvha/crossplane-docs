package engine

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"
)

func writeFixture(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return p
}

func model(t *testing.T) Document {
	t.Helper()
	c := Defaults()
	inputs, err := Load("../../examples/platform.yaml", c)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := Build(inputs, Options{Config: c, AllVersions: true})
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

func TestFormatsShareModel(t *testing.T) {
	doc := model(t)
	for _, format := range []string{"markdown", "html", "asciidoc", "json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			c := Defaults()
			c.Formatter = format
			out, err := Format(doc, c)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(out), "spec.networks[].cidr") || !strings.Contains(string(out), "minReplicas") {
				t.Fatal("schema facts lost")
			}
			switch format {
			case "json":
				var round Document
				if err := json.Unmarshal(out, &round); err != nil {
					t.Fatal(err)
				}
				if len(round.APIs) != 2 {
					t.Fatal("versions lost")
				}
			case "yaml":
				var round Document
				if err := yaml.Unmarshal(out, &round); err != nil {
					t.Fatal(err)
				}
				if len(round.APIs) != 2 {
					t.Fatal("versions lost")
				}
			}
		})
	}
}

func TestReferences(t *testing.T) {
	dir := t.TempDir()
	file := writeFixture(t, dir, "xrd.yaml", `apiVersion: apiextensions.crossplane.io/v2
kind: CompositeResourceDefinition
spec:
  group: example.org
  names: {kind: XExample}
  versions:
    - name: v1
      served: true
      schema:
        openAPIV3Schema:
          type: object
          properties:
            spec:
              type: object
              properties:
                parent: {$ref: '#/definitions/node'}
definitions:
  node:
    type: object
    properties:
      name: {type: string}
      child: {$ref: '#/definitions/node'}
`)
	c := Defaults()
	inputs, err := Load(file, c)
	if err != nil {
		t.Fatal(err)
	}
	doc, err := Build(inputs, Options{Config: c})
	if err != nil {
		t.Fatal(err)
	}
	out, _ := json.Marshal(doc)
	if !strings.Contains(string(out), "spec.parent.name") || !strings.Contains(string(out), `"type":"reference"`) {
		t.Fatalf("references not resolved/cycle marked: %s", out)
	}
	inputs[0].Object["definitions"] = Object{}
	if _, err := Build(inputs, Options{Config: c}); err == nil {
		t.Fatal("missing reference accepted")
	}
}

func TestConfigValidation(t *testing.T) {
	for _, content := range []string{"version: 2\n", "version: 1\nunknown: true\n", "version: 1\nsettings:\n  nested: invalid\n"} {
		t.Run(content, func(t *testing.T) {
			dir := t.TempDir()
			file := writeFixture(t, dir, ".crossplane-docs.yml", content)
			c, err := LoadConfig(dir, file)
			if err == nil {
				err = c.Validate()
			}
			if err == nil {
				t.Fatal("bad config accepted")
			}
		})
	}
}

func TestOutputSafety(t *testing.T) {
	for _, markers := range []string{"no markers", "<!-- END_CROSSPLANE_DOCS --> before <!-- BEGIN_CROSSPLANE_DOCS -->", "<!-- BEGIN_CROSSPLANE_DOCS --><!-- BEGIN_CROSSPLANE_DOCS --><!-- END_CROSSPLANE_DOCS -->"} {
		t.Run(markers, func(t *testing.T) {
			file := writeFixture(t, t.TempDir(), "README.md", markers)
			c := Defaults()
			c.Output = Output{File: file, Mode: "inject"}
			if err := Write([]byte("new"), c, false); err == nil {
				t.Fatal("invalid markers accepted")
			}
			got, _ := os.ReadFile(file)
			if string(got) != markers {
				t.Fatal("error modified file")
			}
		})
	}
}

func TestUnknownFunctionsAndPatchSetErrors(t *testing.T) {
	inputs, err := Load("../../examples/platform/go-templating.yaml", Defaults())
	if err != nil {
		t.Fatal(err)
	}
	pipeline := List(at(inputs[0].Object, "spec", "pipeline"))
	Map(Map(pipeline[0])["input"])["apiVersion"] = "custom.example.org/v1"
	doc, err := Build(inputs, Options{Config: Defaults()})
	if err != nil {
		t.Fatal(err)
	}
	if doc.Compositions[0].Steps[0].Adapter != "custom" {
		t.Fatal("unknown input incorrectly classified")
	}
	sets := map[string][]interface{}{"a": {Object{"type": "PatchSet", "patchSetName": "a"}}}
	if _, err := expandPatches(sets["a"], sets, map[string]bool{}); err == nil {
		t.Fatal("cyclic patch set accepted")
	}
	if _, err := expandPatches([]interface{}{Object{"type": "PatchSet", "patchSetName": "missing"}}, sets, map[string]bool{}); err == nil {
		t.Fatal("missing patch set accepted")
	}
}

func TestRendererFailureAndTimeout(t *testing.T) {
	dir := t.TempDir()
	script := writeFixture(t, dir, "crossplane", "#!/bin/sh\nprintf 'fixture error' >&2\nexit 7\n")
	if err := os.Chmod(script, 0o700); err != nil {
		t.Fatal(err)
	}
	o := RenderOptions{Binary: script, Timeout: time.Second}
	if _, err := o.run(context.Background(), "render"); err == nil || !strings.Contains(err.Error(), "fixture error") {
		t.Fatalf("lost renderer failure: %v", err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := o.run(ctx); err == nil {
		t.Fatal("cancelled execution succeeded")
	}
}

func TestCredentialRedaction(t *testing.T) {
	v := redacted(Object{"credentials": Object{"password": "secret"}, "source": "safe"})
	out, _ := json.Marshal(v)
	if strings.Contains(string(out), "secret") || !strings.Contains(string(out), "safe") {
		t.Fatalf("bad redaction: %s", out)
	}
}

func TestExplicitFileReferences(t *testing.T) {
	dir := t.TempDir()
	ref := writeFixture(t, dir, "types.yaml", "definitions:\n  labels:\n    type: object\n    additionalProperties: {type: string}\n")
	inputs, err := Load("../../examples/platform.yaml", Defaults())
	if err != nil {
		t.Fatal(err)
	}
	schema := Map(at(Map(List(at(inputs[0].Object, "spec", "versions"))[0]), "schema", "openAPIV3Schema"))
	Map(Map(schema["properties"])["spec"])["properties"] = Object{"labels": Object{"$ref": "types.yaml#/definitions/labels"}}
	inputs[0].File = filepath.Join(dir, "xrd.yaml")
	if _, err := Build(inputs, Options{Config: Defaults()}); err == nil {
		t.Fatal("unsupplied external schema read implicitly")
	}
	doc, err := Build(inputs, Options{Config: Defaults(), RefFiles: []string{ref}})
	if err != nil {
		t.Fatal(err)
	}
	data, _ := json.Marshal(doc)
	if !strings.Contains(string(data), "spec.labels.*") {
		t.Fatalf("external map schema missing: %s", data)
	}
}

func TestLosslessDefaults(t *testing.T) {
	for _, value := range []interface{}{false, 0, "", nil, []interface{}{}, Object{}} {
		t.Run(compact(value), func(t *testing.T) {
			f := Field{Path: "spec.value", Type: "any", HasDefault: true, Default: value}
			doc := Document{SchemaVersion: "1", Title: "Defaults", APIs: []API{{Kind: "X", Version: "v1", Fields: []Field{f}}}}
			c := Defaults()
			c.Formatter = "json"
			data, err := Format(doc, c)
			if err != nil {
				t.Fatal(err)
			}
			var parsed map[string]interface{}
			if err := json.Unmarshal(data, &parsed); err != nil {
				t.Fatal(err)
			}
			field := Map(Map(parsed["apis"].([]interface{})[0])["fields"].([]interface{})[0])
			if _, exists := field["default"]; !exists {
				t.Fatalf("default lost: %s", data)
			}
			if field["hasDefault"] != true {
				t.Fatal("default presence lost")
			}
		})
	}
}

func TestSavedCompositeClassification(t *testing.T) {
	data := []byte("apiVersion: example.org/v1\nkind: XExample\nmetadata: {name: test}\nstatus: {phase: Ready}\n---\napiVersion: v1\nkind: ConfigMap\nmetadata: {name: child}\n")
	s, err := ParseScenario(data, "saved", "", Composition{APIVersion: "example.org/v1", Kind: "XExample"})
	if err != nil {
		t.Fatal(err)
	}
	if text(s.Composite, "status", "phase") != "Ready" || len(s.Resources) != 1 {
		t.Fatalf("saved composite misclassified: %+v", s)
	}
}

func TestHTMLNavigationAndEscaping(t *testing.T) {
	doc := model(t)
	doc.Title = "<script>unsafe</script>"
	c := Defaults()
	c.Formatter = "html"
	data, err := Format(doc, c)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "<script>") || strings.Contains(string(data), "href=\"#scenarios\"") {
		t.Fatalf("unsafe HTML or broken navigation: %s", data)
	}
}
