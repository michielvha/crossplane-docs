package engine

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type DriftError struct{ File string }

func (e *DriftError) Error() string { return "documentation is out of date: " + e.File }
func Write(data []byte, c Config, check bool) error {
	file := c.Output.File
	if file == "" {
		return fmt.Errorf("file output is required for --check or injection")
	}
	var old []byte
	mode := os.FileMode(0o644)
	st, err := os.Lstat(file)
	if err == nil {
		if !st.Mode().IsRegular() {
			return fmt.Errorf("output must be a regular file: %s", file)
		}
		mode = st.Mode().Perm()
		old, err = os.ReadFile(file)
		if err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}
	if c.Output.Mode == "inject" {
		if old == nil {
			return fmt.Errorf("injection target does not exist: %s", file)
		}
		begin, end := c.Output.Begin, c.Output.End
		if begin == "" {
			begin = "<!-- BEGIN_CROSSPLANE_DOCS -->"
			end = "<!-- END_CROSSPLANE_DOCS -->"
			if c.Formatter == "asciidoc" {
				begin = "// BEGIN_CROSSPLANE_DOCS"
				end = "// END_CROSSPLANE_DOCS"
			}
		}
		if begin == end || bytes.Count(old, []byte(begin)) != 1 || bytes.Count(old, []byte(end)) != 1 {
			return fmt.Errorf("expected exactly one distinct begin and end marker in %s", file)
		}
		i, j := bytes.Index(old, []byte(begin))+len(begin), bytes.Index(old, []byte(end))
		if i > j {
			return fmt.Errorf("injection markers are reversed")
		}
		next := append([]byte{}, old[:i]...)
		next = append(next, []byte("\n"+strings.TrimSuffix(string(data), "\n")+"\n")...)
		next = append(next, old[j:]...)
		data = next
	}
	if bytes.Equal(old, data) {
		return nil
	}
	if check {
		return &DriftError{File: file}
	}
	tmp, err := os.CreateTemp(filepath.Dir(file), ".crossplane-docs-*")
	if err != nil {
		return err
	}
	name := tmp.Name()
	defer func() { _ = os.Remove(name) }()
	if err := tmp.Chmod(mode); err != nil {
		_ = tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(name, file); err != nil {
		return err
	}
	return nil
}
func IsDrift(err error) bool { var drift *DriftError; return errors.As(err, &drift) }
