package engine

import (
	"encoding/json"
	"html"
	"regexp"
	"strings"
	"testing"
)

func TestReferenceExamplesHaveOneManifestAndCompleteInventory(t *testing.T) {
	doc := Document{Title: "Examples", APIs: []API{{Group: "example.org", Version: "v1", Kind: "XService", ClaimKind: "Service"}}, Examples: []Object{
		{"apiVersion": "example.org/v1", "kind": "XService", "metadata": Object{"name": "service"}, "spec": Object{"value": "composite-payload-marker"}},
		{"apiVersion": "example.org/v1", "kind": "Service", "metadata": Object{"name": "claim"}, "spec": Object{"value": "claim-payload-marker"}},
		{"apiVersion": "other.org/v1", "kind": "Other", "spec": Object{"value": "unmatched-payload-marker"}},
	}}
	before, _ := json.Marshal(doc)
	for _, layout := range []string{"reference", "explorer", "compact"} {
		for _, sections := range [][]string{{"apis", "examples"}, {"examples", "apis"}, {"examples"}} {
			c := htmlConfig(t, layout)
			c.Sections = sections
			out, err := Format(doc, c)
			if err != nil {
				t.Fatal(err)
			}
			for _, marker := range []string{"composite-payload-marker", "claim-payload-marker", "unmatched-payload-marker"} {
				if strings.Count(string(out), marker) != 1 {
					t.Errorf("%s sections=%v repeated or lost manifest %s", layout, sections, marker)
				}
			}
			links := strings.Count(string(out), `href="#api-0-example-`)
			if layout == "reference" && len(sections) > 1 {
				if links != 2 || !strings.Contains(string(out), `id="api-0-example-0"`) || !strings.Contains(string(out), `id="api-0-example-1"`) {
					t.Errorf("matched composite/claim inventory links are incomplete: %s %v", layout, sections)
				}
			} else if links != 0 {
				t.Errorf("inventory links to manifests without paired examples: %s %v", layout, sections)
			}
			after, _ := json.Marshal(doc)
			if string(before) != string(after) {
				t.Fatal("HTML example deduplication changed the shared model")
			}
		}
	}
}

