package engine

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html"
	htmltemplate "html/template"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

var anchorPattern = regexp.MustCompile(`[^a-z0-9]+`)

func anchor(s string) string {
	return strings.Trim(anchorPattern.ReplaceAllString(strings.ToLower(s), "-"), "-")
}

func Format(doc Document, c Config) ([]byte, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}
	if c.Formatter == "json" {
		b, err := json.MarshalIndent(doc, "", strings.Repeat(" ", c.Settings.Indent))
		return append(b, '\n'), err
	}
	if c.Formatter == "yaml" {
		var buf bytes.Buffer
		enc := yaml.NewEncoder(&buf)
		enc.SetIndent(c.Settings.Indent)
		if err := enc.Encode(doc); err != nil {
			return nil, err
		}
		return buf.Bytes(), nil
	}
	if c.Template != "" {
		data, err := os.ReadFile(c.Template)
		if err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		if c.Formatter == "html" {
			t, err := htmltemplate.New("custom").Parse(string(data))
			if err != nil {
				return nil, err
			}
			err = t.Execute(&buf, doc)
			return buf.Bytes(), err
		}
		t, err := template.New("custom").Funcs(template.FuncMap{"json": compact}).Parse(string(data))
		if err != nil {
			return nil, err
		}
		err = t.Execute(&buf, doc)
		return buf.Bytes(), err
	}
	visible := []string{}
	for _, section := range c.Sections {
		if (section == "apis" && len(doc.APIs) > 0) || (section == "compositions" && len(doc.Compositions) > 0) || (section == "scenarios" && len(doc.Scenarios) > 0) || (section == "examples" && c.Settings.Examples && len(doc.Examples) > 0) {
			visible = append(visible, section)
		}
	}
	c.Sections = visible
	if c.Formatter == "html" {
		return formatHTML(doc, c)
	}
	w := writer{config: c}
	w.heading(1, doc.Title)
	if c.HeaderFrom != "" {
		data, err := os.ReadFile(c.HeaderFrom)
		if err != nil {
			return nil, err
		}
		w.buf.Write(data)
		w.buf.WriteByte('\n')
	}
	writeSections(&w, doc, c)
	if c.FooterFrom != "" {
		data, err := os.ReadFile(c.FooterFrom)
		if err != nil {
			return nil, err
		}
		w.buf.Write(data)
		w.buf.WriteByte('\n')
	}
	return append(bytes.TrimRight(w.buf.Bytes(), "\n"), '\n'), nil
}

