package render_test

import (
	"os"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestVersioningRequiresRealChecks(t *testing.T) {
	data, err := os.ReadFile("../../.github/workflows/ci.yml")
	if err != nil {
		t.Fatal(err)
	}
	var config map[string]interface{}
	if err := yaml.Unmarshal(data, &config); err != nil {
		t.Fatal(err)
	}
	jobs := config["jobs"].(map[string]interface{})
	version := jobs["version"].(map[string]interface{})
	if version["needs"] != "verify" {
		t.Fatal("tagging does not depend on verification")
	}
	verify := jobs["verify"].(map[string]interface{})
	foundRender, foundFormat := false, false
	for _, step := range verify["steps"].([]interface{}) {
		s := step.(map[string]interface{})
		run, _ := s["run"].(string)
		if strings.Contains(run, "go run ./scripts/verify-render") {
			foundRender = true
			if s["continue-on-error"] == true {
				t.Fatal("render failures ignored")
			}
		}
		if strings.Contains(run, "bun test --cwd tests/format") {
			foundFormat = true
		}
	}
	if !foundRender || !foundFormat {
		t.Fatal("required formatter or real-render check missing")
	}
}
