package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func extractAPI(in Input, opts Options) ([]API, error) {
	m := in.Object
	spec := Map(m["spec"])
	group, kind := text(spec, "group"), text(spec, "names", "kind")
	if group == "" || kind == "" {
		return nil, fmt.Errorf("%s: XRD requires spec.group and spec.names.kind", in.File)
	}
	versions := List(spec["versions"])
	if len(versions) == 0 {
		return nil, fmt.Errorf("no versions found in XRD")
	}
	selected := -1
	if !opts.AllVersions {
		for i, v := range versions {
			vm := Map(v)
			if opts.APIVersion != "" {
				if text(vm, "name") == opts.APIVersion {
					selected = i
					break
				}
			} else if boolean(vm["served"]) {
				selected = i
				break
			}
		}
		if selected < 0 {
			if opts.APIVersion != "" {
				return nil, fmt.Errorf("XRD %s has no version %q", kind, opts.APIVersion)
			}
			selected = 0
		}
	}
	var apis []API
	for i, v := range versions {
		if !opts.AllVersions && i != selected {
			continue
		}
		version := Map(v)
		if text(version, "name") == "" {
			return nil, fmt.Errorf("XRD version requires name")
		}
		schema := Map(at(version, "schema", "openAPIV3Schema"))
		api := API{Name: text(m, "metadata", "name"), Group: group, Kind: kind, ClaimKind: text(spec, "claimNames", "kind"), Scope: text(spec, "scope"), Version: text(version, "name"), Served: boolean(version["served"]), Referenceable: boolean(version["referenceable"]), Description: text(schema, "description"), Fields: []Field{}}
		for _, c := range List(version["additionalPrinterColumns"]) {
			api.PrinterColumns = append(api.PrinterColumns, Map(c))
		}
		walker := schemaWalker{root: m, file: in.File, opts: opts, stack: map[string]bool{}}
		for _, section := range []string{"spec", "status"} {
			if prop, ok := Map(schema["properties"])[section]; ok {
				if err := walker.walk(Map(prop), section, has(schema["required"], section), 0, &api.Fields); err != nil {
					return nil, err
				}
			}
		}
		apis = append(apis, api)
	}
	return apis, nil
}

type schemaWalker struct {
	root  Object
	file  string
	opts  Options
	stack map[string]bool
}

func (w schemaWalker) resolve(ref string) (Object, Object, string, error) {
	parts := strings.SplitN(ref, "#", 2)
	file := w.file
	root := w.root
	if parts[0] != "" {
		if strings.Contains(parts[0], "://") {
			return nil, nil, "", fmt.Errorf("remote schema reference requires an explicitly supplied local schema: %s", ref)
		}
		file = filepath.Clean(filepath.Join(filepath.Dir(w.file), parts[0]))
		allowed := false
		for _, p := range w.opts.RefFiles {
			abs, _ := filepath.Abs(p)
			if abs == file {
				allowed = true
			}
		}
		if !allowed {
			return nil, nil, "", fmt.Errorf("schema reference %q requires --schema %s", ref, file)
		}
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, nil, "", err
		}
		inputs, err := Read(data, file)
		if err != nil {
			return nil, nil, "", err
		}
		if len(inputs) != 1 {
			return nil, nil, "", fmt.Errorf("schema reference file requires one document")
		}
		root = inputs[0].Object
	}
	var current interface{} = root
	if len(parts) == 2 && parts[1] != "" {
		if !strings.HasPrefix(parts[1], "/") {
			return nil, nil, "", fmt.Errorf("invalid JSON pointer %q", ref)
		}
		for _, token := range strings.Split(strings.TrimPrefix(parts[1], "/"), "/") {
			token = strings.ReplaceAll(strings.ReplaceAll(token, "~1", "/"), "~0", "~")
			if arr, ok := current.([]interface{}); ok {
				index, err := strconv.Atoi(token)
				if err != nil || index < 0 || index >= len(arr) {
					return nil, nil, "", fmt.Errorf("missing schema reference %s", ref)
				}
				current = arr[index]
			} else {
				v, ok := Map(current)[token]
				if !ok {
					return nil, nil, "", fmt.Errorf("missing schema reference %s", ref)
				}
				current = v
			}
		}
	}
	resolved, ok := current.(map[string]interface{})
	if !ok {
		return nil, nil, "", fmt.Errorf("schema reference is not an object: %s", ref)
	}
	return resolved, root, file, nil
}

