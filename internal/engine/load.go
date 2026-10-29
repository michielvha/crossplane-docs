package engine

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type (
	Input struct {
		Object Object
		File   string
		Index  int
	}
	Options struct {
		Kind        string
		APIVersion  string
		AllVersions bool
		Config      Config
		RefFiles    []string
	}
)

func Read(data []byte, file string) ([]Input, error) {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	var out []Input
	for i := 1; ; i++ {
		var m Object
		err := dec.Decode(&m)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("%s document %d: %w", file, i, err)
		}
		if len(m) > 0 {
			out = append(out, Input{Object: m, File: file, Index: i})
		}
	}
	return out, nil
}

func Load(path string, c Config) ([]Input, error) {
	st, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	files := []string{path}
	if st.IsDir() {
		files = nil
		err = filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if p != path && (strings.HasPrefix(d.Name(), ".") || !c.Recursive) {
					return filepath.SkipDir
				}
				return nil
			}
			if d.Type()&os.ModeSymlink != 0 {
				return nil
			}
			if d.Name() == ".crossplane-docs.yml" {
				return nil
			}
			if filepath.Ext(p) != ".yaml" && filepath.Ext(p) != ".yml" {
				return nil
			}
			rel, _ := filepath.Rel(path, p)
			match := func(patterns []string) bool {
				for _, pattern := range patterns {
					if yes, _ := filepath.Match(pattern, filepath.ToSlash(rel)); yes {
						return true
					}
					if yes, _ := filepath.Match(pattern, d.Name()); yes {
						return true
					}
				}
				return false
			}
			if (len(c.Include) == 0 || match(c.Include)) && !match(c.Exclude) {
				files = append(files, p)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	}
	sort.Strings(files)
	var inputs []Input
	seen := map[string]bool{}
	for _, f := range files {
		abs, err := filepath.Abs(f)
		if err != nil {
			return nil, err
		}
		if seen[abs] {
			continue
		}
		seen[abs] = true
		data, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		docs, err := Read(data, abs)
		if err != nil {
			return nil, err
		}
		inputs = append(inputs, docs...)
	}
	return inputs, nil
}

func Build(inputs []Input, opts Options) (Document, error) {
	// Structured output exposes the complete model, independent of document views.
	if opts.Config.Formatter == "json" || opts.Config.Formatter == "yaml" {
		opts.Config.Settings.Nested = true
	}
	doc := Document{SchemaVersion: "1", Title: opts.Config.Title, APIs: []API{}, Compositions: []Composition{}, Scenarios: []Scenario{}}
	if len(inputs) == 0 {
		return doc, fmt.Errorf("no YAML documents found")
	}
	if opts.Kind != "" && (len(inputs) != 1 || text(inputs[0].Object, "kind") != opts.Kind) {
		return doc, fmt.Errorf("expected one %s document", opts.Kind)
	}
	funcs := map[string]Object{}
	for _, in := range inputs {
		if text(in.Object, "kind") == "Function" {
			funcs[text(in.Object, "metadata", "name")] = in.Object
		}
	}
	seen := map[string]bool{}
	for _, in := range inputs {
		m := in.Object
		kind := text(m, "kind")
		if kind != "CompositeResourceDefinition" && kind != "Composition" {
			continue
		}
		key := text(m, "apiVersion") + "/" + kind + "/" + text(m, "metadata", "name")
		if seen[key] {
			return doc, fmt.Errorf("duplicate resource identity %s (%s document %d)", key, in.File, in.Index)
		}
		seen[key] = true
		if !strings.HasPrefix(text(m, "apiVersion"), "apiextensions.crossplane.io/") {
			return doc, fmt.Errorf("%s: expected apiextensions.crossplane.io API", in.File)
		}
		switch kind {
		case "CompositeResourceDefinition":
			apis, err := extractAPI(in, opts)
			if err != nil {
				return doc, err
			}
			doc.APIs = append(doc.APIs, apis...)
		case "Composition":
			comp, err := extractComposition(m, funcs)
			if err != nil {
				return doc, err
			}
			doc.Compositions = append(doc.Compositions, comp)
		}
	}
	if len(doc.APIs)+len(doc.Compositions) == 0 {
		return doc, fmt.Errorf("no XRDs or Compositions found")
	}
	for _, in := range inputs {
		for _, api := range doc.APIs {
			if text(in.Object, "apiVersion") == api.Group+"/"+api.Version && (text(in.Object, "kind") == api.Kind || text(in.Object, "kind") == api.ClaimKind) {
				doc.Examples = append(doc.Examples, Map(redacted(in.Object)))
				break
			}
		}
	}
	sort.SliceStable(doc.APIs, func(i, j int) bool {
		a, b := doc.APIs[i], doc.APIs[j]
		return a.Group+"/"+a.Kind+"/"+a.Version < b.Group+"/"+b.Kind+"/"+b.Version
	})
	sort.SliceStable(doc.Compositions, func(i, j int) bool { return doc.Compositions[i].Name < doc.Compositions[j].Name })
	return doc, nil
}
