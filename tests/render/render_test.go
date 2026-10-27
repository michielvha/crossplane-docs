//go:build integration

package render_test

import (
	"context"
	"encoding/json"
	"html"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/michielvha/crossplane-docs/internal/engine"
	"gopkg.in/yaml.v3"
)

func root(t *testing.T) string {
	t.Helper()
	p, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	return p
}

func runtimeOptions(t *testing.T, name string) engine.RenderOptions {
	t.Helper()
	p := root(t)
	binary := os.Getenv("CROSSPLANE_BIN")
	if binary == "" {
		binary = "crossplane"
	}
	return engine.RenderOptions{Binary: binary, XR: filepath.Join(p, "examples/platform/xr.yaml"), Composition: filepath.Join(p, "examples/platform/"+name+".yaml"), Functions: filepath.Join(p, "examples/platform/functions.yaml"), Name: name, Timeout: 2 * time.Minute, CrossplaneVersion: "v2.4.3", CrossplaneImage: engine.DefaultCrossplaneImage}
}

func render(t *testing.T, o engine.RenderOptions) engine.Scenario {
	t.Helper()
	data, err := engine.Render(context.Background(), o)
	if err != nil {
		t.Fatal(err)
	}
	scenario, err := engine.ParseScenario(data, o.Name, o.XR)
	if err != nil {
		t.Fatal(err)
	}
	scenario.Inputs = engine.ScenarioInputs(o, "")
	assertScenarioDocumentation(t, o, scenario)
	return scenario
}

func TestRealFunctionMatrix(t *testing.T) {
	var matrix struct {
		Cases []struct {
			Name      string
			Resources int
		}
	}
	data, err := os.ReadFile("cases.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if err := yaml.Unmarshal(data, &matrix); err != nil {
		t.Fatal(err)
	}
	for _, row := range matrix.Cases {
		t.Run(row.Name, func(t *testing.T) {
			o := runtimeOptions(t, row.Name)
			scenario := render(t, o)
			if len(scenario.Resources) != row.Resources {
				t.Fatalf("got resources: %+v", scenario.Resources)
			}
			for _, r := range scenario.Resources {
				if row.Name != "application-stack" && (r.Kind != "ConfigMap" || r.Manifest["data"] == nil) {
					t.Fatalf("wrong generated resource: %+v", r)
				}
			}
			config := engine.Defaults()
			inputs, err := engine.Load(o.Composition, config)
			if err != nil {
				t.Fatal(err)
			}
			doc, err := engine.Build(inputs, engine.Options{Config: config})
			if err != nil {
				t.Fatal(err)
			}
			doc.Scenarios = append(doc.Scenarios, scenario)
			for _, format := range []string{"markdown", "html", "asciidoc", "json", "yaml"} {
				config.Formatter = format
				out, err := engine.Format(doc, config)
				if err != nil {
					t.Fatal(err)
				}
				if !strings.Contains(string(out), "eu-west-1") || (row.Name != "application-stack" && !strings.Contains(string(out), "ConfigMap")) {
					t.Fatalf("%s formatter lost rendered resource data", format)
				}
			}
		})
	}
}