func (w schemaWalker) walk(schema Object, path string, required bool, depth int, out *[]Field) error {
	if depth > 128 {
		return fmt.Errorf("schema nesting exceeds 128 levels at %s", path)
	}
	ref := text(schema, "$ref")
	if ref != "" {
		resolved, root, file, err := w.resolve(ref)
		if err != nil {
			return err
		}
		fragment := ""
		if parts := strings.SplitN(ref, "#", 2); len(parts) == 2 {
			fragment = parts[1]
		}
		key := file + "#" + fragment
		if w.stack[key] {
			*out = append(*out, Field{Path: path, Type: "reference", Reference: ref, Required: required})
			return nil
		}
		stack := map[string]bool{}
		for k, v := range w.stack {
			stack[k] = v
		}
		stack[key] = true
		merged := Object{}
		for k, v := range resolved {
			merged[k] = v
		}
		for k, v := range schema {
			if k != "$ref" {
				merged[k] = v
			}
		}
		return (schemaWalker{root: root, file: file, opts: w.opts, stack: stack}).walk(merged, path, required, depth+1, out)
	}
	typ := text(schema, "type")
	if typ == "" && schema["properties"] != nil {
		typ = "object"
	}
	if typ == "array" {
		itemType := text(Map(schema["items"]), "type")
		if itemType == "" {
			itemType = "any"
		}
		typ = "list(" + itemType + ")"
	}
	if typ == "" {
		typ = "any"
	}
	f := Field{Path: path, Type: typ, Description: text(schema, "description"), Required: required, Nullable: boolean(schema["nullable"]), Constraints: Object{}}
	f.Default, f.HasDefault = schema["default"]
	for _, key := range []string{"enum", "minimum", "maximum", "exclusiveMinimum", "exclusiveMaximum", "multipleOf", "minLength", "maxLength", "pattern", "format", "minItems", "maxItems", "uniqueItems", "minProperties", "maxProperties", "x-kubernetes-validations", "x-kubernetes-list-type", "x-kubernetes-list-map-keys", "x-kubernetes-map-type", "x-kubernetes-int-or-string", "x-kubernetes-preserve-unknown-fields", "oneOf", "anyOf", "allOf", "not"} {
		if v, ok := schema[key]; ok {
			f.Constraints[key] = v
		}
	}
	if (path != "spec" && path != "status") || len(f.Constraints) > 0 {
		*out = append(*out, f)
	}
	if !w.opts.Config.Settings.Nested && depth > 0 {
		return nil
	}
	props := Map(schema["properties"])
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if w.opts.Config.Sort.RequiredFirst {
			a, b := has(schema["required"], keys[i]), has(schema["required"], keys[j])
			if a != b {
				return a
			}
		}
		if w.opts.Config.Sort.By == "type" {
			a, b := text(Map(props[keys[i]]), "type"), text(Map(props[keys[j]]), "type")
			if a != b {
				return a < b
			}
		}
		return keys[i] < keys[j]
	})
	for _, k := range keys {
		if err := w.walk(Map(props[k]), path+"."+k, has(schema["required"], k), depth+1, out); err != nil {
			return err
		}
	}
	if items, ok := schema["items"].(map[string]interface{}); ok && w.opts.Config.Settings.Nested {
		if err := w.walk(items, path+"[]", false, depth+1, out); err != nil {
			return err
		}
	}
	if boolean(schema["additionalProperties"]) && w.opts.Config.Settings.Nested {
		if err := w.walk(Object{}, path+".*", false, depth+1, out); err != nil {
			return err
		}
	}
	if additional, ok := schema["additionalProperties"].(map[string]interface{}); ok && w.opts.Config.Settings.Nested {
		if err := w.walk(additional, path+".*", false, depth+1, out); err != nil {
			return err
		}
	}
	for _, alternative := range []string{"oneOf", "anyOf", "allOf"} {
		for i, v := range List(schema[alternative]) {
			if err := w.walk(Map(v), fmt.Sprintf("%s.%s[%d]", path, alternative, i), false, depth+1, out); err != nil {
				return err
			}
		}
	}
	if v, ok := schema["not"].(map[string]interface{}); ok {
		if err := w.walk(v, path+".not", false, depth+1, out); err != nil {
			return err
		}
	}
	return nil
}
