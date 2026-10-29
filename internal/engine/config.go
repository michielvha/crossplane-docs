package engine

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Version    int      `yaml:"version"`
	Formatter  string   `yaml:"formatter"`
	Title      string   `yaml:"title"`
	Layout     string   `yaml:"layout"`
	HTMLLayout string   `yaml:"html-layout"`
	Sections   []string `yaml:"sections"`
	Columns    []string `yaml:"columns"`
	Template   string   `yaml:"template"`
	HeaderFrom string   `yaml:"header-from"`
	FooterFrom string   `yaml:"footer-from"`
	Settings   Settings `yaml:"settings"`
	Sort       Sort     `yaml:"sort"`
	Output     Output   `yaml:"output"`
	Recursive  bool     `yaml:"recursive"`
	Include    []string `yaml:"include"`
	Exclude    []string `yaml:"exclude"`
}
type Settings struct {
	Nested       bool   `yaml:"nested"`
	Descriptions bool   `yaml:"descriptions"`
	Anchors      bool   `yaml:"anchors"`
	Examples     bool   `yaml:"examples"`
	Patches      bool   `yaml:"patches"`
	Indent       int    `yaml:"indent"`
	Style        string `yaml:"style"`
}
type (
	Sort struct {
		By            string `yaml:"by"`
		RequiredFirst bool   `yaml:"required-first"`
	}
	Output struct {
		File  string `yaml:"file"`
		Mode  string `yaml:"mode"`
		Begin string `yaml:"begin"`
		End   string `yaml:"end"`
	}
)

func Defaults() Config {
	return Config{Version: 1, Formatter: "markdown", Title: "Crossplane API reference", Layout: "table", HTMLLayout: "reference", Sections: []string{"apis", "compositions", "scenarios", "examples"}, Columns: []string{"path", "type", "description", "required", "default", "constraints"}, Settings: Settings{Nested: true, Descriptions: true, Anchors: true, Examples: true, Patches: true, Indent: 2}, Sort: Sort{By: "name", RequiredFirst: true}, Output: Output{Mode: "replace"}, Recursive: true}
}

func LoadConfig(input, explicit string) (Config, error) {
	c := Defaults()
	root := input
	if st, err := os.Stat(root); err == nil && !st.IsDir() {
		root = filepath.Dir(root)
	}
	file := explicit
	if file == "" {
		for _, candidate := range []string{filepath.Join(root, ".crossplane-docs.yml"), ".crossplane-docs.yml"} {
			if _, err := os.Stat(candidate); err == nil {
				file = candidate
				break
			} else if !os.IsNotExist(err) {
				return c, err
			}
		}
	}
	if file != "" {
		data, err := os.ReadFile(file)
		if err != nil {
			return c, err
		}
		dec := yaml.NewDecoder(bytes.NewReader(data))
		dec.KnownFields(true)
		if err := dec.Decode(&c); err != nil {
			return c, fmt.Errorf("configuration %s: %w", file, err)
		}
		var extra interface{}
		if err := dec.Decode(&extra); err != io.EOF {
			return c, fmt.Errorf("configuration must contain exactly one YAML document")
		}
		base := filepath.Dir(file)
		for _, p := range []*string{&c.Template, &c.HeaderFrom, &c.FooterFrom, &c.Output.File} {
			if *p != "" && !filepath.IsAbs(*p) {
				*p = filepath.Join(base, *p)
			}
		}
	}
	return c, nil
}

func (c Config) Validate() error {
	if c.Version != 1 {
		return fmt.Errorf("unsupported configuration version %d", c.Version)
	}
	if !slices.Contains([]string{"markdown", "html", "asciidoc", "json", "yaml"}, c.Formatter) {
		return fmt.Errorf("unknown formatter %q", c.Formatter)
	}
	if c.Template != "" && (c.Formatter == "json" || c.Formatter == "yaml") {
		return fmt.Errorf("custom templates require a document formatter")
	}
	if !slices.Contains([]string{"reference", "explorer", "compact"}, c.HTMLLayout) {
		return fmt.Errorf("unknown html-layout %q", c.HTMLLayout)
	}
	if !slices.Contains([]string{"table", "document"}, c.Layout) {
		return fmt.Errorf("unknown layout %q", c.Layout)
	}
	if !slices.Contains([]string{"name", "path", "type"}, c.Sort.By) {
		return fmt.Errorf("unknown sort field %q", c.Sort.By)
	}
	if !slices.Contains([]string{"replace", "inject"}, c.Output.Mode) {
		return fmt.Errorf("unknown output mode %q", c.Output.Mode)
	}
	if c.Output.Mode == "inject" && (c.Formatter == "json" || c.Formatter == "yaml" || c.Output.File == "") {
		return fmt.Errorf("injection requires a document formatter and an output file")
	}
	for _, s := range c.Sections {
		if !slices.Contains([]string{"apis", "compositions", "scenarios", "examples"}, s) {
			return fmt.Errorf("unknown section %q", s)
		}
	}
	for _, s := range c.Columns {
		if !slices.Contains([]string{"path", "type", "description", "required", "default", "constraints"}, s) {
			return fmt.Errorf("unknown field column %q", s)
		}
	}
	if len(c.Columns) == 0 {
		return fmt.Errorf("at least one field column is required")
	}
	if c.Settings.Indent < 0 || c.Settings.Indent > 8 {
		return fmt.Errorf("indent must be between 0 and 8")
	}
	if (c.Output.Begin == "") != (c.Output.End == "") {
		return fmt.Errorf("both injection markers must be supplied")
	}
	if strings.ContainsAny(c.Title, "\r\n") {
		return fmt.Errorf("title must be a single line")
	}
	return nil
}