func TestInvalidFunctionPrograms(t *testing.T) {
	for _, name := range []string{"go-templating", "kcl", "cue", "python", "patch-and-transform", "custom"} {
		t.Run(name, func(t *testing.T) {
			fixture := name
			if name == "custom" {
				fixture = "python"
			}
			o := runtimeOptions(t, fixture)
			data, err := os.ReadFile(o.Composition)
			if err != nil {
				t.Fatal(err)
			}
			var m engine.Object
			if err := yaml.Unmarshal(data, &m); err != nil {
				t.Fatal(err)
			}
			steps := engine.Map(m["spec"])["pipeline"].([]interface{})
			input := engine.Map(engine.Map(steps[0])["input"])
			switch name {
			case "go-templating":
				input["inline"] = engine.Object{"template": "{{ missingFunction }}"}
			case "kcl":
				engine.Map(input["spec"])["source"] = "invalid !!!"
			case "cue":
				input["script"] = "invalid !!!"
			case "custom":
				input["apiVersion"] = "custom.platform.example.org/v1"
				input["script"] = "def compose(req, rsp):\n    raise Exception('custom fixture failure')\n"
			case "python":
				input["script"] = "def compose(req, rsp):\n    raise Exception('fixture failure')\n"
			case "patch-and-transform":
				input["resources"] = "invalid"
			}
			data, err = yaml.Marshal(m)
			if err != nil {
				t.Fatal(err)
			}
			file := filepath.Join(t.TempDir(), "bad.yaml")
			if err := os.WriteFile(file, data, 0o600); err != nil {
				t.Fatal(err)
			}
			o.Composition = file
			if _, err := engine.Render(context.Background(), o); err == nil {
				t.Fatal("invalid program was accepted")
			}
			output := filepath.Join(t.TempDir(), "reference.md")
			if err := os.WriteFile(output, []byte("keep existing docs"), 0o600); err != nil {
				t.Fatal(err)
			}
			cmd := exec.Command(reviewBinary(t), "composition", o.Composition, "--render", "--xr", o.XR, "--functions", o.Functions, "--crossplane", o.Binary, "-o", output)
			if bytes, err := cmd.CombinedOutput(); err == nil {
				t.Fatalf("failing runtime accepted by CLI: %s", bytes)
			}
			existing, err := os.ReadFile(output)
			if err != nil || string(existing) != "keep existing docs" {
				t.Fatal("failed render replaced existing output")
			}
		})
	}
}

func TestDistributedCLI(t *testing.T) {
	binary := filepath.Join(t.TempDir(), "crossplane-docs")
	build := exec.Command("go", "build", "-o", binary, ".")
	build.Dir = root(t)
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build: %s: %v", out, err)
	}
	o := runtimeOptions(t, "go-templating")
	cmd := exec.Command(binary, "composition", o.Composition, "--render", "--xr", o.XR, "--functions", o.Functions, "--crossplane", o.Binary, "--format", "json")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("CLI: %s: %v", out, err)
	}
	var doc engine.Document
	if err := json.Unmarshal(out, &doc); err != nil {
		t.Fatal(err)
	}
	if len(doc.Scenarios) != 1 || len(doc.Scenarios[0].Resources) != 2 {
		t.Fatalf("CLI lost render output: %s", out)
	}
}

func compositionObject(t *testing.T, o engine.RenderOptions) engine.Object {
	t.Helper()
	data, err := os.ReadFile(o.Composition)
	if err != nil {
		t.Fatal(err)
	}
	var m engine.Object
	if err := yaml.Unmarshal(data, &m); err != nil {
		t.Fatal(err)
	}
	return m
}

func pipelineInput(m engine.Object) engine.Object {
	return engine.Map(engine.Map(engine.Map(m["spec"])["pipeline"].([]interface{})[0])["input"])
}