func writeSections(w *writer, doc Document, c Config) {
	for _, section := range c.Sections {
		switch section {
		case "apis":
			if len(doc.APIs) == 0 {
				continue
			}
			w.section("apis", "API definitions")
			for _, a := range doc.APIs {
				w.heading(3, a.Kind+" · "+a.Version)
				w.para(a.Description)
				w.table([]string{"Property", "Value"}, [][]string{{"API Group", a.Group}, {"API Version", a.Version}, {"Kind", a.Kind}, {"Scope", a.Scope}, {"Claim Kind", a.ClaimKind}, {"Served", fmt.Sprint(a.Served)}, {"Referenceable", fmt.Sprint(a.Referenceable)}})
				for _, area := range []string{"spec", "status"} {
					var rows [][]string
					for _, f := range a.Fields {
						if f.Path != area && !strings.HasPrefix(f.Path, area+".") {
							continue
						}
						var row []string
						for _, column := range c.Columns {
							switch column {
							case "path":
								row = append(row, f.Path)
							case "type":
								typ := f.Type
								if f.Nullable {
									typ += " (nullable)"
								}
								row = append(row, typ)
							case "description":
								if c.Settings.Descriptions {
									row = append(row, f.Description)
								} else {
									row = append(row, "")
								}
							case "required":
								row = append(row, fmt.Sprint(f.Required))
							case "default":
								value := "-"
								if f.HasDefault {
									value = compact(f.Default)
									if f.Default == nil {
										value = "null"
									} else if str, ok := f.Default.(string); ok {
										value = strconv.Quote(str)
									}
								}
								row = append(row, value)
							case "constraints":
								value := compact(f.Constraints)
								if f.Reference != "" {
									value = "Reference: " + f.Reference
								}
								row = append(row, value)
							}
						}
						rows = append(rows, row)
					}
					if len(rows) > 0 {
						w.heading(4, strings.ToUpper(area[:1])+area[1:]+" Fields")
						if c.Layout == "document" {
							for _, row := range rows {
								w.table(c.Columns, [][]string{row})
							}
						} else {
							w.table(c.Columns, rows)
						}
					}
				}
				if len(a.PrinterColumns) > 0 {
					w.heading(4, "Printer columns")
					w.code("yaml", yamlText(a.PrinterColumns))
				}
			}
		case "compositions":
			if len(doc.Compositions) == 0 {
				continue
			}
			w.section("compositions", "Compositions")
			for _, comp := range doc.Compositions {
				w.heading(3, comp.Name)
				w.para(comp.APIVersion + " / " + comp.Kind + " · " + comp.Mode)
				if len(comp.Steps) > 0 {
					w.heading(4, "Pipeline steps")
					var rows [][]string
					for _, s := range comp.Steps {
						rows = append(rows, []string{s.Name, s.Function, s.Adapter, s.Source, s.Package})
					}
					w.table([]string{"Step", "Function", "Type", "Source", "Package"}, rows)
					w.para("Declared definitions are a source reference. Functions may create additional resources at runtime; see the named render scenarios for actual results.")
				}
				w.heading(4, "Declared resources")
				resourceTable(w, comp.Resources)
				if c.Settings.Patches {
					w.heading(4, "Field Mappings")
					for _, r := range comp.Resources {
						w.heading(5, r.Name)
						rows := [][]string{}
						for _, p := range r.Patches {
							rows = append(rows, []string{p.Type, p.From, p.To, p.Description})
						}
						w.table([]string{"Patch type", "Source field", "Destination field", "Transformation"}, rows)
						if r.Readiness != nil {
							w.para("Readiness checks")
							w.code("yaml", yamlText(r.Readiness))
						}
						if r.Connections != nil {
							w.para("Connection details")
							w.code("yaml", yamlText(r.Connections))
						}
					}
				}
				for _, s := range comp.Steps {
					if len(s.Input) > 0 {
						w.heading(4, s.Name+" source")
						if c.Formatter == "html" {
							writeHTMLSource(w, s)
						} else {
							w.code("yaml", yamlText(s.Input))
						}
					}
					if len(s.Requirements) > 0 {
						w.heading(5, "Requirements")
						w.code("yaml", yamlText(s.Requirements))
					}
				}
				if comp.Environment != nil {
					w.heading(4, "Environment configuration")
					w.code("yaml", yamlText(comp.Environment))
				}
			}
		case "scenarios":
			if len(doc.Scenarios) == 0 {
				continue
			}
			w.section("scenarios", "Rendered scenarios")
			for _, s := range doc.Scenarios {
				w.heading(3, s.Name)
				w.para("Concrete resources for this sample XR and its supplied mock inputs. Other inputs may produce different resources.")
				w.para(fmt.Sprintf("Schema validation: %t", s.Validated))
				if len(s.Inputs) > 0 {
					w.heading(4, "Scenario inputs")
					w.code("yaml", yamlText(s.Inputs))
				}
				resourceTable(w, s.Resources)
				for _, r := range s.Resources {
					w.heading(4, r.Name+" · "+r.Kind)
					w.code("yaml", yamlText(r.Manifest))
				}
				if len(s.Composite) > 0 {
					w.heading(4, "Composite result")
					w.code("yaml", yamlText(s.Composite))
				}
				if len(s.Context) > 0 {
					w.heading(4, "Pipeline context")
					w.code("yaml", yamlText(s.Context))
				}
				if len(s.Results) > 0 {
					w.heading(4, "Function results")
					w.code("yaml", yamlText(s.Results))
				}
			}
		case "examples":
			if !c.Settings.Examples || len(doc.Examples) == 0 {
				continue
			}
			w.section("examples", "Example resources")
			for _, e := range doc.Examples {
				w.code("yaml", yamlText(e))
			}
		}
	}
}

func resourceTable(w *writer, resources []Resource) {
	rows := [][]string{}
	for _, r := range resources {
		rows = append(rows, []string{r.Name, r.Kind, r.APIVersion, r.Provenance})
	}
	w.table([]string{"Resource", "Kind", "API version", "Evidence"}, rows)
}

func yamlText(v interface{}) string {
	b, err := yaml.Marshal(v)
	if err != nil {
		return compact(v)
	}
	return string(b)
}

type writer struct {
	buf          bytes.Buffer
	config       Config
	headingCount int
}

