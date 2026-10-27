// Package document provides input and Markdown handling shared by generators.
package document

import (
	"bytes"
	"fmt"
	"html"
	"io"
	"strings"

	"gopkg.in/yaml.v3"
)

// Decode reads exactly one YAML document so bundles cannot be silently truncated.
func Decode(data []byte, target interface{}) error {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	if err := decoder.Decode(target); err != nil {
		return err
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return err
		}
		return fmt.Errorf("expected one YAML document; split the input into separate files")
	}
	return nil
}

// Cell keeps descriptions and values inside a single Markdown table cell.
func Cell(value string) string {
	value = html.EscapeString(strings.TrimSpace(value))
	return strings.NewReplacer("&#34;", "\"", "&#39;", "'", "|", "\\|", "\r\n", "<br>", "\n", "<br>", "\r", "<br>").Replace(value)
}