func replaceComposition(t *testing.T, o engine.RenderOptions, m engine.Object) engine.RenderOptions {
	t.Helper()
	data, err := yaml.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	o.Composition = filepath.Join(t.TempDir(), "composition.yaml")
	if err := os.WriteFile(o.Composition, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return o
}

func TestGoTemplateSourceVariants(t *testing.T) {
	base := runtimeOptions(t, "go-templating")
	object := compositionObject(t, base)
	original := pipelineInput(object)
	source := engine.Map(original["inline"])["template"].(string)
	for _, variant := range []string{"template-list", "environment", "condition", "extra-resources"} {
		t.Run(variant, func(t *testing.T) {
			o := base
			m := compositionObject(t, o)
			input := pipelineInput(m)
			switch variant {
			case "template-list":
				input["inline"] = engine.Object{"templates": []string{source}}
			case "environment":
				delete(input, "inline")
				input["source"] = "Environment"
				input["environment"] = engine.Object{"key": "templates"}
				contextJSON, _ := json.Marshal(engine.Object{"templates": source})
				o.Context = []string{"apiextensions.crossplane.io/environment=" + string(contextJSON)}
			case "condition":
				input["inline"] = engine.Object{"template": "{{ if .observed.composite.resource.spec.storage.backup }}\n" + source + "\n{{ end }}"}
			case "extra-resources":
				input["inline"] = engine.Object{"template": source + `---
apiVersion: meta.gotemplating.fn.crossplane.io/v1alpha1
kind: ExtraResources
requirements:
  defaults:
    apiVersion: v1
    kind: ConfigMap
    matchName: platform-defaults
    namespace: default
{{- if .requiredResources }}
{{- $found := index .requiredResources "defaults" }}
{{- if $found.items }}
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: required-consumer
  annotations:
    {{ setResourceNameAnnotation "required-consumer" }}
data:
  environment: {{ (index $found.items 0).resource.data.environment | quote }}
{{- end }}
{{- end }}
`}
				o.Required = []string{filepath.Join(root(t), "examples/platform/required.yaml")}
			}
			o = replaceComposition(t, o, m)
			s := render(t, o)
			want := 2
			if variant == "extra-resources" {
				want = 3
				found := false
				for _, r := range s.Resources {
					if r.Name == "required-consumer" && engine.String(engine.Map(r.Manifest["data"])["environment"]) == "integration" {
						found = true
					}
				}
				if !found {
					t.Fatalf("required resource value was not consumed: %+v", s.Resources)
				}
			}
			if variant == "condition" {
				want = 0
			}
			if len(s.Resources) != want {
				t.Fatalf("%s: %+v", variant, s.Resources)
			}
		})
	}
}

func TestMixedPipelineAndObservedState(t *testing.T) {
	o := runtimeOptions(t, "mixed")
	o.Required = []string{filepath.Join(root(t), "examples/platform/required.yaml")}
	for _, observed := range []bool{false, true} {
		t.Run(map[bool]string{false: "create", true: "update"}[observed], func(t *testing.T) {
			opt := o
			if observed {
				opt.Observed = filepath.Join(root(t), "examples/platform/observed.yaml")
			}
			s := render(t, opt)
			want := 3
			if observed {
				want = 4
			}
			if len(s.Resources) != want {
				t.Fatalf("mixed resources: %+v", s.Resources)
			}
			found := false
			for _, r := range s.Resources {
				if r.Name == "network-0" {
					found = true
					if engine.String(engine.Map(r.Manifest["data"])["environment"]) != "integration" {
						t.Fatal("environment not propagated")
					}
				}
			}
			if !found {
				t.Fatal("template resources missing")
			}
			status := engine.Map(s.Composite["status"])
			wantPhase := "Pending"
			if observed {
				wantPhase = "Ready"
			}
			if engine.String(status["phase"]) != wantPhase || engine.String(status["resourceCount"]) != "2" {
				t.Fatalf("mixed composite status missing: %+v", status)
			}
			ready := false
			for _, condition := range status["conditions"].([]interface{}) {
				m := engine.Map(condition)
				if engine.String(m["type"]) == "Ready" && engine.String(m["status"]) == "True" {
					ready = true
				}
			}
			if ready != observed {
				t.Fatalf("mixed readiness differs from observed state: %+v", status)
			}
			if observed && engine.String(engine.Map(s.Composite["status"])["observedRegion"]) != "eu-central-1" {
				t.Fatalf("observed patch missing: %+v", s.Composite)
			}
		})
	}
}

func TestUnknownFunctionInput(t *testing.T) {
	o := runtimeOptions(t, "python")
	m := compositionObject(t, o)
	pipelineInput(m)["apiVersion"] = "custom.platform.example.org/v1"
	pipelineInput(m)["resources"] = []interface{}{"alpha", "beta"}
	o = replaceComposition(t, o, m)
	s := render(t, o)
	if len(s.Resources) != 2 {
		t.Fatal("generic renderer lost unknown-function resources")
	}
	doc, err := engine.Build([]engine.Input{{Object: m}}, engine.Options{Config: engine.Defaults()})
	if err != nil {
		t.Fatal(err)
	}
	if doc.Compositions[0].Steps[0].Adapter != "custom" {
		t.Fatal("fixture was not an unknown source adapter")
	}
}

func TestFileSystemSources(t *testing.T) {
	for _, target := range []string{"go-files", "kcl-files"} {
		docker(t, "build", "--target", target, "-t", "crossplane-docs-"+target+":integration", filepath.Join(root(t), "examples/platform/runtime"))
	}
	for _, name := range []string{"go-templating", "kcl"} {
		t.Run(name, func(t *testing.T) {
			o := runtimeOptions(t, name)
			m := compositionObject(t, o)
			input := pipelineInput(m)
			image := "crossplane-docs-go-files:integration"
			fn := "function-go-templating"
			if name == "go-templating" {
				delete(input, "inline")
				input["source"] = "FileSystem"
				input["fileSystem"] = engine.Object{"dirPath": "/templates"}
			} else {
				image = "crossplane-docs-kcl-files:integration"
				fn = "function-kcl"
				engine.Map(input["spec"])["source"] = "/templates/main.k"
			}
			o = replaceComposition(t, o, m)
			functions, err := engine.Load(o.Functions, engine.Defaults())
			if err != nil {
				t.Fatal(err)
			}
			var buf strings.Builder
			for _, f := range functions {
				if engine.String(engine.Map(f.Object["metadata"])["name"]) == fn {
					engine.Map(f.Object["spec"])["package"] = image
					engine.Map(f.Object["metadata"])["annotations"] = engine.Object{"render.crossplane.io/runtime-docker-pull-policy": "Never"}
				}
				data, err := yaml.Marshal(f.Object)
				if err != nil {
					t.Fatal(err)
				}
				buf.WriteString("---\n")
				buf.Write(data)
			}
			o.Functions = filepath.Join(t.TempDir(), "functions.yaml")
			if err := os.WriteFile(o.Functions, []byte(buf.String()), 0o600); err != nil {
				t.Fatal(err)
			}
			s := render(t, o)
			if len(s.Resources) != 2 {
				t.Fatalf("filesystem output: %+v", s.Resources)
			}
		})
	}
}

func TestRealSchemaValidation(t *testing.T) {
	o := runtimeOptions(t, "go-templating")
	data, err := os.ReadFile(o.XR)
	if err != nil {
		t.Fatal(err)
	}
	extensions := filepath.Join(root(t), "examples/platform.yaml")
	if err := engine.Validate(context.Background(), o, extensions, data); err != nil {
		t.Fatal(err)
	}
	nested := compositionObject(t, o)
	pipelineInput(nested)["inline"] = engine.Object{"template": `apiVersion: platform.example.org/v1beta1
kind: XPlatform
metadata:
  name: nested-platform
  annotations:
    {{ setResourceNameAnnotation "nested-platform" }}
spec:
  region: eu-west-1
  minReplicas: 0
  maxReplicas: 1
  networks:
    - name: nested
      cidr: 10.44.0.0/16
`}
	validationRuntime := replaceComposition(t, o, nested)
	rendered, err := engine.Render(context.Background(), validationRuntime)
	if err != nil {
		t.Fatal(err)
	}
	if err := engine.Validate(context.Background(), o, extensions, rendered); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"field-type", "cel", "pattern", "unknown-field", "missing-schema"} {
		t.Run(name, func(t *testing.T) {
			var xr engine.Object
			if err := yaml.Unmarshal(data, &xr); err != nil {
				t.Fatal(err)
			}
			spec := engine.Map(xr["spec"])
			switch name {
			case "field-type":
				spec["region"] = 42
			case "cel":
				spec["minReplicas"] = 20
				spec["maxReplicas"] = 1
			case "pattern":
				engine.Map(spec["networks"].([]interface{})[0])["name"] = "INVALID NAME"
			case "unknown-field":
				spec["typo"] = true
			case "missing-schema":
				xr["apiVersion"] = "missing.example.org/v1"
			}
			bad, err := yaml.Marshal(xr)
			if err != nil {
				t.Fatal(err)
			}
			if err := engine.Validate(context.Background(), o, extensions, bad); err == nil {
				t.Fatal("invalid scenario passed validation")
			}
		})
	}
}

