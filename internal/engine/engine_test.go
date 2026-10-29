package engine

import (
	"context"
	"encoding/json"
	"fmt"
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

func TestScenarioInputProvenance(t *testing.T) {
	o := RenderOptions{XR: "xr.yaml", Composition: "composition.yaml", Functions: "functions.yaml", Observed: "observed.yaml", Required: []string{"required.yaml"}, Schemas: "schemas", Context: []string{`platform={"owner":"atlas","apiToken":"private"}`}}
	inputs := ScenarioInputs(o, "")
	raw, _ := json.Marshal(inputs)
	if strings.Contains(string(raw), "private") || !strings.Contains(string(raw), "atlas") || !strings.Contains(string(raw), "required.yaml") {
		t.Fatalf("bad input provenance: %s", raw)
	}
	doc := Document{Title: "Scenarios", SchemaVersion: "1", Scenarios: []Scenario{{Name: "update", Inputs: inputs}}}
	for _, format := range []string{"markdown", "html", "asciidoc", "json", "yaml"} {
		c := Defaults()
		c.Formatter = format
		data, err := Format(doc, c)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(data), "observed.yaml") || !strings.Contains(string(data), "update") {
			t.Fatalf("%s lost scenario inputs", format)
		}
	}
	o.Observed = "another-state.yaml"
	other, _ := json.Marshal(ScenarioInputs(o, ""))
	if string(other) == string(raw) {
		t.Fatal("different mock inputs are indistinguishable")
	}
}

func TestPatchSummariesPreserveRedaction(t *testing.T) {
	inputs, err := Read([]byte(`apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata: {name: redaction}
spec:
  compositeTypeRef: {apiVersion: example.org/v1, kind: XExample}
  resources:
  - name: config
    base: {apiVersion: v1, kind: ConfigMap}
    patches:
    - fromFieldPath: spec.mode
      toFieldPath: data.value
      transforms:
      - type: map
        map: {password: AUDIT_FIXTURE_SECRET, normal: visible-transform-value}
`), "redaction.yaml")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := Build(inputs, Options{Config: Defaults()})
	if err != nil {
		t.Fatal(err)
	}
	for _, format := range []string{"markdown", "html", "asciidoc", "json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			c := Defaults()
			c.Formatter = format
			data, err := Format(doc, c)
			if err != nil || strings.Contains(string(data), "AUDIT_FIXTURE_SECRET") || !strings.Contains(string(data), "visible-transform-value") {
				t.Fatalf("patch summary lost redaction or safe content: error=%v sensitive=%t safe=%t", err, strings.Contains(string(data), "AUDIT_FIXTURE_SECRET"), strings.Contains(string(data), "visible-transform-value"))
			}
		})
	}
}

func TestRootConstraintsAndContainerMetadata(t *testing.T) {
	inputs, err := Read([]byte(`apiVersion: apiextensions.crossplane.io/v2
kind: CompositeResourceDefinition
metadata: {name: xexamples.example.org}
spec:
  group: example.org
  names: {kind: XExample, plural: xexamples}
  versions:
  - name: v1
    served: true
    referenceable: true
    schema:
      openAPIV3Schema:
        type: object
        required: [spec]
        minProperties: 1
        x-kubernetes-validations:
        - rule: self.spec.min <= self.spec.max
          message: root-range-rule
        properties:
          spec:
            type: object
            description: spec-container-description
            nullable: true
            default: {min: 0, max: 2}
            properties:
              min: {type: integer}
              max: {type: integer}
          status:
            type: object
            description: status-container-description
            default: {}
            properties:
              ready: {type: boolean}
`), "root.yaml")
	if err != nil {
		t.Fatal(err)
	}
	doc, err := Build(inputs, Options{Config: Defaults()})
	if err != nil {
		t.Fatal(err)
	}
	fields := map[string]Field{}
	for _, field := range doc.APIs[0].Fields {
		fields[field.Path] = field
	}
	for _, path := range []string{"spec", "status"} {
		field, exists := fields[path]
		if !exists || !field.HasDefault || field.Description != path+"-container-description" {
			t.Errorf("%s container metadata was discarded: %+v", path, field)
		}
	}
	if !fields["spec"].Required || !fields["spec"].Nullable || compact(fields["status"].Default) != "{}" {
		t.Error("container required, nullable or empty default was lost")
	}
	for _, format := range []string{"markdown", "html", "asciidoc", "json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			c := Defaults()
			c.Formatter = format
			data, err := Format(doc, c)
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{"root-range-rule", "minProperties", "spec-container-description", "status-container-description"} {
				if !strings.Contains(string(data), want) {
					t.Errorf("%s lost %s", format, want)
				}
			}
		})
	}
}

func TestShallowReferencedContainers(t *testing.T) {
	for _, ref := range []string{"#/definitions/container", "#/definitions/chain", "types.yaml#/definitions/container", "types.yaml#/definitions/chain"} {
		t.Run(ref, func(t *testing.T) {
			dir := t.TempDir()
			definitions := `definitions:
  chain: {$ref: '#/definitions/container'}
  container:
    type: object
    properties:
      region: {type: string}
      settings:
        type: object
        properties:
          enabled: {type: boolean}
`
			types := writeFixture(t, dir, "types.yaml", definitions)
			file := writeFixture(t, dir, "xrd.yaml", fmt.Sprintf(`apiVersion: apiextensions.crossplane.io/v2
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
          spec: {$ref: %q}
`, ref)+definitions)
			c := Defaults()
			c.Settings.Nested = false
			inputs, err := Load(file, c)
			if err != nil {
				t.Fatal(err)
			}
			doc, err := Build(inputs, Options{Config: c, RefFiles: []string{types}})
			if err != nil {
				t.Fatal(err)
			}
			var paths []string
			for _, field := range doc.APIs[0].Fields {
				paths = append(paths, field.Path)
			}
			if strings.Join(paths, ",") != "spec.region,spec.settings" {
				t.Fatalf("reference indirection changed shallow properties: %v", paths)
			}
		})
	}
}

func TestScenarioMetadataRequiresExactAPIGroup(t *testing.T) {
	data := []byte(`apiVersion: render.crossplane.io/v1beta1
kind: Context
---
apiVersion: render.crossplane.io/v1beta1
kind: Result
---
apiVersion: audit.render.crossplane.io/v1
kind: Context
metadata: {name: external-context}
---
apiVersion: render.crossplane.io.example.org/v1
kind: Result
metadata: {name: external-result}
`)
	scenario, err := ParseScenario(data, "groups", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(scenario.Resources) != 2 || len(scenario.Context) != 1 || len(scenario.Results) != 1 {
		t.Fatalf("foreign resources mistaken for renderer metadata: resources=%d context=%d results=%d", len(scenario.Resources), len(scenario.Context), len(scenario.Results))
	}
}

func TestDescriptionPresentationAcrossFormats(t *testing.T) {
	doc := model(t)
	doc.APIs[0].Description = "api-description-marker"
	doc.APIs[0].Fields[0].Description = "field-description-marker"
	for _, format := range []string{"markdown", "html", "asciidoc", "json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			for _, enabled := range []bool{false, true} {
				c := Defaults()
				c.Formatter = format
				c.Settings.Descriptions = enabled
				data, err := Format(doc, c)
				if err != nil {
					t.Fatal(err)
				}
				want := enabled || format == "json" || format == "yaml"
				for _, description := range []string{"api-description-marker", "field-description-marker"} {
					if strings.Contains(string(data), description) != want {
						t.Errorf("%s ignored descriptions=%t for %s", format, enabled, description)
					}
				}
			}
		})
	}
}
