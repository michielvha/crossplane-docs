package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStructuredAndHTMLOutput(t *testing.T) {
	for _, format := range []string{"json", "yaml", "html", "asciidoc"} {
		t.Run(format, func(t *testing.T) {
			out, _, err := execute("xrd", "../examples/xrd.yaml", "--format", format)
			if err != nil {
				t.Fatal(err)
			}
			if format == "json" {
				var v map[string]interface{}
				if err := json.Unmarshal([]byte(out), &v); err != nil {
					t.Fatal(err)
				}
				if v["schemaVersion"] != "1" {
					t.Fatalf("missing model contract: %s", out)
				}
			}
			if format == "html" && (!strings.Contains(out, "<!doctype html>") || !strings.Contains(out, `id="apis"`) || !strings.Contains(out, `class="field-entry"`)) {
				t.Fatalf("not standalone HTML: %s", out)
			}
		})
	}
}

func TestHTMLLayoutFlagOverridesConfig(t *testing.T) {
	dir := t.TempDir()
	config := filepath.Join(dir, "config.yml")
	if err := os.WriteFile(config, []byte("version: 1\nformatter: html\nhtml-layout: explorer\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, layout := range []string{"reference", "explorer", "compact"} {
		out, _, err := execute("xrd", "../examples/xrd.yaml", "--config", config, "--html-layout", layout)
		if err != nil || !strings.Contains(out, `data-html-layout="`+layout+`"`) {
			t.Fatalf("HTML layout override %s failed: %v", layout, err)
		}
	}
	if _, _, err := execute("xrd", "../examples/xrd.yaml", "--html-layout", "invalid"); err == nil {
		t.Fatal("invalid HTML layout accepted")
	}
}

func TestConfigInjectionAndCheck(t *testing.T) {
	dir := t.TempDir()
	input, err := os.ReadFile("../examples/xrd.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "xrd.yaml"), input, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".crossplane-docs.yml"), []byte("version: 1\nformatter: markdown\noutput:\n  file: README.md\n  mode: inject\nsettings:\n  nested: false\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "README.md")
	if err := os.WriteFile(target, []byte("before\n<!-- BEGIN_CROSSPLANE_DOCS -->\nold\n<!-- END_CROSSPLANE_DOCS -->\nafter\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := execute("generate", dir, "--check"); err == nil {
		t.Fatal("stale docs must report drift")
	}
	if _, _, err := execute("generate", dir); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(target)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(first), "before\n") || !strings.HasSuffix(string(first), "after\n") {
		t.Fatal("injection changed surrounding content")
	}
	if _, _, err := execute("generate", dir, "--check"); err != nil {
		t.Fatal(err)
	}
	if _, _, err := execute("generate", dir); err != nil {
		t.Fatal(err)
	}
	second, _ := os.ReadFile(target)
	if string(first) != string(second) {
		t.Fatal("injection is not idempotent")
	}
}

func TestEmptyTemplateOutputAndCheck(t *testing.T) {
	dir := t.TempDir()
	template := filepath.Join(dir, "empty.tmpl")
	if err := os.WriteFile(template, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(dir, "output.md")
	args := []string{"xrd", "../examples/xrd.yaml", "--template", template, "--output", output}
	if _, _, err := execute(append(args, "--check")...); err == nil {
		t.Fatal("missing output must report drift even when the template is empty")
	}
	if _, err := os.Stat(output); !os.IsNotExist(err) {
		t.Fatalf("drift check created output: %v", err)
	}
	if _, _, err := execute(args...); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(output)
	if err != nil || len(data) != 0 {
		t.Fatalf("empty output was not created: %q %v", data, err)
	}
	if _, _, err := execute(append(args, "--check")...); err != nil {
		t.Fatalf("existing empty output should be current: %v", err)
	}
}

func TestStructuredOutputRetainsFullNestedModel(t *testing.T) {
	config := filepath.Join(t.TempDir(), "shallow.yml")
	if err := os.WriteFile(config, []byte("version: 1\nsettings:\n  nested: false\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, format := range []string{"json", "yaml"} {
		t.Run(format, func(t *testing.T) {
			for _, options := range [][]string{{"--show-nested=false"}, {"--config", config}} {
				args := append([]string{"xrd", "../examples/platform.yaml", "--format", format}, options...)
				out, _, err := execute(args...)
				if err != nil || !strings.Contains(out, "spec.networks[].subnets[].zone") {
					t.Fatalf("structured model was trimmed by presentation settings %v: %v", options, err)
				}
			}
		})
	}
}

func TestVersionSelectionAndSchemaPaths(t *testing.T) {
	out, _, err := execute("generate", "../examples/platform.yaml", "--format", "json", "--all-versions")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"v1alpha1", "v1beta1", "spec.networks[].cidr", "spec.tags.*", "x-kubernetes-validations", "pattern"} {
		if !strings.Contains(out, want) {
			t.Errorf("missing %s", want)
		}
	}
	if _, _, err := execute("generate", "../examples/platform.yaml", "--api-version", "missing"); err == nil {
		t.Fatal("unknown version should fail")
	}
}

func TestConfigExplicitFalseOverrideAndUnknownKeys(t *testing.T) {
	dir := t.TempDir()
	input, err := os.ReadFile("../examples/xrd.yaml")
	if err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "xrd.yaml")
	if err := os.WriteFile(file, input, 0o600); err != nil {
		t.Fatal(err)
	}
	config := filepath.Join(dir, ".crossplane-docs.yml")
	if err := os.WriteFile(config, []byte("version: 1\nsettings:\n  nested: false\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	out, _, err := execute("xrd", file, "--format", "markdown")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "spec.parameters.region") {
		t.Fatal("config false ignored")
	}
	out, _, err = execute("xrd", file, "--format", "markdown", "--show-nested=true")
	if err != nil || !strings.Contains(out, "spec.parameters.region") {
		t.Fatalf("CLI did not override config: %v", err)
	}
	if err := os.WriteFile(config, []byte("version: 1\nformater: html\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, _, err := execute("xrd", file); err == nil {
		t.Fatal("unknown config key accepted")
	}
}

func TestDiscoveryAndDuplicateIdentity(t *testing.T) {
	dir := t.TempDir()
	input, err := os.ReadFile("../examples/xrd.yaml")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"one.yaml", "two.yaml"} {
		if err := os.WriteFile(filepath.Join(dir, name), input, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := execute("generate", dir); err == nil {
		t.Fatal("duplicate identities silently merged")
	}
	if _, _, err := execute("generate", dir, "--exclude", "two.yaml"); err != nil {
		t.Fatal(err)
	}
}

func TestEmptySavedRenderFails(t *testing.T) {
	file := filepath.Join(t.TempDir(), "rendered.yaml")
	if err := os.WriteFile(file, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if out, _, err := execute("composition", "../examples/platform/go-templating.yaml", "--rendered", file); err == nil || out != "" {
		t.Fatalf("empty render fell back to static success: %q %v", out, err)
	}
}

func TestUnknownFunctionResourcesStayOpaque(t *testing.T) {
	file := filepath.Join(t.TempDir(), "custom.yaml")
	data := `apiVersion: apiextensions.crossplane.io/v1
kind: Composition
metadata: {name: custom}
spec:
  compositeTypeRef: {apiVersion: example.org/v1, kind: XCustom}
  mode: Pipeline
  pipeline:
    - step: custom
      functionRef: {name: arbitrary-function}
      input:
        apiVersion: custom.example.org/v1
        kind: Input
        resources: [alpha, beta]
`
	if err := os.WriteFile(file, []byte(data), 0o600); err != nil {
		t.Fatal(err)
	}
	out, _, err := execute("composition", file, "--format", "json")
	if err != nil || !strings.Contains(out, "alpha") {
		t.Fatalf("unknown input was interpreted instead of documented: %s %v", out, err)
	}
}

func TestBatchSemanticErrorsIdentifySourceDocument(t *testing.T) {
	for _, tc := range []struct {
		name, manifest, detail string
	}{
		{"composition", "apiVersion: apiextensions.crossplane.io/v1\nkind: Composition\nmetadata: {name: broken}\nspec: {}\n", "compositeTypeRef"},
		{"xrd", "apiVersion: apiextensions.crossplane.io/v2\nkind: CompositeResourceDefinition\nmetadata: {name: broken}\nspec:\n  group: example.org\n  names: {kind: XBroken, plural: xbrokens}\n", "no versions"},
		{"xrd-identity", "apiVersion: apiextensions.crossplane.io/v2\nkind: CompositeResourceDefinition\nmetadata: {name: broken}\nspec: {}\n", "XRD requires spec.group and spec.names.kind"},
		{"api-group", "apiVersion: example.org/v1\nkind: Composition\nmetadata: {name: broken}\nspec: {}\n", "expected apiextensions.crossplane.io API"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			valid, err := os.ReadFile("../examples/xrd.yaml")
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "a-valid.yaml"), valid, 0o600); err != nil {
				t.Fatal(err)
			}
			file := filepath.Join(dir, "b-broken.yaml")
			bundle := "apiVersion: v1\nkind: ConfigMap\nmetadata: {name: ignored}\n---\n" + tc.manifest
			if err := os.WriteFile(file, []byte(bundle), 0o600); err != nil {
				t.Fatal(err)
			}
			target := filepath.Join(dir, "reference.md")
			if err := os.WriteFile(target, []byte("existing reference\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			out, _, err := execute("generate", dir, "--output", target)
			if err == nil || out != "" || !strings.Contains(err.Error(), file+" document 2:") || strings.Count(err.Error(), file) != 1 || !strings.Contains(err.Error(), tc.detail) {
				t.Fatalf("semantic error lost source or cause: stdout=%q error=%v", out, err)
			}
			data, err := os.ReadFile(target)
			if err != nil || string(data) != "existing reference\n" {
				t.Fatalf("failed batch changed output: %q %v", data, err)
			}
		})
	}
}