func TestKCLTargets(t *testing.T) {
	for _, target := range []string{"Default", "Resources", "XR", "PatchDesired", "PatchResources"} {
		t.Run(target, func(t *testing.T) {
			o := runtimeOptions(t, "kcl")
			m := compositionObject(t, o)
			spec := engine.Map(pipelineInput(m)["spec"])
			spec["target"] = target
			switch target {
			case "XR":
				spec["source"] = "dxr = option(\"params\").dxr\nitems = [{**dxr, status.phase = \"KCLReady\"}]\n"
			case "PatchDesired":
				spec["source"] = "items = [{metadata.annotations = {\"krm.kcl.dev/composition-resource-name\" = \"summary\"}, data.patched = \"true\"}]\n"
				patch := compositionObject(t, runtimeOptions(t, "patch-and-transform"))
				producer := engine.Map(patch["spec"])["pipeline"].([]interface{})[0]
				pipeline := engine.Map(m["spec"])["pipeline"].([]interface{})
				engine.Map(m["spec"])["pipeline"] = append([]interface{}{producer}, pipeline...)
			case "PatchResources":
				spec["resources"] = []interface{}{engine.Object{"name": "fixture", "base": engine.Object{"apiVersion": "v1", "kind": "ConfigMap", "metadata": engine.Object{"name": "kcl-patched"}, "data": engine.Object{"region": "eu-west-1"}}}}
				spec["source"] = "items = [{metadata.annotations = {\"krm.kcl.dev/composition-resource-name\" = \"fixture\"}, data.patched = \"true\"}]\n"
			}
			o = replaceComposition(t, o, m)
			s := render(t, o)
			switch target {
			case "Default", "Resources":
				if len(s.Resources) != 2 {
					t.Fatalf("resources lost: %+v", s.Resources)
				}
			case "XR":
				if engine.String(engine.Map(s.Composite["status"])["phase"]) != "KCLReady" {
					t.Fatalf("XR result lost: %+v", s.Composite)
				}
			case "PatchDesired", "PatchResources":
				if len(s.Resources) > 0 && engine.String(engine.Map(s.Resources[0].Manifest["data"])["patched"]) != "true" {
					t.Fatalf("patch not applied: %+v", s.Resources)
				}
				if len(s.Resources) != 1 {
					t.Fatalf("patch resources lost: %+v", s.Resources)
				}
			}
		})
	}
}

