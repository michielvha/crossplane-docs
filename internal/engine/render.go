package engine

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

const DefaultCrossplaneImage = "xpkg.crossplane.io/crossplane/crossplane:v2.4.3@sha256:1c66261d805a73cade46e2ac94b665c73aebd4a00844de5fcae4bcadaa2be638"

type RenderOptions struct {
	Binary            string
	XR                string
	Composition       string
	Functions         string
	Observed          string
	Required          []string
	Schemas           string
	Context           []string
	Name              string
	Timeout           time.Duration
	CrossplaneVersion string
	CrossplaneImage   string
}

func (o RenderOptions) binary() string {
	if o.Binary != "" {
		return o.Binary
	}
	return "crossplane"
}

func (o RenderOptions) run(ctx context.Context, args ...string) ([]byte, error) {
	duration := o.Timeout
	if duration <= 0 {
		duration = 2 * time.Minute
	}
	ctx, cancel := context.WithTimeout(ctx, duration)
	defer cancel()
	command := exec.CommandContext(ctx, o.binary(), args...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	out, err := command.Output()
	if ctx.Err() != nil {
		return nil, fmt.Errorf("crossplane execution timed out or cancelled: %w", ctx.Err())
	}
	if err != nil {
		return nil, fmt.Errorf("crossplane command failed: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

func Render(ctx context.Context, o RenderOptions) ([]byte, error) {
	if o.XR == "" || o.Composition == "" || o.Functions == "" {
		return nil, fmt.Errorf("render requires --xr, a Composition file and --functions")
	}
	functions, err := neededFunctions(o.Composition, o.Functions)
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.Remove(functions) }()
	args := []string{"composition", "render", o.XR, o.Composition, functions, "--include-full-xr", "--include-context", "--include-function-results"}
	if o.CrossplaneImage != "" {
		args = append(args, "--crossplane-image="+o.CrossplaneImage)
	} else if o.CrossplaneVersion != "" {
		args = append(args, "--crossplane-version="+o.CrossplaneVersion)
	}
	if o.Observed != "" {
		args = append(args, "--observed-resources="+o.Observed)
	}
	for _, f := range o.Required {
		args = append(args, "--required-resources="+f)
	}
	if o.Schemas != "" {
		args = append(args, "--required-schemas="+o.Schemas)
	}
	for _, value := range o.Context {
		args = append(args, "--context-values="+value)
	}
	return o.run(ctx, args...)
}

func Validate(ctx context.Context, o RenderOptions, extensions string, resources []byte) error {
	if extensions == "" {
		return fmt.Errorf("validation requires --extensions with XRD/CRD schemas")
	}
	inputs, err := Read(resources, "validation input")
	if err != nil {
		return err
	}
	var payload bytes.Buffer
	for _, in := range inputs {
		api, kind := text(in.Object, "apiVersion"), text(in.Object, "kind")
		if strings.HasPrefix(api, "render.crossplane.io/") && (kind == "Context" || kind == "Result") {
			continue
		}
		data, err := yaml.Marshal(in.Object)
		if err != nil {
			return err
		}
		payload.WriteString("---\n")
		payload.Write(data)
	}
	if payload.Len() == 0 {
		return fmt.Errorf("no Kubernetes resources to validate")
	}
	resources = payload.Bytes()
	f, err := os.CreateTemp("", "crossplane-docs-validate-*.yaml")
	if err != nil {
		return err
	}
	defer func() { _ = os.Remove(f.Name()) }()
	if _, err := f.Write(resources); err != nil {
		_ = f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	args := []string{"resource", "validate", extensions, f.Name(), "--error-on-missing-schemas"}
	if o.CrossplaneImage != "" {
		args = append(args, "--crossplane-image="+o.CrossplaneImage)
	} else if o.CrossplaneVersion != "" {
		args = append(args, "--crossplane-image=xpkg.crossplane.io/crossplane/crossplane:"+o.CrossplaneVersion)
	}
	_, err = o.run(ctx, args...)
	return err
}

func ParseScenario(data []byte, name, xr string, refs ...Composition) (Scenario, error) {
	inputs, err := Read(data, "rendered output")
	s := Scenario{Name: name, XR: xr, Resources: []Resource{}}
	if err != nil {
		return s, err
	}
	if len(inputs) == 0 {
		return s, fmt.Errorf("renderer produced no YAML documents")
	}
	isComposite := func(m Object) bool {
		if text(m, "metadata", "annotations", "crossplane.io/composition-resource-name") != "" {
			return false
		}
		for _, ref := range refs {
			if text(m, "apiVersion") == ref.APIVersion && text(m, "kind") == ref.Kind {
				return true
			}
		}
		return false
	}
	var xrObject Object
	if xr != "" {
		b, err := os.ReadFile(xr)
		if err != nil {
			return s, err
		}
		v, err := Read(b, xr)
		if err != nil {
			return s, err
		}
		if len(v) != 1 {
			return s, fmt.Errorf("sample XR requires one document")
		}
		xrObject = v[0].Object
	}
	for _, in := range inputs {
		m := in.Object
		kind := text(m, "kind")
		api := text(m, "apiVersion")
		switch {
		case kind == "Context" && strings.HasPrefix(api, "render.crossplane.io/"):
			s.Context = append(s.Context, Map(redacted(m)))
		case kind == "Result" && strings.HasPrefix(api, "render.crossplane.io/"):
			s.Results = append(s.Results, Map(redacted(m)))
		case (xrObject != nil && api == text(xrObject, "apiVersion") && kind == text(xrObject, "kind") && text(m, "metadata", "name") == text(xrObject, "metadata", "name")) || (xrObject == nil && isComposite(m)):
			s.Composite = Map(redacted(m))
		default:
			if api == "" || kind == "" {
				return s, fmt.Errorf("rendered resource requires apiVersion and kind")
			}
			name := text(m, "metadata", "annotations", "crossplane.io/composition-resource-name")
			if name == "" {
				name = text(m, "metadata", "name")
			}
			if name == "" {
				name = text(m, "metadata", "generateName")
			}
			s.Resources = append(s.Resources, Resource{Name: name, Kind: kind, APIVersion: api, Provenance: "rendered", Manifest: Map(redacted(m))})
		}
	}
	sort.SliceStable(s.Resources, func(i, j int) bool { return s.Resources[i].Name < s.Resources[j].Name })
	return s, nil
}

// The upstream CLI starts every Function in the supplied file, including unused
// runtimes. Restrict it to this pipeline to avoid unrelated downloads/failures.
func neededFunctions(composition, functions string) (string, error) {
	comps, err := Load(composition, Defaults())
	if err != nil {
		return "", err
	}
	if len(comps) != 1 {
		return "", fmt.Errorf("render requires one Composition")
	}
	needed := map[string]bool{}
	for _, v := range List(at(comps[0].Object, "spec", "pipeline")) {
		needed[text(Map(v), "functionRef", "name")] = true
	}
	inputs, err := Load(functions, Defaults())
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	for _, in := range inputs {
		name := text(in.Object, "metadata", "name")
		if needed[name] && text(in.Object, "kind") == "Function" {
			data, err := yaml.Marshal(in.Object)
			if err != nil {
				return "", err
			}
			buf.WriteString("---\n")
			buf.Write(data)
			delete(needed, name)
		}
	}
	if len(needed) > 0 {
		return "", fmt.Errorf("missing Function definitions: %v", needed)
	}
	f, err := os.CreateTemp("", "crossplane-docs-functions-*.yaml")
	if err != nil {
		return "", err
	}
	if _, err := f.Write(buf.Bytes()); err != nil {
		_ = f.Close()
		_ = os.Remove(f.Name())
		return "", err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(f.Name())
		return "", err
	}
	return f.Name(), nil
}

// ScenarioInputs records supplied inputs without embedding credential-bearing files.
func ScenarioInputs(o RenderOptions, saved string) Object {
	inputs := Object{"xr": o.XR}
	if saved != "" {
		inputs["mode"] = "saved output"
		inputs["renderedFile"] = saved
	} else {
		inputs["mode"] = "live render"
		inputs["composition"] = o.Composition
		inputs["functions"] = o.Functions
	}
	if o.Observed != "" {
		inputs["observedResources"] = o.Observed
	}
	if len(o.Required) > 0 {
		inputs["requiredResources"] = o.Required
	}
	if o.Schemas != "" {
		inputs["requiredSchemas"] = o.Schemas
	}
	contextValues := Object{}
	for _, entry := range o.Context {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			continue
		}
		var value interface{}
		if err := json.Unmarshal([]byte(parts[1]), &value); err == nil {
			contextValues[parts[0]] = redacted(value)
		}
	}
	if len(contextValues) > 0 {
		inputs["contextValues"] = redacted(contextValues)
	}
	return inputs
}