func htmlConfig(t *testing.T, layout string) Config {
	t.Helper()
	path := writeFixture(t, t.TempDir(), "config.yml", "version: 1\nformatter: html\nhtml-layout: "+layout+"\n")
	c, err := LoadConfig(path, path)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestHTMLLayoutConfiguration(t *testing.T) {
	for _, layout := range []string{"reference", "explorer", "compact"} {
		t.Run(layout, func(t *testing.T) {
			c := htmlConfig(t, layout)
			out, err := Format(model(t), c)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(string(out), `data-html-layout="`+layout+`"`) || strings.Contains(string(out), `class="metrics"`) {
				t.Fatal("layout was not selected or dashboard metrics remain")
			}
		})
	}
	c := htmlConfig(t, "unknown")
	if err := c.Validate(); err == nil || !strings.Contains(err.Error(), "html-layout") {
		t.Fatalf("unknown HTML layout accepted: %v", err)
	}
	c = Defaults()
	c.Formatter = "html"
	out, err := Format(model(t), c)
	if err != nil || !strings.Contains(string(out), `data-html-layout="reference"`) {
		t.Fatalf("default reference layout missing: %v", err)
	}
}

func TestHTMLSourceProgramsStayReadable(t *testing.T) {
	program := "{{ range .spec.networks }}\nkind: ConfigMap\r\n\t# Unicode λ; literal \\n stays literal\n{{ end }}\n</script><img src=x onerror=alert(1)>\n"
	doc := model(t)
	doc.Compositions = []Composition{{Name: "source", Steps: []Step{{Name: "template", Adapter: "go-templating", Input: Object{"apiVersion": "gotemplating.fn.crossplane.io/v1beta1", "kind": "GoTemplate", "inline": Object{"template": program}}}}}}
	for _, layout := range []string{"reference", "explorer", "compact"} {
		c := htmlConfig(t, layout)
		out, err := Format(doc, c)
		if err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(html.UnescapeString(string(out)), program) || !strings.Contains(string(out), "source-code") {
			t.Fatalf("%s serialized the embedded program instead of displaying it", layout)
		}
		if strings.Contains(string(out), "<img src=x") || strings.Contains(string(out), "</script><img") {
			t.Fatalf("%s allowed source HTML injection", layout)
		}
	}
}

func TestHTMLHeadingAndFieldIDNamespaces(t *testing.T) {
	doc := Document{Title: "Collision", APIs: []API{{Kind: "XExample", Group: "example.org", Version: "v1", Fields: []Field{{Path: "spec.alpha", Type: "string"}, {Path: "spec.beta", Type: "string"}}}}, Compositions: []Composition{{Name: "api-0-field"}}}
	ids := regexp.MustCompile(`\bid="([^"]+)"`)
	for _, layout := range []string{"reference", "explorer", "compact"} {
		out, err := Format(doc, htmlConfig(t, layout))
		if err != nil {
			t.Fatal(err)
		}
		seen := map[string]bool{}
		for _, id := range ids.FindAllStringSubmatch(string(out), -1) {
			if seen[id[1]] {
				t.Fatalf("%s duplicates field/heading anchor %s", layout, id[1])
			}
			seen[id[1]] = true
		}
	}
}

func TestHTMLModelCoverageAndDeterminism(t *testing.T) {
	doc := model(t)
	doc.APIs = append(doc.APIs, API{Name: "other", Group: "other.example.org", Kind: doc.APIs[0].Kind, Version: doc.APIs[0].Version, Fields: []Field{{Path: "spec.region", Type: "string", Description: "other API field", Reference: "#/definitions/region"}}})
	doc.Examples = []Object{{"apiVersion": "platform.example.org/v1beta1", "kind": "XPlatform", "spec": Object{"enabled": false, "count": 0}}}
	doc.Compositions = []Composition{{Name: "platform-pipeline", Resources: []Resource{{Name: "database", Patches: []Patch{{Type: "FromCompositeFieldPath", From: "spec.region", To: "spec.forProvider.region", Description: "copy region"}}, Readiness: Object{"type": "None"}, Connections: Object{"name": "endpoint"}}}}}
	doc.Scenarios = []Scenario{{Name: "create-example", Inputs: Object{"mode": "saved output"}, Resources: []Resource{{Name: "rendered-database", Kind: "ConfigMap", Manifest: Object{"data": Object{"region": "eu-west-1"}}}}, Composite: Object{"status": Object{"phase": "Pending"}}, Context: []Object{{"context-marker": "visible"}}, Results: []Object{{"result-marker": "visible"}}}}
	ids := regexp.MustCompile(`\bid="([^"]+)"`)
	for _, layout := range []string{"reference", "explorer", "compact"} {
		c := htmlConfig(t, layout)
		out, err := Format(doc, c)
		if err != nil {
			t.Fatal(err)
		}
		again, _ := Format(doc, c)
		if string(out) != string(again) {
			t.Fatal("HTML generation is not deterministic")
		}
		for _, want := range []string{"other API field", "#/definitions/region", "Printer columns", "minReplicas", "oneOf", "platform-pipeline", "copy region", "Readiness", "Connection details", "create-example", "rendered-database", "Pending", "context-marker", "result-marker", "enabled: false", "count: 0"} {
			if !strings.Contains(html.UnescapeString(string(out)), want) {
				t.Errorf("%s lost %s", layout, want)
			}
		}
		seen := map[string]bool{}
		for _, match := range ids.FindAllStringSubmatch(string(out), -1) {
			if seen[match[1]] {
				t.Errorf("duplicate HTML ID %q", match[1])
			}
			seen[match[1]] = true
		}
		c.Sections = []string{"apis"}
		c.Columns = []string{"path", "type"}
		c.Settings.Descriptions = false
		filtered, err := Format(doc, c)
		if err != nil {
			t.Fatal(err)
		}
		for _, absent := range []string{"other API field", "platform-pipeline", "create-example", "Example manifest", "minReplicas must not exceed maxReplicas"} {
			if strings.Contains(string(filtered), absent) {
				t.Errorf("%s ignored presentation filter %s", layout, absent)
			}
		}
	}
}

func TestHTMLCustomTemplatePrecedence(t *testing.T) {
	c := htmlConfig(t, "explorer")
	c.Template = writeFixture(t, t.TempDir(), "custom.html", "<h1>{{ .Title }}</h1>")
	out, err := Format(model(t), c)
	if err != nil || !strings.HasPrefix(string(out), "<h1>") || strings.Contains(string(out), "data-html-layout") {
		t.Fatalf("custom template lost precedence: %v", err)
	}
	c.Formatter = "json"
	c.Template = ""
	out, err = Format(model(t), c)
	if err != nil || strings.Contains(string(out), "html-layout") || !strings.Contains(string(out), "schemaVersion") {
		t.Fatalf("HTML layout changed structured output: %v", err)
	}
}
