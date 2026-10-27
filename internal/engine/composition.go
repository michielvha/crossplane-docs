package engine

import (
	"fmt"
	"sort"
	"strings"
)

func extractComposition(m Object, funcs map[string]Object) (Composition, error) {
	spec := Map(m["spec"])
	c := Composition{Name: text(m, "metadata", "name"), APIVersion: text(spec, "compositeTypeRef", "apiVersion"), Kind: text(spec, "compositeTypeRef", "kind"), Mode: text(spec, "mode"), Steps: []Step{}, Resources: []Resource{}, Environment: redacted(spec["environment"])}
	if c.Kind == "" || c.APIVersion == "" {
		return c, fmt.Errorf("Composition requires compositeTypeRef apiVersion and kind")
	}
	if c.Mode == "" {
		c.Mode = "Resources"
	}
	if c.Mode != "Resources" && c.Mode != "Pipeline" {
		return c, fmt.Errorf("unsupported Composition mode %q", c.Mode)
	}
	if c.Mode == "Resources" {
		resources, err := resourcesFrom(spec, "legacy", "")
		c.Resources = resources
		return c, err
	}
	if len(List(spec["pipeline"])) == 0 {
		return c, fmt.Errorf("pipeline mode requires steps")
	}
	for _, raw := range List(spec["pipeline"]) {
		p := Map(raw)
		if raw, exists := p["input"]; exists && raw != nil {
			if _, ok := raw.(map[string]interface{}); !ok {
				return c, fmt.Errorf("pipeline step input must be an object")
			}
		}
		input := Map(p["input"])
		s := Step{Name: text(p, "step"), Function: text(p, "functionRef", "name"), Input: Map(redacted(input)), Requirements: Map(redacted(p["requirements"])), Adapter: "custom"}
		s.Package = text(funcs[s.Function], "spec", "package")
		switch {
		case strings.HasPrefix(text(input, "apiVersion"), "pt.fn.crossplane.io/") && text(input, "kind") == "Resources":
			s.Adapter = "patch-and-transform"
			s.Source = "Static resource templates and patches"
		case strings.HasPrefix(text(input, "apiVersion"), "gotemplating.fn.crossplane.io/"):
			s.Adapter = "go-templating"
			s.Source = text(input, "source")
			if s.Source == "" {
				s.Source = "Function default source"
			}
		case strings.HasPrefix(text(input, "apiVersion"), "krm.kcl.dev/"):
			s.Adapter = "kcl"
			s.Source = text(input, "spec", "source")
			if strings.Contains(s.Source, "\n") {
				s.Source = "Inline KCL"
			}
		case strings.HasPrefix(text(input, "apiVersion"), "python.fn.crossplane.io/"):
			s.Adapter = "python"
			s.Source = "Inline Python"
		case strings.Contains(strings.ToLower(s.Package), "cue") || strings.Contains(strings.ToLower(text(input, "apiVersion")), "cue"):
			s.Adapter = "cue"
			s.Source = text(input, "source")
		}
		if s.Adapter == "patch-and-transform" || input["resources"] != nil {
			resources, err := resourcesFrom(input, "declared", s.Name)
			if err != nil {
				return c, fmt.Errorf("step %s: %w", s.Name, err)
			}
			c.Resources = append(c.Resources, resources...)
		}
		c.Steps = append(c.Steps, s)
	}
	return c, nil
}

func resourcesFrom(input Object, provenance, step string) ([]Resource, error) {
	var out []Resource
	sets := map[string][]interface{}{}
	for _, p := range List(input["patchSets"]) {
		set := Map(p)
		name := text(set, "name")
		if _, exists := sets[name]; exists {
			return nil, fmt.Errorf("duplicate patch set %q", name)
		}
		sets[name] = List(set["patches"])
	}
	raw, exists := input["resources"]
	if exists && raw != nil {
		if _, ok := raw.([]interface{}); !ok {
			return nil, fmt.Errorf("resources must be a list")
		}
	}
	for _, raw := range List(input["resources"]) {
		m, ok := raw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("resource entry must be an object")
		}
		base := Map(m["base"])
		r := Resource{Name: text(m, "name"), Kind: text(base, "kind"), APIVersion: text(base, "apiVersion"), Manifest: Map(redacted(base)), Provenance: provenance, Step: step, Readiness: redacted(m["readinessChecks"]), Connections: redacted(m["connectionDetails"])}
		patches, err := expandPatches(List(m["patches"]), sets, map[string]bool{})
		if err != nil {
			return nil, err
		}
		for _, p := range patches {
			typ := text(p, "type")
			if typ == "" {
				typ = "FromCompositeFieldPath"
			}
			from := text(p, "fromFieldPath")
			var paths []string
			for _, v := range List(at(p, "combine", "variables")) {
				paths = append(paths, text(Map(v), "fromFieldPath"))
			}
			if len(paths) > 0 {
				from = strings.Join(paths, ", ")
			}
			description := "Direct copy"
			var parts []string
			if p["combine"] != nil {
				parts = append(parts, "Combine "+text(p, "combine", "strategy")+" "+text(p, "combine", "string", "fmt"))
			}
			for _, t := range List(p["transforms"]) {
				tm := Map(t)
				parts = append(parts, "Apply "+text(tm, "type")+": "+compact(tm[text(tm, "type")]))
			}
			if len(parts) > 0 {
				description = strings.Join(parts, " → ")
			}
			if p["policy"] != nil {
				description += "; policy: " + compact(p["policy"])
			}
			r.Patches = append(r.Patches, Patch{Type: typ, From: from, To: text(p, "toFieldPath"), Description: description, Definition: Map(redacted(p))})
		}
		out = append(out, r)
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func expandPatches(patches []interface{}, sets map[string][]interface{}, stack map[string]bool) ([]Object, error) {
	var out []Object
	for _, v := range patches {
		p := Map(v)
		if text(p, "type") != "PatchSet" {
			out = append(out, p)
			continue
		}
		name := text(p, "patchSetName")
		set, ok := sets[name]
		if !ok {
			return nil, fmt.Errorf("missing patch set %q", name)
		}
		if stack[name] {
			return nil, fmt.Errorf("cyclic patch set %q", name)
		}
		stack[name] = true
		expanded, err := expandPatches(set, sets, stack)
		delete(stack, name)
		if err != nil {
			return nil, err
		}
		out = append(out, expanded...)
	}
	return out, nil
}