func TestKCLAndPythonMockInputs(t *testing.T) {
	for _, name := range []string{"kcl", "python"} {
		t.Run(name, func(t *testing.T) {
			o := runtimeOptions(t, name)
			o.Observed = filepath.Join(root(t), "examples/platform/observed.yaml")
			o.Context = []string{`platform={"owner":"platform-team"}`}
			m := compositionObject(t, o)
			input := pipelineInput(m)
			if name == "kcl" {
				spec := engine.Map(input["spec"])
				spec["params"] = engine.Object{"owner": "platform-team"}
				source := engine.String(spec["source"])
				source = strings.Replace(source, "cidr = n.cidr", "cidr = n.cidr, owner = option(\"params\").owner, observedCount = str(len(option(\"params\").ocds))", 1)
				spec["source"] = source
			} else {
				source := engine.String(input["script"])
				source = strings.Replace(source, `"cidr": network["cidr"]`, `"cidr": network["cidr"], "owner": req.context["platform"]["owner"], "observedCount": str(len(req.observed.resources))`, 1)
				input["script"] = source
			}
			o = replaceComposition(t, o, m)
			s := render(t, o)
			if len(s.Resources) != 2 {
				t.Fatal("mock-input resources missing")
			}
			for _, r := range s.Resources {
				data := engine.Map(r.Manifest["data"])
				if engine.String(data["owner"]) != "platform-team" || engine.String(data["observedCount"]) != "3" {
					t.Fatalf("mock/parameter data missing: %+v", data)
				}
			}
		})
	}
}