func (w *writer) section(id, title string) {
	if w.config.Formatter == "html" {
		fmt.Fprintf(&w.buf, "<h2 id=\"%s\">%s</h2>\n", id, html.EscapeString(title))
	} else {
		w.heading(2, title)
	}
}

func (w *writer) heading(level int, title string) {
	w.headingCount++
	switch w.config.Formatter {
	case "html":
		id := ""
		if w.config.Settings.Anchors {
			id = fmt.Sprintf(" id=\"heading-%s-%d\"", anchor(title), w.headingCount)
		}
		fmt.Fprintf(&w.buf, "<h%d%s>%s</h%d>\n", level, id, html.EscapeString(title), level)
	case "asciidoc":
		fmt.Fprintf(&w.buf, "%s %s\n\n", strings.Repeat("=", level), title)
	default:
		if w.config.Settings.Anchors {
			fmt.Fprintf(&w.buf, "<a id=\"%s-%d\"></a>\n", anchor(title), w.headingCount)
		}
		fmt.Fprintf(&w.buf, "%s %s\n\n", strings.Repeat("#", level), title)
	}
}

func (w *writer) para(s string) {
	if s == "" {
		return
	}
	if w.config.Formatter == "html" {
		fmt.Fprintf(&w.buf, "<p>%s</p>\n", html.EscapeString(s))
	} else {
		fmt.Fprintf(&w.buf, "%s\n\n", s)
	}
}

func (w *writer) code(language, s string) {
	switch w.config.Formatter {
	case "html":
		fmt.Fprintf(&w.buf, "<div class=\"code-block\"><div class=\"code-bar\"><span>%s</span><button type=\"button\" data-copy hidden>Copy</button></div><pre><code class=\"language-%s\">%s</code></pre><span class=\"copy-status\" role=\"status\"></span></div>\n", html.EscapeString(language), html.EscapeString(language), escapedCode(s))
	case "asciidoc":
		fmt.Fprintf(&w.buf, "[source,%s]\n----\n%s\n----\n\n", language, strings.ReplaceAll(s, "\n----", "\n ----"))
	default:
		fence := "```"
		for strings.Contains(s, fence) {
			fence += "`"
		}
		fmt.Fprintf(&w.buf, "%s%s\n%s\n%s\n\n", fence, language, strings.TrimSuffix(s, "\n"), fence)
	}
}

func (w *writer) table(headers []string, rows [][]string) {
	if len(rows) == 0 {
		w.para("None declared.")
		return
	}
	switch w.config.Formatter {
	case "html":
		w.buf.WriteString("<div class=\"table-wrap\"><table><thead><tr>")
		for _, h := range headers {
			fmt.Fprintf(&w.buf, "<th scope=\"col\">%s</th>", html.EscapeString(h))
		}
		w.buf.WriteString("</tr></thead><tbody>")
		for _, row := range rows {
			w.buf.WriteString("<tr>")
			for i, cell := range row {
				fmt.Fprintf(&w.buf, "<td data-label=\"%s\">%s</td>", html.EscapeString(headers[i]), strings.ReplaceAll(html.EscapeString(cell), "\n", "<br>"))
			}
			w.buf.WriteString("</tr>")
		}
		w.buf.WriteString("</tbody></table></div>\n")
	case "asciidoc":
		w.buf.WriteString("[options=\"header\"]\n|===\n")
		for _, h := range headers {
			fmt.Fprintf(&w.buf, "|%s ", h)
		}
		w.buf.WriteByte('\n')
		for _, row := range rows {
			for _, cell := range row {
				fmt.Fprintf(&w.buf, "|%s ", strings.NewReplacer("|", "\\|", "\n", " + ").Replace(cell))
			}
			w.buf.WriteByte('\n')
		}
		w.buf.WriteString("|===\n\n")
	default:
		cell := func(s string) string {
			return strings.NewReplacer("\\", "\\\\", "|", "\\|", "\r", "", "\n", "<br>").Replace(html.EscapeString(s))
		}
		w.buf.WriteString("| " + strings.Join(headers, " | ") + " |\n| " + strings.TrimSpace(strings.Repeat("--- | ", len(headers))) + "\n")
		for _, row := range rows {
			for i := range row {
				row[i] = cell(row[i])
			}
			w.buf.WriteString("| " + strings.Join(row, " | ") + " |\n")
		}
		w.buf.WriteByte('\n')
	}
}
