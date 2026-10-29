// Package engine builds and formats a versioned documentation model.
package engine

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Object = map[string]interface{}

type Document struct {
	SchemaVersion string        `json:"schemaVersion" yaml:"schemaVersion"`
	Title         string        `json:"title" yaml:"title"`
	APIs          []API         `json:"apis" yaml:"apis"`
	Compositions  []Composition `json:"compositions" yaml:"compositions"`
	Scenarios     []Scenario    `json:"scenarios" yaml:"scenarios"`
	Examples      []Object      `json:"examples,omitempty" yaml:"examples,omitempty"`
}
type API struct {
	Name           string   `json:"name" yaml:"name"`
	Group          string   `json:"group" yaml:"group"`
	Kind           string   `json:"kind" yaml:"kind"`
	ClaimKind      string   `json:"claimKind,omitempty" yaml:"claimKind,omitempty"`
	Scope          string   `json:"scope,omitempty" yaml:"scope,omitempty"`
	Version        string   `json:"version" yaml:"version"`
	Served         bool     `json:"served" yaml:"served"`
	Referenceable  bool     `json:"referenceable" yaml:"referenceable"`
	Description    string   `json:"description,omitempty" yaml:"description,omitempty"`
	Constraints    Object   `json:"constraints,omitempty" yaml:"constraints,omitempty"`
	PrinterColumns []Object `json:"printerColumns,omitempty" yaml:"printerColumns,omitempty"`
	Fields         []Field  `json:"fields" yaml:"fields"`
}
type Field struct {
	Path        string      `json:"path" yaml:"path"`
	Type        string      `json:"type" yaml:"type"`
	Description string      `json:"description,omitempty" yaml:"description,omitempty"`
	Required    bool        `json:"required" yaml:"required"`
	Default     interface{} `json:"default" yaml:"default"`
	HasDefault  bool        `json:"hasDefault" yaml:"hasDefault"`
	Nullable    bool        `json:"nullable,omitempty" yaml:"nullable,omitempty"`
	Constraints Object      `json:"constraints,omitempty" yaml:"constraints,omitempty"`
	Reference   string      `json:"reference,omitempty" yaml:"reference,omitempty"`
}
type Composition struct {
	Name        string      `json:"name" yaml:"name"`
	APIVersion  string      `json:"apiVersion" yaml:"apiVersion"`
	Kind        string      `json:"kind" yaml:"kind"`
	Mode        string      `json:"mode" yaml:"mode"`
	Steps       []Step      `json:"steps" yaml:"steps"`
	Resources   []Resource  `json:"resources" yaml:"resources"`
	Environment interface{} `json:"environment,omitempty" yaml:"environment,omitempty"`
}
type Step struct {
	Name         string `json:"name" yaml:"name"`
	Function     string `json:"function" yaml:"function"`
	Package      string `json:"package,omitempty" yaml:"package,omitempty"`
	Adapter      string `json:"adapter" yaml:"adapter"`
	Source       string `json:"source,omitempty" yaml:"source,omitempty"`
	Input        Object `json:"input,omitempty" yaml:"input,omitempty"`
	Requirements Object `json:"requirements,omitempty" yaml:"requirements,omitempty"`
}
type Resource struct {
	Name        string      `json:"name" yaml:"name"`
	Kind        string      `json:"kind" yaml:"kind"`
	APIVersion  string      `json:"apiVersion" yaml:"apiVersion"`
	Provenance  string      `json:"provenance" yaml:"provenance"`
	Step        string      `json:"step,omitempty" yaml:"step,omitempty"`
	Patches     []Patch     `json:"patches,omitempty" yaml:"patches,omitempty"`
	Readiness   interface{} `json:"readiness,omitempty" yaml:"readiness,omitempty"`
	Connections interface{} `json:"connections,omitempty" yaml:"connections,omitempty"`
	Manifest    Object      `json:"manifest,omitempty" yaml:"manifest,omitempty"`
}
type Patch struct {
	Type        string `json:"type" yaml:"type"`
	From        string `json:"from" yaml:"from"`
	To          string `json:"to" yaml:"to"`
	Description string `json:"description" yaml:"description"`
	Definition  Object `json:"definition" yaml:"definition"`
}
type Scenario struct {
	Inputs    Object     `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Name      string     `json:"name" yaml:"name"`
	XR        string     `json:"xr,omitempty" yaml:"xr,omitempty"`
	Validated bool       `json:"validated" yaml:"validated"`
	Composite Object     `json:"composite,omitempty" yaml:"composite,omitempty"`
	Resources []Resource `json:"resources" yaml:"resources"`
	Context   []Object   `json:"context,omitempty" yaml:"context,omitempty"`
	Results   []Object   `json:"results,omitempty" yaml:"results,omitempty"`
}

func Map(v interface{}) Object {
	if m, ok := v.(map[string]interface{}); ok {
		return m
	}
	return Object{}
}

func List(v interface{}) []interface{} {
	if a, ok := v.([]interface{}); ok {
		return a
	}
	return nil
}

func String(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprint(v)
}

func at(m Object, keys ...string) interface{} {
	var v interface{} = m
	for _, k := range keys {
		v = Map(v)[k]
	}
	return v
}
func text(m Object, keys ...string) string { return String(at(m, keys...)) }
func boolean(v interface{}) bool           { b, _ := v.(bool); return b }
func compact(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	b, err := json.Marshal(v)
	if err != nil {
		return fmt.Sprint(v)
	}
	return string(b)
}

func has(v interface{}, s string) bool {
	for _, x := range List(v) {
		if String(x) == s {
			return true
		}
	}
	return false
}

func redacted(v interface{}) interface{} {
	switch x := v.(type) {
	case map[string]interface{}:
		out := Object{}
		for k, val := range x {
			lower := strings.ToLower(k)
			if (text(x, "kind") == "Secret" && (k == "data" || k == "stringData")) || strings.Contains(lower, "password") || strings.Contains(lower, "credential") || strings.HasSuffix(lower, "token") || lower == "apikey" || lower == "api-key" || strings.HasSuffix(lower, "accesskey") || strings.HasSuffix(lower, "accesskeyid") || lower == "clientsecret" {
				out[k] = "[redacted]"
			} else {
				out[k] = redacted(val)
			}
		}
		return out
	case []interface{}:
		out := make([]interface{}, len(x))
		for i, v := range x {
			out[i] = redacted(v)
		}
		return out
	default:
		return v
	}
}