func assertScenarioDocumentation(t *testing.T, o engine.RenderOptions, s engine.Scenario) {
	t.Helper()
	c := engine.Defaults()
	inputs, err := engine.Load(o.Composition, c)
	if err != nil {
		t.Fatal(err)
	}
	functions, err := engine.Load(o.Functions, c)
	if err != nil {
		t.Fatal(err)
	}
	inputs = append(inputs, functions...)
	doc, err := engine.Build(inputs, engine.Options{Config: c})
	if err != nil {
		t.Fatal(err)
	}
	doc.Scenarios = append(doc.Scenarios, s)
	for _, format := range []string{"markdown", "html", "asciidoc", "json", "yaml"} {
		c.Formatter = format
		data, err := engine.Format(doc, c)
		if err != nil {
			t.Fatal(err)
		}
		for _, r := range s.Resources {
			if !strings.Contains(string(data), r.Name) {
				t.Fatalf("%s lost rendered identity %s", format, r.Name)
			}
			for key, value := range engine.Map(r.Manifest["data"]) {
				if scalar, ok := value.(string); ok && scalar != "" && !strings.Contains(html.UnescapeString(string(data)), scalar) {
					t.Fatalf("%s lost rendered %s.%s = %q", format, r.Name, key, scalar)
				}
			}
		}
		if !strings.Contains(string(data), s.Name) || !strings.Contains(string(data), "live render") {
			t.Fatalf("%s lost scenario provenance", format)
		}
		if format == "json" {
			var actual engine.Document
			if err := json.Unmarshal(data, &actual); err != nil {
				t.Fatal(err)
			}
			if len(actual.Scenarios) != 1 {
				t.Fatal("scenario omitted")
			}
			want, _ := json.Marshal(s)
			got, _ := json.Marshal(actual.Scenarios[0])
			if string(want) != string(got) {
				t.Fatalf("structured output changed scenario facts:\n%s\n%s", want, got)
			}
		}
	}
}

var (
	reviewBuildOnce sync.Once
	reviewBuildPath string
)

func reviewBinary(t *testing.T) string {
	t.Helper()
	reviewBuildOnce.Do(func() {
		dir, err := os.MkdirTemp("", "crossplane-docs-review-cli-*")
		if err != nil {
			t.Fatal(err)
		}
		reviewBuildPath = filepath.Join(dir, "crossplane-docs")
		command := exec.Command("go", "build", "-o", reviewBuildPath, ".")
		command.Dir = root(t)
		if out, err := command.CombinedOutput(); err != nil {
			t.Fatalf("build: %s: %v", out, err)
		}
	})
	return reviewBuildPath
}

func TestCheckModeBooleanExitCodes(t *testing.T) {
	for _, flag := range []string{"--check", "--check=true", "--check=TRUE", "--check=1"} {
		t.Run(flag, func(t *testing.T) {
			command := exec.Command(reviewBinary(t), "xrd", filepath.Join(t.TempDir(), "missing.yaml"), flag, "-o", filepath.Join(t.TempDir(), "docs.md"))
			if err := command.Run(); err == nil {
				t.Fatal("missing input succeeded")
			} else if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 2 {
				t.Fatalf("wrong check failure exit: %v", err)
			}
		})
	}
}

func TestRequiredSchemasAreConsumed(t *testing.T) {
	o := runtimeOptions(t, "python")
	m := compositionObject(t, o)
	pipelineInput(m)["script"] = `from crossplane.function.proto.v1 import run_function_pb2 as fnv1

def compose(req, rsp):
    rsp.requirements.schemas["platform"].api_version = "platform.example.org/v1beta1"
    rsp.requirements.schemas["platform"].kind = "XPlatform"
    if "platform" not in req.required_schemas:
        return
    schema = req.required_schemas["platform"].openapi_v3
    rsp.desired.resources["schema-consumer"].resource.update({
        "apiVersion": "v1", "kind": "ConfigMap",
        "metadata": {"name": "schema-consumer"},
        "data": {"description": schema["description"]},
    })
    rsp.desired.resources["schema-consumer"].ready = fnv1.READY_TRUE
`
	o = replaceComposition(t, o, m)
	o.Schemas = t.TempDir()
	schema := engine.Object{"components": engine.Object{"schemas": engine.Object{"platform": engine.Object{
		"type": "object", "description": "received-platform-schema",
		"x-kubernetes-group-version-kind": []engine.Object{{"group": "platform.example.org", "version": "v1beta1", "kind": "XPlatform"}},
	}}}}
	data, err := json.Marshal(schema)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(o.Schemas, "platform.json"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	scenario := render(t, o)
	if len(scenario.Resources) != 1 || engine.String(engine.Map(scenario.Resources[0].Manifest["data"])["description"]) != "received-platform-schema" {
		t.Fatalf("required schema was not consumed: %+v", scenario.Resources)
	}
}
