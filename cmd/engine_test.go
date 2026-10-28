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
			if format == "html" && (!strings.Contains(out, "<!doctype html>") || !strings.Contains(out, "<table")) {
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
	out, _, err := execute("xrd", file, "--format", "json")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(out, "spec.parameters.region") {
		t.Fatal("config false ignored")
	}
	out, _, err = execute("xrd", file, "--format", "json", "--show-nested=true")
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
