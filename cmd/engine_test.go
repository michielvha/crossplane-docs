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
